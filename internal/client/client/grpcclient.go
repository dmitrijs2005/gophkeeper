package client

import (
	"context"
	"fmt"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
	pb "github.com/dmitrijs2005/gophkeeper/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type GRPCClient struct {
	endpointURL  string
	conn         *grpc.ClientConn
	client       pb.GophKeeperServiceClient
	accessToken  string
	refreshToken string
}

func (s *GRPCClient) accessTokenInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {

	md := metadata.New(map[string]string{"access_token": s.accessToken})
	ctx = metadata.NewOutgoingContext(ctx, md)

	err := invoker(ctx, method, req, reply, cc, opts...)

	return err
}

func NewGophKeeperClientService(endpointURL string) (*GRPCClient, error) {
	c := &GRPCClient{endpointURL: endpointURL}
	err := c.InitGRPCClient()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *GRPCClient) InitGRPCClient() error {

	conn, err := grpc.NewClient(s.endpointURL, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(s.accessTokenInterceptor))
	if err != nil {
		return err
	}
	s.conn = conn
	s.client = pb.NewGophKeeperServiceClient(conn)
	return nil
}

func (s *GRPCClient) Register(ctx context.Context, userName string, salt []byte, key []byte) error {

	req := &pb.RegisterUserRequest{Username: userName, Salt: salt, Verifier: key}

	_, err := s.client.RegisterUser(ctx, req)

	if err != nil {
		return s.mapError(err)
	}

	return nil

}

func (s *GRPCClient) GetSalt(ctx context.Context, userName string) ([]byte, error) {

	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()

	req := &pb.GetSaltRequest{Username: userName}

	resp, err := s.client.GetSalt(ctx, req)

	if err != nil {
		return nil, s.mapError(err)
	}
	return resp.Salt, nil
}

func (s *GRPCClient) Login(ctx context.Context, userName string, key []byte) error {

	req := &pb.LoginRequest{Username: userName, VerifierCandidate: key}

	resp, err := s.client.Login(ctx, req)

	if err != nil {
		return s.mapError(err)
	}

	s.accessToken = resp.AccessToken
	s.refreshToken = resp.RefreshToken

	return nil

}

func (s *GRPCClient) Close() error {
	return s.conn.Close()
}

func (s *GRPCClient) AddEntry(ctx context.Context, entryType models.EntryType, title string, сypherText []byte, nonce []byte) error {

	req := &pb.AddEntryRequest{Type: string(entryType), Title: title, Cyphertext: сypherText, Nonce: nonce}

	_, err := s.client.AddEntry(ctx, req)
	if err != nil {
		return s.mapError(err)
	}

	return nil

}

func (s *GRPCClient) GetPresignedPutURL(ctx context.Context) (string, string, error) {
	req := &pb.GetPresignedPutUrlRequest{}

	resp, err := s.client.GetPresignedPutUrl(ctx, req)
	if err != nil {
		return "", "", s.mapError(err)
	}

	return resp.Key, resp.Url, nil

}

func (s *GRPCClient) GetPresignedGetURL(ctx context.Context, key string) (string, error) {
	req := &pb.GetPresignedPutUrlRequest{}

	resp, err := s.client.GetPresignedPutUrl(ctx, req)
	if err != nil {
		return "", s.mapError(err)
	}

	return resp.Url, nil

}

func (s *GRPCClient) mapError(err error) error {
	if err == nil {
		return nil
	}
	st, _ := status.FromError(err)
	switch st.Code() {
	case codes.Unauthenticated, codes.PermissionDenied:
		return ErrUnauthorized
	case codes.Unavailable, codes.DeadlineExceeded:
		return ErrUnavailable
	default:
		return fmt.Errorf("rpc error: %w", err)
	}
}
