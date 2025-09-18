package service

import (
	"context"

	"github.com/dmitrijs2005/gophkeeper/internal/client/crypto"
	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
	pb "github.com/dmitrijs2005/gophkeeper/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type GophKeeperClientService struct {
	endprointURL string
	conn         *grpc.ClientConn
	client       pb.GophKeeperServiceClient
	accessToken  string
	refreshToken string
}

func (s *GophKeeperClientService) accessTokenInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {

	md := metadata.New(map[string]string{"access_token": s.accessToken})
	ctx = metadata.NewOutgoingContext(context.Background(), md)

	err := invoker(ctx, method, req, reply, cc, opts...)

	return err
}

func NewGophKeeperClientService(endprointURL string) (*GophKeeperClientService, error) {
	return &GophKeeperClientService{endprointURL: endprointURL}, nil
}

func (s *GophKeeperClientService) InitGRPCClient() error {
	conn, err := grpc.NewClient(s.endprointURL, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(s.accessTokenInterceptor))
	if err != nil {
		return err
	}
	s.conn = conn
	s.client = pb.NewGophKeeperServiceClient(conn)
	return nil
}

func (s *GophKeeperClientService) Register(ctx context.Context, userName string, password []byte) error {

	salt := crypto.GenerateSalt(32)
	key := crypto.DeriveMasterKey(password, salt)

	req := &pb.RegisterUserRequest{Username: userName, Salt: salt, Verifier: key}

	_, err := s.client.RegisterUser(ctx, req)

	if err != nil {
		return err
	}

	return nil

}

func (s *GophKeeperClientService) GetSalt(ctx context.Context, userName string) ([]byte, error) {
	req := &pb.GetSaltRequest{Username: userName}
	resp, err := s.client.GetSalt(ctx, req)

	if err != nil {
		return nil, err
	}
	return resp.Salt, nil
}

func (s *GophKeeperClientService) Login(ctx context.Context, userName string, key []byte) error {

	req := &pb.LoginRequest{Username: userName, VerifierCandidate: key}

	resp, err := s.client.Login(ctx, req)

	if err != nil {
		return err
	}

	s.accessToken = resp.AccessToken
	s.refreshToken = resp.RefreshToken

	return nil

}

func (s *GophKeeperClientService) Close() error {
	return s.conn.Close()
}

func (s *GophKeeperClientService) AddEntry(ctx context.Context, entryType models.EntryType, title string, сypherText []byte, nonce []byte) error {

	req := &pb.AddEntryRequest{Type: string(entryType), Title: title, Cyphertext: сypherText, Nonce: nonce}

	_, err := s.client.AddEntry(ctx, req)
	if err != nil {
		return err
	}

	return nil

}
