package grpc

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
	pb "github.com/dmitrijs2005/gophkeeper/internal/proto"
	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
	"github.com/dmitrijs2005/gophkeeper/internal/server/services"
	"github.com/dmitrijs2005/gophkeeper/internal/server/shared"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ---- test logger ----
// type nopLogger struct{}

// func (n nopLogger) Debug(context.Context, string, ...any) {}
// func (n nopLogger) Info(context.Context, string, ...any)  {}
// func (n nopLogger) Warn(context.Context, string, ...any)  {}
// func (n nopLogger) Error(context.Context, string, ...any) {}
// func (n nopLogger) With(...any) logging.Logger            { return n }

// func (n nopLogger) WithContext(context.Context) logging.Logger { return n }

// ---- fakes ----

type fakeUser struct {
	refreshResp *services.TokenPair
	refreshErr  error

	regResp *models.User
	regErr  error

	saltResp []byte
	saltErr  error

	loginResp *services.TokenPair
	loginErr  error
}

func (f *fakeUser) RefreshToken(ctx context.Context, refresh string) (*services.TokenPair, error) {
	return f.refreshResp, f.refreshErr
}
func (f *fakeUser) Register(ctx context.Context, username string, salt []byte, verifier []byte) (*models.User, error) {
	return f.regResp, f.regErr
}
func (f *fakeUser) GetSalt(ctx context.Context, username string) ([]byte, error) {
	return f.saltResp, f.saltErr
}
func (f *fakeUser) Login(ctx context.Context, username string, verifierCandidate []byte) (*services.TokenPair, error) {
	return f.loginResp, f.loginErr
}

type fakeEntry struct {
	syncOut struct {
		processed   []*models.Entry
		newEntries  []*models.Entry
		newFiles    []*models.File
		uploadTasks []*models.FileUploadTask
		maxVersion  int64
		err         error
	}
	markErr error
	url     string
	urlErr  error
}

func (f *fakeEntry) Sync(ctx context.Context, userID string, pendingEntries []*models.Entry, pendingFiles []*models.File,
	clientMaxVersion int64) ([]*models.Entry, []*models.Entry, []*models.File, []*models.FileUploadTask, int64, error) {
	return f.syncOut.processed, f.syncOut.newEntries, f.syncOut.newFiles, f.syncOut.uploadTasks, f.syncOut.maxVersion, f.syncOut.err
}
func (f *fakeEntry) MarkUploaded(ctx context.Context, entryID string) error { return f.markErr }
func (f *fakeEntry) GetPresignedGetURL(ctx context.Context, entryID string) (string, error) {
	return f.url, f.urlErr
}

// ---- helpers ----

func newServer(u userSvc, e entrySvc) *GRPCServer {
	return &GRPCServer{
		address:   "127.0.0.1:0",
		users:     u,
		entries:   e,
		logger:    nopLogger{},
		jwtSecret: []byte("k"),
	}
}

// ---- tests ----

func TestPing_OK(t *testing.T) {
	s := newServer(&fakeUser{}, &fakeEntry{})
	resp, err := s.Ping(context.Background(), &pb.PingRequest{})
	if err != nil {
		t.Fatalf("Ping error: %v", err)
	}
	if resp.GetStatus() != "OK" {
		t.Fatalf("unexpected status: %q", resp.GetStatus())
	}
}

func TestRefreshToken_OK(t *testing.T) {
	u := &fakeUser{
		refreshResp: &services.TokenPair{AccessToken: "a", RefreshToken: "r"},
	}
	s := newServer(u, &fakeEntry{})
	resp, err := s.RefreshToken(context.Background(), &pb.RefreshTokenRequest{RefreshToken: "r0"})
	if err != nil {
		t.Fatalf("RefreshToken error: %v", err)
	}
	if resp.GetAccessToken() != "a" || resp.GetRefreshToken() != "r" {
		t.Fatalf("unexpected tokens: %+v", resp)
	}
}

func TestRefreshToken_InternalOnError(t *testing.T) {
	u := &fakeUser{refreshErr: errors.New("oops")}
	s := newServer(u, &fakeEntry{})
	_, err := s.RefreshToken(context.Background(), &pb.RefreshTokenRequest{RefreshToken: "r0"})
	if status.Code(err) != codes.Internal {
		t.Fatalf("want Internal, got %v (err=%v)", status.Code(err), err)
	}
}

