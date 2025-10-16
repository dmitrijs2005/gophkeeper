package client

import (
	"context"
	"fmt"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
	"github.com/dmitrijs2005/gophkeeper/internal/common"
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

func withAccessToken(ctx context.Context, token string) context.Context {
	md, _ := metadata.FromOutgoingContext(ctx)
	md = md.Copy()
	if md == nil {
		md = metadata.MD{}
	}
	md.Delete(common.AccessTokenHeaderName)
	md.Set(common.AccessTokenHeaderName, token)

	return metadata.NewOutgoingContext(ctx, md)
}

func (s *GRPCClient) accessTokenInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {

	ctx = withAccessToken(ctx, s.accessToken)

	err := invoker(ctx, method, req, reply, cc, opts...)

	if err != nil {

		st, ok := status.FromError(err)
		if !ok {
			return err
		}

		if st.Code() != codes.Unauthenticated {
			return err
		}
		if st.Message() != common.ErrTokenExpired.Error() {
			return err
		}

		if s.refreshToken == "" {
			return err
		}

		refreshTokenResponse, err := s.client.RefreshToken(ctx, &pb.RefreshTokenRequest{RefreshToken: s.refreshToken})
		if err != nil {
			return err
		}

		s.accessToken = refreshTokenResponse.AccessToken
		s.refreshToken = refreshTokenResponse.RefreshToken

		// TOKENS REFRESHED, creating context with new Access Token
		ctx = withAccessToken(ctx, s.accessToken)
		return invoker(ctx, method, req, reply, cc, opts...)

	}

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

// func (s *GRPCClient) AddEntry(ctx context.Context, entryType models.EntryType, title string, сypherText []byte, nonce []byte) error {

// 	req := &pb.AddEntryRequest{Type: string(entryType), Title: title, Cyphertext: сypherText, Nonce: nonce}

// 	_, err := s.client.AddEntry(ctx, req)
// 	if err != nil {
// 		return s.mapError(err)
// 	}

// 	return nil

// }

// func (s *GRPCClient) GetPresignedPutURL(ctx context.Context) (string, string, error) {
// 	req := &pb.GetPresignedPutUrlRequest{}

// 	resp, err := s.client.GetPresignedPutUrl(ctx, req)
// 	if err != nil {
// 		return "", "", s.mapError(err)
// 	}

// 	return resp.Key, resp.Url, nil

// }

// func (s *GRPCClient) GetPresignedGetURL(ctx context.Context, key string) (string, error) {
// 	req := &pb.GetPresignedPutUrlRequest{}

// 	resp, err := s.client.GetPresignedPutUrl(ctx, req)
// 	if err != nil {
// 		return "", s.mapError(err)
// 	}

// 	return resp.Url, nil

// }

func (s *GRPCClient) Ping(ctx context.Context) error {

	req := &pb.PingRequest{}

	resp, err := s.client.Ping(ctx, req)
	if err != nil {
		return s.mapError(err)
	}

	if resp.Status != "OK" {
		return ErrUnavailable
	}

	return nil

}

func (s *GRPCClient) Sync(ctx context.Context,
	entries []*models.Entry, files []*models.File,
	maxVersion int64) ([]*models.Entry, []*models.Entry, []*models.File, []*models.FileUploadTask, int64, error) {

	reqEntries := make([]*pb.Entry, 0, len(entries))

	for _, e := range entries {
		entry := &pb.Entry{Id: e.Id,
			Version:       e.Version,
			Overview:      e.Overview,
			NonceOverview: e.NonceOverview,
			Details:       e.Details,
			NonceDetails:  e.NonceDetails,
			Deleted:       e.Deleted,
			IsFile:        e.IsFile,
		}
		reqEntries = append(reqEntries, entry)
	}

	reqFiles := make([]*pb.File, 0, len(files))
	for _, f := range files {
		file := &pb.File{EntryId: f.EntryID, FileKey: f.EncryptedFileKey, Nonce: f.Nonce}
		reqFiles = append(reqFiles, file)
	}

	req := &pb.SyncRequest{Entries: reqEntries, Files: reqFiles, MaxVersion: maxVersion}

	resp, err := s.client.Sync(ctx, req)
	if err != nil {
		return nil, nil, nil, nil, 0, s.mapError(err)
	}

	v := resp.GlobalMaxVersion

	var pe, ne []*models.Entry

	for _, e := range resp.ProcessedEntries {
		pe = append(pe, &models.Entry{
			Id:            e.Id,
			Version:       e.Version,
			Deleted:       e.Deleted,
			Overview:      e.Overview,
			NonceOverview: e.NonceOverview,
			Details:       e.Details,
			NonceDetails:  e.NonceDetails,
		})
	}

	for _, e := range resp.NewEntries {
		ne = append(ne, &models.Entry{
			Id:            e.Id,
			Version:       e.Version,
			Deleted:       e.Deleted,
			Overview:      e.Overview,
			NonceOverview: e.NonceOverview,
			Details:       e.Details,
			NonceDetails:  e.NonceDetails,
		})
	}

	var nf []*models.File
	for _, f := range resp.NewFiles {
		nf = append(nf, &models.File{EntryID: f.EntryId, EncryptedFileKey: f.FileKey, Nonce: f.Nonce, UploadStatus: "completed"})
	}

	var ut []*models.FileUploadTask
	for _, t := range resp.UploadTasks {
		ut = append(ut, &models.FileUploadTask{EntryID: t.EntryId, URL: t.Url})
	}

	return pe, ne, nf, ut, v, nil

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

func (s *GRPCClient) MarkUploaded(ctx context.Context, entryID string) error {
	req := &pb.MarkUploadedRequest{EntryId: entryID}
	_, err := s.client.MarkUploaded(ctx, req)
	if err != nil {
		return s.mapError(err)
	}
	return nil
}

func (s *GRPCClient) GetPresignedGetURL(ctx context.Context, entryID string) (string, error) {
	req := &pb.GetPresignedGetUrlRequest{EntryId: entryID}
	res, err := s.client.GetPresignedGetUrl(ctx, req)
	if err != nil {
		return "", s.mapError(err)
	}
	return res.Url, nil

}
