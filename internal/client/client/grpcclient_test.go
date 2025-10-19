package client

import (
	"context"
	"errors"
	"testing"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
	"github.com/dmitrijs2005/gophkeeper/internal/common"
	pb "github.com/dmitrijs2005/gophkeeper/internal/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

/*************
 * Fake pb client
 *************/

type fakePB struct {
	// inputs captured
	lastRefreshTokenReq *pb.RefreshTokenRequest
	lastPingReq         *pb.PingRequest
	lastGetSaltReq      *pb.GetSaltRequest
	lastLoginReq        *pb.LoginRequest
	lastRegisterReq     *pb.RegisterUserRequest
	lastSyncReq         *pb.SyncRequest
	lastMarkUploadedReq *pb.MarkUploadedRequest
	lastGetURLReq       *pb.GetPresignedGetUrlRequest

	// outputs preset
	refreshTokenResp *pb.RefreshTokenResponse
	refreshTokenErr  error

	pingResp *pb.PingResponse
	pingErr  error

	getSaltResp *pb.GetSaltResponse
	getSaltErr  error

	loginResp *pb.LoginResponse
	loginErr  error

	registerErr error

	syncResp *pb.SyncResponse
	syncErr  error

	markUploadedErr error

	getURLResp *pb.GetPresignedGetUrlResponse
	getURLErr  error
}

func (f *fakePB) RefreshToken(ctx context.Context, in *pb.RefreshTokenRequest, opts ...grpc.CallOption) (*pb.RefreshTokenResponse, error) {
	f.lastRefreshTokenReq = in
	return f.refreshTokenResp, f.refreshTokenErr
}
func (f *fakePB) Ping(ctx context.Context, in *pb.PingRequest, opts ...grpc.CallOption) (*pb.PingResponse, error) {
	f.lastPingReq = in
	return f.pingResp, f.pingErr
}
func (f *fakePB) GetSalt(ctx context.Context, in *pb.GetSaltRequest, opts ...grpc.CallOption) (*pb.GetSaltResponse, error) {
	f.lastGetSaltReq = in
	return f.getSaltResp, f.getSaltErr
}
func (f *fakePB) Login(ctx context.Context, in *pb.LoginRequest, opts ...grpc.CallOption) (*pb.LoginResponse, error) {
	f.lastLoginReq = in
	return f.loginResp, f.loginErr
}
func (f *fakePB) RegisterUser(ctx context.Context, in *pb.RegisterUserRequest, opts ...grpc.CallOption) (*pb.RegisterUserResponse, error) {
	f.lastRegisterReq = in
	return &pb.RegisterUserResponse{}, f.registerErr
}
func (f *fakePB) Sync(ctx context.Context, in *pb.SyncRequest, opts ...grpc.CallOption) (*pb.SyncResponse, error) {
	f.lastSyncReq = in
	return f.syncResp, f.syncErr
}
func (f *fakePB) MarkUploaded(ctx context.Context, in *pb.MarkUploadedRequest, opts ...grpc.CallOption) (*pb.MarkUploadedResponse, error) {
	f.lastMarkUploadedReq = in
	return &pb.MarkUploadedResponse{}, f.markUploadedErr
}
func (f *fakePB) GetPresignedGetUrl(ctx context.Context, in *pb.GetPresignedGetUrlRequest, opts ...grpc.CallOption) (*pb.GetPresignedGetUrlResponse, error) {
	f.lastGetURLReq = in
	return f.getURLResp, f.getURLErr
}

/*************
 * accessTokenInterceptor tests
 *************/

func TestInterceptor_RefreshesTokenOnExpiredAndRetries(t *testing.T) {
	f := &fakePB{
		refreshTokenResp: &pb.RefreshTokenResponse{AccessToken: "A2", RefreshToken: "R2"},
	}
	c := &GRPCClient{
		client:       f,
		accessToken:  "A1",
		refreshToken: "R1",
	}

	callCount := 0
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		callCount++
		md, _ := metadata.FromOutgoingContext(ctx)
		toks := md.Get(common.AccessTokenHeaderName)
		require.Len(t, toks, 1)

		if callCount == 1 {
			require.Equal(t, "A1", toks[0])
			return status.Error(codes.Unauthenticated, common.ErrTokenExpired.Error())
		}
		require.Equal(t, "A2", toks[0])
		return nil
	}

	err := c.accessTokenInterceptor(context.Background(), "/svc/Method", nil, nil, nil, invoker)
	require.NoError(t, err)
	require.Equal(t, 2, callCount)
	require.Equal(t, "A2", c.accessToken)
	require.Equal(t, "R2", c.refreshToken)
	require.Equal(t, "R1", f.lastRefreshTokenReq.RefreshToken)
}

