package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
	"github.com/dmitrijs2005/gophkeeper/internal/server/auth"
	"github.com/dmitrijs2005/gophkeeper/internal/server/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ---- test logger ----

// helper to build server
func newTestServer(secret string) *GRPCServer {
	return &GRPCServer{
		logger:    nopLogger{},
		jwtSecret: []byte(secret),
		users:     (*services.UserService)(nil),
		entries:   (*services.EntryService)(nil),
	}
}

func TestInterceptor_NonSync_AllowsWithoutToken(t *testing.T) {
	s := newTestServer("secret")

	ctx := context.Background()
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Service/OtherMethod"}
	handlerCalled := false

	h := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "ok", nil
	}

	resp, err := s.accessTokenInterceptor(ctx, nil, info, h) // see below note
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handlerCalled {
		t.Fatal("handler was not called")
	}
	if resp != "ok" {
		t.Fatalf("unexpected handler resp: %v", resp)
	}
}

func TestInterceptor_Sync_MissingToken(t *testing.T) {
	s := newTestServer("secret")

	ctx := context.Background()
	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.service.GophKeeperService/Sync"}

	h := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler should not be called when token missing")
		return nil, nil
	}

	_, err := s.accessTokenInterceptor(ctx, nil, info, h)
	if err == nil {
		t.Fatal("expected error")
	}
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", status.Code(err))
	}
	if status.Convert(err).Message() != "missing token" {
		t.Fatalf("expected 'missing token', got %q", status.Convert(err).Message())
	}
}

func TestInterceptor_Sync_InvalidToken(t *testing.T) {
	s := newTestServer("secret")

	md := metadata.New(map[string]string{
		common.AccessTokenHeaderName: "not-a-valid-jwt",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.service.GophKeeperService/Sync"}

	h := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler should not be called for invalid token")
		return nil, nil
	}

	_, err := s.accessTokenInterceptor(ctx, nil, info, h)
	if err == nil {
		t.Fatal("expected error")
	}
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestInterceptor_Sync_ValidToken_SetsUserID(t *testing.T) {
	secret := "super-secret"
	s := newTestServer(secret)

	userID := "user-123"
	token, err := auth.GenerateToken(userID, []byte(secret), time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	md := metadata.New(map[string]string{
		common.AccessTokenHeaderName: token,
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.service.GophKeeperService/Sync"}

	var gotFromCtx any
	h := func(ctx context.Context, req interface{}) (interface{}, error) {
		gotFromCtx = ctx.Value(UserIDKey)
		return "ok", nil
	}

	resp, err := s.accessTokenInterceptor(ctx, nil, info, h)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("unexpected handler resp: %v", resp)
	}
	if gotFromCtx != userID {
		t.Fatalf("user id not propagated in context: got %v want %v", gotFromCtx, userID)
	}
}