func TestRegisterUser_OK(t *testing.T) {
	u := &fakeUser{regResp: &models.User{ID: "42"}}
	s := newServer(u, &fakeEntry{})
	resp, err := s.RegisterUser(context.Background(), &pb.RegisterUserRequest{
		Username: "u", Salt: []byte("s"), Verifier: []byte("v"),
	})
	if err != nil {
		t.Fatalf("RegisterUser error: %v", err)
	}
	if resp.GetUsername() == "" {
		t.Fatalf("empty response")
	}
}

func TestRegisterUser_InternalOnError(t *testing.T) {
	u := &fakeUser{regErr: errors.New("db down")}
	s := newServer(u, &fakeEntry{})
	_, err := s.RegisterUser(context.Background(), &pb.RegisterUserRequest{
		Username: "u", Salt: []byte("s"), Verifier: []byte("v"),
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("want Internal, got %v", status.Code(err))
	}
}

func TestGetSalt_OK(t *testing.T) {
	u := &fakeUser{saltResp: []byte("SALT123")}
	s := newServer(u, &fakeEntry{})
	resp, err := s.GetSalt(context.Background(), &pb.GetSaltRequest{Username: "u"})
	if err != nil {
		t.Fatalf("GetSalt error: %v", err)
	}
	if !bytes.Equal(resp.GetSalt(), []byte("SALT123")) {
		t.Fatalf("unexpected salt: %q", resp.GetSalt())
	}
}

func TestGetSalt_InternalOnError(t *testing.T) {
	u := &fakeUser{saltErr: errors.New("no user")}
	s := newServer(u, &fakeEntry{})
	_, err := s.GetSalt(context.Background(), &pb.GetSaltRequest{Username: "u"})
	if status.Code(err) != codes.Internal {
		t.Fatalf("want Internal, got %v", status.Code(err))
	}
}

func TestLogin_OK(t *testing.T) {
	u := &fakeUser{loginResp: &services.TokenPair{AccessToken: "A", RefreshToken: "R"}}
	s := newServer(u, &fakeEntry{})
	resp, err := s.Login(context.Background(), &pb.LoginRequest{
		Username: "u", VerifierCandidate: []byte("vv"),
	})
	if err != nil {
		t.Fatalf("Login error: %v", err)
	}
	if resp.GetAccessToken() != "A" || resp.GetRefreshToken() != "R" {
		t.Fatalf("unexpected tokens: %+v", resp)
	}
}

func TestLogin_UnauthorizedAndInternal(t *testing.T) {
	s := newServer(&fakeUser{loginErr: common.ErrorUnauthorized}, &fakeEntry{})
	_, err := s.Login(context.Background(), &pb.LoginRequest{Username: "u", VerifierCandidate: []byte("x")})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("want Unauthenticated, got %v", status.Code(err))
	}

	s2 := newServer(&fakeUser{loginErr: errors.New("boom")}, &fakeEntry{})
	_, err = s2.Login(context.Background(), &pb.LoginRequest{Username: "u", VerifierCandidate: []byte("x")})
	if status.Code(err) != codes.Internal {
		t.Fatalf("want Internal, got %v", status.Code(err))
	}
}

