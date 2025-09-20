package grpc

import (
	"context"
	"fmt"

	pb "github.com/dmitrijs2005/gophkeeper/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {

	s.logger.Info(ctx, "Registration request")

	result, err := s.users.Register(ctx, req.Username, req.Salt, req.Verifier)

	if err != nil {
		s.logger.Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.logger.Info(ctx, "Registered", "username", req.Username)
	return &pb.RegisterUserResponse{Username: "registered id=" + result.ID}, nil

}

func (s *GRPCServer) GetSalt(ctx context.Context, req *pb.GetSaltRequest) (*pb.GetSaltResponse, error) {

	s.logger.Info(ctx, "Get salt request")

	result, err := s.users.GetSalt(ctx, req.Username)

	if err != nil {
		s.logger.Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetSaltResponse{Salt: result}, nil

}

func (s *GRPCServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {

	s.logger.Info(ctx, "Login request")

	accessToken, refreshToken, err := s.users.Login(ctx, req.Username, req.VerifierCandidate)

	if err != nil {
		s.logger.Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil

}

func (s *GRPCServer) AddEntry(ctx context.Context, req *pb.AddEntryRequest) (*pb.AddEntryResponse, error) {

	userID, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return nil, fmt.Errorf("no user id in context")
	}

	s.logger.Info(ctx, "Add entry request")

	_, err := s.entries.Create(ctx, userID, req.Title, req.Type, req.Cyphertext, req.Nonce)

	if err != nil {
		s.logger.Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AddEntryResponse{Result: "ok"}, nil

}

func (s *GRPCServer) GetPresignedPutUrl(ctx context.Context, req *pb.GetPresignedPutUrlRequest) (*pb.GetPresignedPutUrlResponse, error) {
	key, url, err := s.entries.GetPresignedPutUrl(ctx)
	if err != nil {
		s.logger.Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetPresignedPutUrlResponse{Key: key, Url: url}, nil
}

func (s *GRPCServer) GetPresignedGetUrl(ctx context.Context, req *pb.GetPresignedGetUrlRequest) (*pb.GetPresignedGetUrlResponse, error) {
	result, err := s.entries.GetPresignedGetUrl(ctx, req.Key)
	if err != nil {
		s.logger.Error(ctx, err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetPresignedGetUrlResponse{Url: result}, nil
}