func TestInterceptor_NoRefreshIfNoRefreshToken(t *testing.T) {
	f := &fakePB{}
	c := &GRPCClient{
		client:      f,
		accessToken: "A1",
	}

	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return status.Error(codes.Unauthenticated, common.ErrTokenExpired.Error())
	}

	err := c.accessTokenInterceptor(context.Background(), "/svc/Method", nil, nil, nil, invoker)
	require.Error(t, err)
	require.Nil(t, f.lastRefreshTokenReq)
}

func TestInterceptor_IgnoresOtherErrors(t *testing.T) {
	c := &GRPCClient{accessToken: "X"}
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return status.Error(codes.Internal, "boom")
	}
	err := c.accessTokenInterceptor(context.Background(), "/svc/Method", nil, nil, nil, invoker)
	require.Error(t, err)
}

func TestInterceptor_UnauthenticatedButDifferentMessage_NoRefresh(t *testing.T) {
	c := &GRPCClient{accessToken: "X", refreshToken: "R"}
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return status.Error(codes.Unauthenticated, "some other reason")
	}
	err := c.accessTokenInterceptor(context.Background(), "/svc/Method", nil, nil, nil, invoker)
	require.Error(t, err)
}

/*************
 * mapError tests
 *************/

func TestMapError(t *testing.T) {
	c := &GRPCClient{}

	require.Equal(t, ErrUnauthorized, c.mapError(status.Error(codes.Unauthenticated, "x")))
	require.Equal(t, ErrUnauthorized, c.mapError(status.Error(codes.PermissionDenied, "x")))
	require.Equal(t, ErrUnavailable, c.mapError(status.Error(codes.Unavailable, "x")))
	require.Equal(t, ErrUnavailable, c.mapError(status.Error(codes.DeadlineExceeded, "x")))
	e := errors.New("plain")
	require.ErrorContains(t, c.mapError(e), "rpc error:")
}

/*************
 * Ping tests
 *************/

func TestPing_OK(t *testing.T) {
	f := &fakePB{pingResp: &pb.PingResponse{Status: "OK"}}
	c := &GRPCClient{client: f}
	require.NoError(t, c.Ping(context.Background()))
}

func TestPing_NotOK_ReturnsUnavailable(t *testing.T) {
	f := &fakePB{pingResp: &pb.PingResponse{Status: "NOT_OK"}}
	c := &GRPCClient{client: f}
	require.ErrorIs(t, c.Ping(context.Background()), ErrUnavailable)
}

func TestPing_MapsRPCError(t *testing.T) {
	f := &fakePB{pingErr: status.Error(codes.Unavailable, "down")}
	c := &GRPCClient{client: f}
	require.ErrorIs(t, c.Ping(context.Background()), ErrUnavailable)
}

/*************
 * GetSalt / Login / Register tests
 *************/

func TestGetSalt_Success(t *testing.T) {
	f := &fakePB{getSaltResp: &pb.GetSaltResponse{Salt: []byte{1, 2, 3}}}
	c := &GRPCClient{client: f}
	salt, err := c.GetSalt(context.Background(), "u")
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2, 3}, salt)
	require.Equal(t, "u", f.lastGetSaltReq.Username)
}

func TestGetSalt_MapsError(t *testing.T) {
	f := &fakePB{getSaltErr: status.Error(codes.Unavailable, "x")}
	c := &GRPCClient{client: f}
	_, err := c.GetSalt(context.Background(), "u")
	require.ErrorIs(t, err, ErrUnavailable)
}

func TestLogin_SetsTokens(t *testing.T) {
	f := &fakePB{loginResp: &pb.LoginResponse{AccessToken: "A", RefreshToken: "R"}}
	c := &GRPCClient{client: f}
	require.NoError(t, c.Login(context.Background(), "u", []byte{9}))
	require.Equal(t, "A", c.accessToken)
	require.Equal(t, "R", c.refreshToken)
	require.Equal(t, "u", f.lastLoginReq.Username)
	require.Equal(t, []byte{9}, f.lastLoginReq.VerifierCandidate)
}

func TestRegister_MapsError(t *testing.T) {
	f := &fakePB{registerErr: status.Error(codes.PermissionDenied, "no")}
	c := &GRPCClient{client: f}
	err := c.Register(context.Background(), "u", []byte{1}, []byte{2})
	require.ErrorIs(t, err, ErrUnauthorized)
	require.Equal(t, "u", f.lastRegisterReq.Username)
	require.Equal(t, []byte{1}, f.lastRegisterReq.Salt)
	require.Equal(t, []byte{2}, f.lastRegisterReq.Verifier)
}