func TestSync_OK(t *testing.T) {
	e := &fakeEntry{}
	e.syncOut.processed = []*models.Entry{
		{ID: "e1", Version: 1, Overview: []byte("o1"), NonceOverview: []byte("n1"), Details: []byte("d1"), NonceDetails: []byte("m1"), Deleted: false},
	}
	e.syncOut.newEntries = []*models.Entry{
		{ID: "e2", Version: 2, Overview: []byte("o2"), NonceOverview: []byte("n2"), Details: []byte("d2"), NonceDetails: []byte("m2"), Deleted: true},
	}
	e.syncOut.newFiles = []*models.File{
		{EntryID: "e3", EncryptedFileKey: []byte("fk"), Nonce: []byte("fn")},
	}
	e.syncOut.uploadTasks = []*models.FileUploadTask{
		{EntryID: "e4", URL: "http://u"},
	}
	e.syncOut.maxVersion = 7

	s := newServer(&fakeUser{}, e)

	ctx := context.WithValue(context.Background(), shared.UserIDKey, "user-1")
	req := &pb.SyncRequest{
		MaxVersion: 1,
		Entries: []*pb.Entry{
			{Id: "e1", Deleted: false, Overview: []byte("O"), NonceOverview: []byte("N"), Details: []byte("D"), NonceDetails: []byte("M")},
		},
		Files: []*pb.File{
			{EntryId: "e1", FileKey: []byte("K"), Nonce: []byte("Z")},
		},
	}

	resp, err := s.Sync(ctx, req)
	if err != nil {
		t.Fatalf("Sync error: %v", err)
	}
	if resp.GetGlobalMaxVersion() != 7 {
		t.Fatalf("unexpected maxVersion: %d", resp.GetGlobalMaxVersion())
	}
	if len(resp.ProcessedEntries) != 1 || resp.ProcessedEntries[0].GetId() != "e1" {
		t.Fatalf("mapped processed entries unexpected: %+v", resp.ProcessedEntries)
	}
	if len(resp.NewEntries) != 1 || resp.NewEntries[0].GetId() != "e2" || !resp.NewEntries[0].GetDeleted() {
		t.Fatalf("mapped new entries unexpected: %+v", resp.NewEntries)
	}
	if len(resp.NewFiles) != 1 || resp.NewFiles[0].GetEntryId() != "e3" {
		t.Fatalf("mapped new files unexpected: %+v", resp.NewFiles)
	}
	if len(resp.UploadTasks) != 1 || resp.UploadTasks[0].GetEntryId() != "e4" || resp.UploadTasks[0].GetUrl() != "http://u" {
		t.Fatalf("mapped upload tasks unexpected: %+v", resp.UploadTasks)
	}
}

func TestSync_ContextMissingUserID(t *testing.T) {
	s := newServer(&fakeUser{}, &fakeEntry{})
	_, err := s.Sync(context.Background(), &pb.SyncRequest{})
	if status.Code(err) != codes.Internal {
		t.Fatalf("want Internal, got %v", status.Code(err))
	}
}

func TestSync_PropagatesErrors(t *testing.T) {
	e := &fakeEntry{}
	e.syncOut.err = common.ErrorUnauthorized
	s := newServer(&fakeUser{}, e)

	ctx := context.WithValue(context.Background(), shared.UserIDKey, "u")
	_, err := s.Sync(ctx, &pb.SyncRequest{})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("want Unauthenticated, got %v", status.Code(err))
	}

	e2 := &fakeEntry{}
	e2.syncOut.err = errors.New("db")
	s2 := newServer(&fakeUser{}, e2)
	_, err = s2.Sync(ctx, &pb.SyncRequest{})
	if status.Code(err) != codes.Internal {
		t.Fatalf("want Internal, got %v", status.Code(err))
	}
}

func TestMarkUploaded_OK_and_Error(t *testing.T) {
	e := &fakeEntry{}
	s := newServer(&fakeUser{}, e)
	if _, err := s.MarkUploaded(context.Background(), &pb.MarkUploadedRequest{EntryId: "e"}); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	e2 := &fakeEntry{markErr: errors.New("boom")}
	s2 := newServer(&fakeUser{}, e2)
	_, err := s2.MarkUploaded(context.Background(), &pb.MarkUploadedRequest{EntryId: "e"})
	if status.Code(err) != codes.Internal {
		t.Fatalf("want Internal, got %v", status.Code(err))
	}
}

func TestGetPresignedGetUrl_OK_and_Error(t *testing.T) {
	e := &fakeEntry{url: "http://ok"}
	s := newServer(&fakeUser{}, e)
	resp, err := s.GetPresignedGetUrl(context.Background(), &pb.GetPresignedGetUrlRequest{EntryId: "e"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp.GetUrl() != "http://ok" {
		t.Fatalf("unexpected url: %q", resp.GetUrl())
	}

	e2 := &fakeEntry{urlErr: errors.New("x")}
	s2 := newServer(&fakeUser{}, e2)
	_, err = s2.GetPresignedGetUrl(context.Background(), &pb.GetPresignedGetUrlRequest{EntryId: "e"})
	if status.Code(err) != codes.Internal {
		t.Fatalf("want Internal, got %v", status.Code(err))
	}
}

func TestTimeoutGuard(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = ctx
}
