package logging

import (
	"context"
	"log/slog"
)

type SlogLogger struct {
	l *slog.Logger
}

func NewSlogLogger(l *slog.Logger) *SlogLogger {
	return &SlogLogger{l: l}
}

func (s *SlogLogger) Debug(ctx context.Context, msg string, args ...any) {
	s.l.DebugContext(ctx, msg, args...)
}

func (s *SlogLogger) Info(ctx context.Context, msg string, args ...any) {
	s.l.InfoContext(ctx, msg, args...)
}

func (s *SlogLogger) Warn(ctx context.Context, msg string, args ...any) {
	s.l.WarnContext(ctx, msg, args...)
}

func (s *SlogLogger) Error(ctx context.Context, msg string, args ...any) {
	s.l.ErrorContext(ctx, msg, args...)
}

func (s *SlogLogger) With(args ...any) Logger {
	return &SlogLogger{l: s.l.With(args...)}
}