/*************
 * Sync tests
 *************/

func TestSync_MapsReqAndResp(t *testing.T) {
	entries := []*models.Entry{
		{Id: "e1", Version: 1, Overview: []byte("ov1"), NonceOverview: []byte("no1"), Details: []byte("d1"), NonceDetails: []byte("nd1"), Deleted: false, IsFile: false},
	}
	files := []*models.File{
		{EntryID: "e1", EncryptedFileKey: []byte("fk"), Nonce: []byte("fn")},
	}

	f := &fakePB{
		syncResp: &pb.SyncResponse{
			ProcessedEntries: []*pb.Entry{
				{Id: "e1", Version: 2, Deleted: false, Overview: []byte("ov2"), NonceOverview: []byte("no2"), Details: []byte("d2"), NonceDetails: []byte("nd2")},
			},
			NewEntries: []*pb.Entry{
				{Id: "e2", Version: 1, Deleted: false, Overview: []byte("ovN"), NonceOverview: []byte("noN"), Details: []byte("dN"), NonceDetails: []byte("ndN")},
			},
			NewFiles: []*pb.File{
				{EntryId: "e2", FileKey: []byte("fk2"), Nonce: []byte("fn2")},
			},
			UploadTasks: []*pb.UploadTask{
				{EntryId: "e3", Url: "https://u"},
			},
			GlobalMaxVersion: 42,
		},
	}
	c := &GRPCClient{client: f}

	pe, ne, nf, ut, v, err := c.Sync(context.Background(), entries, files, 7)
	require.NoError(t, err)
	require.EqualValues(t, 42, v)

	require.Equal(t, int64(7), f.lastSyncReq.MaxVersion)
	require.Len(t, f.lastSyncReq.Entries, 1)
	require.Equal(t, "e1", f.lastSyncReq.Entries[0].Id)
	require.Equal(t, []byte("ov1"), f.lastSyncReq.Entries[0].Overview)
	require.Len(t, f.lastSyncReq.Files, 1)
	require.Equal(t, "e1", f.lastSyncReq.Files[0].EntryId)
	require.Equal(t, []byte("fk"), f.lastSyncReq.Files[0].FileKey)

	require.Len(t, pe, 1)
	require.Equal(t, "e1", pe[0].Id)
	require.Equal(t, []byte("ov2"), pe[0].Overview)
	require.Len(t, ne, 1)
	require.Equal(t, "e2", ne[0].Id)
	require.Len(t, nf, 1)
	require.Equal(t, "e2", nf[0].EntryID)
	require.Equal(t, "completed", nf[0].UploadStatus)
	require.Len(t, ut, 1)
	require.Equal(t, "e3", ut[0].EntryID)
	require.Equal(t, "https://u", ut[0].URL)
}

func TestSync_MapsError(t *testing.T) {
	f := &fakePB{syncErr: status.Error(codes.Unavailable, "x")}
	c := &GRPCClient{client: f}
	_, _, _, _, _, err := c.Sync(context.Background(), nil, nil, 0)
	require.ErrorIs(t, err, ErrUnavailable)
}

/*************
 * MarkUploaded / GetPresignedGetURL tests
 *************/

func TestMarkUploaded_Success(t *testing.T) {
	f := &fakePB{}
	c := &GRPCClient{client: f}
	require.NoError(t, c.MarkUploaded(context.Background(), "e1"))
	require.Equal(t, "e1", f.lastMarkUploadedReq.EntryId)
}

func TestMarkUploaded_MapsError(t *testing.T) {
	f := &fakePB{markUploadedErr: status.Error(codes.PermissionDenied, "x")}
	c := &GRPCClient{client: f}
	require.ErrorIs(t, c.MarkUploaded(context.Background(), "e1"), ErrUnauthorized)
}

func TestGetPresignedGetURL_Success(t *testing.T) {
	f := &fakePB{getURLResp: &pb.GetPresignedGetUrlResponse{Url: "https://dl"}}
	c := &GRPCClient{client: f}
	url, err := c.GetPresignedGetURL(context.Background(), "e1")
	require.NoError(t, err)
	require.Equal(t, "https://dl", url)
	require.Equal(t, "e1", f.lastGetURLReq.EntryId)
}

func TestGetPresignedGetURL_MapsError(t *testing.T) {
	f := &fakePB{getURLErr: status.Error(codes.Unavailable, "x")}
	c := &GRPCClient{client: f}
	_, err := c.GetPresignedGetURL(context.Background(), "e1")
	require.ErrorIs(t, err, ErrUnavailable)
}
