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

// GRPCClient implements the Client interface over gRPC.
// It keeps the current access/refresh tokens and injects the access token
// into outgoing requests via a unary interceptor.
type GRPCClient struct {
	endpointURL  string
	conn         *grpc.ClientConn
	client       pb.GophKeeperServiceClient
	accessToken  string
	refreshToken string
}

// withAccessToken returns a child context that carries the provided access token
// in the gRPC outgoing metadata under common.AccessTokenHeaderName.
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

// accessTokenInterceptor injects the current access token and, on receiving
// an Unauthenticated error with an "expired" message, attempts a refresh and
// retries the original RPC once with the new token.
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
	if err == nil {
		return nil
	}

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

	// Refresh tokens and retry once.
	refreshTokenResponse, rerr := s.client.RefreshToken(ctx, &pb.RefreshTokenRequest{RefreshToken: s.refreshToken})
	if rerr != nil {
		return rerr
	}
	s.accessToken = refreshTokenResponse.AccessToken
	s.refreshToken = refreshTokenResponse.RefreshToken

	ctx = withAccessToken(ctx, s.accessToken)
	return invoker(ctx, method, req, reply, cc, opts...)
}

// NewGophKeeperClientService constructs a GRPCClient for the given endpoint URL
// and initializes the gRPC connection and service stub.
func NewGophKeeperClientService(endpointURL string) (*GRPCClient, error) {
	c := &GRPCClient{endpointURL: endpointURL}
	if err := c.InitGRPCClient(); err != nil {
		return nil, err
	}
	return c, nil
}

// InitGRPCClient dials the server with insecure credentials (dev use) and
// installs the access-token interceptor. It also creates the typed service client.
func (s *GRPCClient) InitGRPCClient() error {
	conn, err := grpc.NewClient(
		s.endpointURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(s.accessTokenInterceptor),
	)
	if err != nil {
		return err
	}
	s.conn = conn
	s.client = pb.NewGophKeeperServiceClient(conn)
	return nil
}

// Register creates a new user account by sending username, salt, and verifier key.
func (s *GRPCClient) Register(ctx context.Context, userName string, salt []byte, key []byte) error {
	req := &pb.RegisterUserRequest{Username: userName, Salt: salt, Verifier: key}
	if _, err := s.client.RegisterUser(ctx, req); err != nil {
		return s.mapError(err)
	}
	return nil
}

// GetSalt fetches the server-stored salt for the given username.
// A 12s timeout is applied to the request context.
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

// Login authenticates with the server using the local verifier candidate,
// caching returned access/refresh tokens on success.
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

// Close closes the underlying gRPC connection.
func (s *GRPCClient) Close() error {
	return s.conn.Close()
}

// Ping performs a liveness probe; it maps non-OK statuses to ErrUnavailable.
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

// Sync performs bidirectional synchronization of entries/files with the server.
// It converts local models to protobuf messages, calls the RPC, and maps the
// response back to local models along with the new global version and any file
// upload tasks the client should fulfill.
func (s *GRPCClient) Sync(
	ctx context.Context,
	entries []*models.Entry,
	files []*models.File,
	maxVersion int64,
) (
	processedEntries []*models.Entry,
	newEntries []*models.Entry,
	newFiles []*models.File,
	uploadTasks []*models.FileUploadTask,
	newMaxVersion int64,
	err error,
) {
	reqEntries := make([]*pb.Entry, 0, len(entries))
	for _, e := range entries {
		reqEntries = append(reqEntries, &pb.Entry{
			Id:            e.Id,
			Version:       e.Version,
			Overview:      e.Overview,
			NonceOverview: e.NonceOverview,
			Details:       e.Details,
			NonceDetails:  e.NonceDetails,
			Deleted:       e.Deleted,
			IsFile:        e.IsFile,
		})
	}

	reqFiles := make([]*pb.File, 0, len(files))
	for _, f := range files {
		reqFiles = append(reqFiles, &pb.File{
			EntryId: f.EntryID,
			FileKey: f.EncryptedFileKey,
			Nonce:   f.Nonce,
		})
	}

	req := &pb.SyncRequest{Entries: reqEntries, Files: reqFiles, MaxVersion: maxVersion}
	resp, callErr := s.client.Sync(ctx, req)
	if callErr != nil {
		return nil, nil, nil, nil, 0, s.mapError(callErr)
	}

	v := resp.GlobalMaxVersion

	var pe []*models.Entry
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

	var ne []*models.Entry
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
		nf = append(nf, &models.File{
			EntryID:          f.EntryId,
			EncryptedFileKey: f.FileKey,
			Nonce:            f.Nonce,
			UploadStatus:     "completed",
		})
	}

	var ut []*models.FileUploadTask
	for _, t := range resp.UploadTasks {
		ut = append(ut, &models.FileUploadTask{
			EntryID: t.EntryId,
			URL:     t.Url,
		})
	}

	return pe, ne, nf, ut, v, nil
}

// mapError converts gRPC status errors to package-level sentinel errors
// (ErrUnauthorized, ErrUnavailable) or wraps the original error otherwise.
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

// MarkUploaded notifies the server that the file for the given entry has been
// uploaded successfully, which typically clears the pending upload task.
func (s *GRPCClient) MarkUploaded(ctx context.Context, entryID string) error {
	req := &pb.MarkUploadedRequest{EntryId: entryID}
	if _, err := s.client.MarkUploaded(ctx, req); err != nil {
		return s.mapError(err)
	}
	return nil
}

// GetPresignedGetURL requests a temporary signed URL for downloading the
// encrypted file associated with entryID.
func (s *GRPCClient) GetPresignedGetURL(ctx context.Context, entryID string) (string, error) {
	req := &pb.GetPresignedGetUrlRequest{EntryId: entryID}
	res, err := s.client.GetPresignedGetUrl(ctx, req)
	if err != nil {
		return "", s.mapError(err)
	}
	return res.Url, nil
}
