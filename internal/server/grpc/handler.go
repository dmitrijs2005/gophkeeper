package grpc

import (
	"context"
	"errors"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
	pb "github.com/dmitrijs2005/gophkeeper/internal/proto"
	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
	"github.com/dmitrijs2005/gophkeeper/internal/server/shared"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{Status: "OK"}, nil
}

func (s *GRPCServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {

	tokenPair, err := s.users.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		s.logger.Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.logger.Info(ctx, "Refresh token generated")
	return &pb.RefreshTokenResponse{AccessToken: tokenPair.AccessToken, RefreshToken: tokenPair.RefreshToken}, nil

}

func (s *GRPCServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {

	result, err := s.users.Register(ctx, req.Username, req.Salt, req.Verifier)

	if err != nil {
		s.logger.Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.logger.Info(ctx, "Registered", "username", req.Username)
	return &pb.RegisterUserResponse{Username: "registered id=" + result.ID}, nil

}

func (s *GRPCServer) GetSalt(ctx context.Context, req *pb.GetSaltRequest) (*pb.GetSaltResponse, error) {

	result, err := s.users.GetSalt(ctx, req.Username)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.logger.Info(ctx, "Salt returned", "username", req.Username)

	return &pb.GetSaltResponse{Salt: result}, nil

}

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

func (s *GRPCServer) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {

	var pendingEntries []*models.Entry

	v := ctx.Value(shared.UserIDKey)
	userID, ok := v.(string)
	if !ok {
		return nil, status.Error(codes.Internal, "internal error")
	}

	for _, e := range req.Entries {
		pe := &models.Entry{
			UserID:        userID,
			ID:            e.Id,
			Deleted:       e.Deleted,
			Overview:      e.Overview,
			NonceOverview: e.NonceOverview,
			Details:       e.Details,
			NonceDetails:  e.NonceDetails,
		}
		pendingEntries = append(pendingEntries, pe)
	}

	err := s.entries.Sync(ctx, pendingEntries)
	if err != nil {
		s.logger.Error(ctx, err.Error())
		if errors.Is(err, common.ErrorUnauthorized) {
			return nil, status.Error(codes.Unauthenticated, "unauthorized")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	var processedEntries []*pb.Entry

	return &pb.SyncResponse{ProcessedEntries: processedEntries}, nil

}
