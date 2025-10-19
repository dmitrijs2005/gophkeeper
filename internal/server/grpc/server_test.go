package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/logging"
	"github.com/dmitrijs2005/gophkeeper/internal/server/services"
)

type nopLogger struct{}

func (n nopLogger) Debug(context.Context, string, ...any) {}
func (n nopLogger) Info(context.Context, string, ...any)  {}
func (n nopLogger) Warn(context.Context, string, ...any)  {}
func (n nopLogger) Error(context.Context, string, ...any) {}
func (n nopLogger) With(...any) logging.Logger            { return n }

func TestRun_StopsOnContextCancel(t *testing.T) {
	t.Parallel()

	srv, err := NewgGRPCServer("127.0.0.1:0", nopLogger{}, (*services.UserService)(nil), (*services.EntryService)(nil), "secret")
	if err != nil {
		t.Fatalf("NewgGRPCServer error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- srv.Run(ctx)
	}()

	select {
	case err := <-done:
		t.Fatalf("server exited too early: %v", err)
	case <-time.After(150 * time.Millisecond):
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error on graceful stop: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop within timeout after context cancel")
	}
}

func TestRun_ReturnsErrorOnBadAddress(t *testing.T) {
	t.Parallel()

	srv, err := NewgGRPCServer("127.0.0.1:99999", nopLogger{}, (*services.UserService)(nil), (*services.EntryService)(nil), "secret")
	if err != nil {
		t.Fatalf("NewgGRPCServer error (constructor should not fail here): %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := srv.Run(ctx); err == nil {
		t.Fatal("expected error from Run on bad address, got nil")
	}
}
