// Package grpc exposes the gRPC transport for the server, wiring protobuf
// requests to domain services, mapping domain errors to gRPC status codes,
// and enriching logs with structured context.
package grpc

import (
	"context"
	"errors"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
	pb "github.com/dmitrijs2005/gophkeeper/internal/proto"
	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Ping is a simple liveness probe that always returns "OK".
func (s *GRPCServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{Status: "OK"}, nil
}

// RefreshToken exchanges a valid refresh token for a new (access, refresh) pair.
// Returns codes.Internal on service errors.
func (s *GRPCServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	tokenPair, err := s.users.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		s.logger.Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.logger.Info(ctx, "Refresh token generated")
	return &pb.RefreshTokenResponse{AccessToken: tokenPair.AccessToken, RefreshToken: tokenPair.RefreshToken}, nil
}

// RegisterUser creates a new user with the provided username, salt, and verifier.
// Returns codes.Internal on service errors.
func (s *GRPCServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	result, err := s.users.Register(ctx, req.Username, req.Salt, req.Verifier)
	if err != nil {
		s.logger.Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.logger.Info(ctx, "Registered", "username", req.Username)
	return &pb.RegisterUserResponse{Username: "registered id=" + result.ID}, nil
}

// GetSalt fetches the server-stored salt for the given username.
// Returns codes.Internal on service errors.
func (s *GRPCServer) GetSalt(ctx context.Context, req *pb.GetSaltRequest) (*pb.GetSaltResponse, error) {
	result, err := s.users.GetSalt(ctx, req.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.logger.Info(ctx, "Salt returned", "username", req.Username)
	return &pb.GetSaltResponse{Salt: result}, nil
}

// Login validates the verifier candidate and returns new access/refresh tokens.
// Returns codes.Unauthenticated for invalid credentials, codes.Internal otherwise.
func (s *GRPCServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	tokens, err := s.users.Login(ctx, req.Username, req.VerifierCandidate)
	if err != nil {
		if errors.Is(err, common.ErrorUnauthorized) {
			return nil, status.Error(codes.Unauthenticated, "unauthorized")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	s.logger.Info(ctx, "Logged in", "username", req.Username)
	return &pb.LoginResponse{AccessToken: tokens.AccessToken, RefreshToken: tokens.RefreshToken}, nil
}

// Sync reconciles client-submitted pending entries/files with the server state,
// returns merged updates, server-side new items, upload tasks, and the new
// global max version. Authentication is inferred from context.
func (s *GRPCServer) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	v := ctx.Value(UserIDKey)
	userID, ok := v.(string)
	if !ok {
		return nil, status.Error(codes.Internal, "internal error")
	}

	var pendingEntries []*models.Entry
	for _, e := range req.Entries {
		pendingEntries = append(pendingEntries, &models.Entry{
			UserID:        userID,
			ID:            e.Id,
			Deleted:       e.Deleted,
			Overview:      e.Overview,
			NonceOverview: e.NonceOverview,
			Details:       e.Details,
			NonceDetails:  e.NonceDetails,
		})
	}

	var pendingFiles []*models.File
	for _, f := range req.Files {
		pendingFiles = append(pendingFiles, &models.File{
			UserID:           userID,
			EntryID:          f.EntryId,
			EncryptedFileKey: f.FileKey,
			Nonce:            f.Nonce,
		})
	}

	processedEntries, newEntries, newFiles, uploadTasks, maxVersion, err := s.entries.Sync(ctx, userID, pendingEntries, pendingFiles, req.MaxVersion)
	if err != nil {
		s.logger.Error(ctx, err.Error())
		if errors.Is(err, common.ErrorUnauthorized) {
			return nil, status.Error(codes.Unauthenticated, "unauthorized")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	var pe, ne []*pb.Entry
	for _, e := range processedEntries {
		pe = append(pe, &pb.Entry{
			Id:            e.ID,
			Version:       e.Version,
			Overview:      e.Overview,
			NonceOverview: e.NonceOverview,
			Details:       e.Details,
			NonceDetails:  e.NonceDetails,
			Deleted:       e.Deleted,
		})
	}
	for _, e := range newEntries {
		ne = append(ne, &pb.Entry{
			Id:            e.ID,
			Version:       e.Version,
			Overview:      e.Overview,
			NonceOverview: e.NonceOverview,
			Details:       e.Details,
			NonceDetails:  e.NonceDetails,
			Deleted:       e.Deleted,
		})
	}

	var nf []*pb.File
	for _, f := range newFiles {
		nf = append(nf, &pb.File{
			EntryId: f.EntryID,
			FileKey: f.EncryptedFileKey,
			Nonce:   f.Nonce,
		})
	}

	var ut []*pb.UploadTask
	for _, t := range uploadTasks {
		ut = append(ut, &pb.UploadTask{
			EntryId: t.EntryID,
			Url:     t.URL,
		})
	}

	return &pb.SyncResponse{
		ProcessedEntries: pe,
		NewEntries:       ne,
		NewFiles:         nf,
		UploadTasks:      ut,
		GlobalMaxVersion: maxVersion,
	}, nil
}

// MarkUploaded acknowledges that the client finished uploading a file
// for the given entry. Returns codes.Internal on errors.
func (s *GRPCServer) MarkUploaded(ctx context.Context, req *pb.MarkUploadedRequest) (*pb.MarkUploadedResponse, error) {
	entryID := req.EntryId
	if err := s.entries.MarkUploaded(ctx, entryID); err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &pb.MarkUploadedResponse{}, nil
}

// GetPresignedGetUrl returns a presigned GET URL for downloading the encrypted
// file associated with the given entry. Returns codes.Internal on errors.
func (s *GRPCServer) GetPresignedGetUrl(ctx context.Context, req *pb.GetPresignedGetUrlRequest) (*pb.GetPresignedGetUrlResponse, error) {
	entryID := req.EntryId
	url, err := s.entries.GetPresignedGetURL(ctx, entryID)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &pb.GetPresignedGetUrlResponse{Url: url}, nil
}
