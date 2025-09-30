package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
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

	result, err := s.users.GetSalt(ctx, req.Username)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

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

	return &pb.LoginResponse{AccessToken: tokens.AccessToken, RefreshToken: tokens.RefreshToken}, nil

}

func (s *GRPCServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {

	return &pb.PingResponse{Status: "OK"}, nil

}

func (s *GRPCServer) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {

	for a, b := range req.Entries {
		fmt.Println(a, b)
	}

	var ProcessedEntries []*pb.Entry

	//x := s.entries.Sync(ctx, )

	return &pb.SyncResponse{ProcessedEntries: ProcessedEntries}, nil

}

//func (s *GRPCServer) reqToEntries()
