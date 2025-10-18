package logging

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func newTestLogger(t *testing.T) (*SlogLogger, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug, // чтобы логировался и Debug
	})
	l := slog.New(h)
	return NewSlogLogger(l), &buf
}

func TestSlogLogger_Levels_WriteExpectedOutput(t *testing.T) {
	log, buf := newTestLogger(t)
	ctx := context.Background()

	log.Debug(ctx, "dbg", "a", 1)
	log.Info(ctx, "inf", "b", 2)
	log.Warn(ctx, "wrn", "c", 3)
	log.Error(ctx, "err", "d", 4)

	out := buf.String()

	tests := []struct {
		level string
		msg   string
		key   string
		val   string
	}{
		{"DEBUG", "dbg", "a", "1"},
		{"INFO", "inf", "b", "2"},
		{"WARN", "wrn", "c", "3"},
		{"ERROR", "err", "d", "4"},
	}

	for _, tc := range tests {
		if !strings.Contains(out, "level="+tc.level) {
			t.Fatalf("expected line with level=%s in output:\n%s", tc.level, out)
		}
		if !strings.Contains(out, "msg="+tc.msg) {
			t.Fatalf("expected line with msg=%q in output:\n%s", tc.msg, out)
		}
		if !strings.Contains(out, tc.key+"="+tc.val) {
			t.Fatalf("expected attribute %s=%s in output:\n%s", tc.key, tc.val, out)
		}
	}
}

func TestSlogLogger_With_AddsAttributes(t *testing.T) {
	log, buf := newTestLogger(t)
	ctx := context.Background()

	log2 := log.With("req_id", "123", "user", "alice")
	log2.Info(ctx, "hello", "k", "v")

	out := buf.String()
	wantSubs := []string{
		"level=INFO",
		"msg=hello",
		"req_id=123",
		"user=alice",
		"k=v",
	}
	for _, s := range wantSubs {
		if !strings.Contains(out, s) {
			t.Fatalf("expected %q in output, got:\n%s", s, out)
		}
	}
}

func TestSlogLogger_ContextDoesNotPanic(t *testing.T) {
	log, _ := newTestLogger(t)

	ctx := context.TODO()
	log.Info(ctx, "ctx-ok")
	log.Debug(ctx, "ctx-ok")
	log.Warn(ctx, "ctx-ok")
	log.Error(ctx, "ctx-ok")
}
