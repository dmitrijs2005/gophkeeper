// Package logging provides adapters and helpers for the project's logging API.
// This file implements a wrapper around the standard library's slog.Logger.
package logging

import (
	"context"
	"log/slog"
)

// SlogLogger adapts a *slog.Logger to the project's Logger interface.
// It forwards calls to the underlying slog logger, using Context-enabled methods.
type SlogLogger struct {
	l *slog.Logger
}

// NewSlogLogger wraps the given *slog.Logger and returns an adapter that
// satisfies the Logger interface.
func NewSlogLogger(l *slog.Logger) *SlogLogger {
	return &SlogLogger{l: l}
}

// Debug logs a debug-level message via slog.
func (s *SlogLogger) Debug(ctx context.Context, msg string, args ...any) {
	s.l.DebugContext(ctx, msg, args...)
}

// Info logs an info-level message via slog.
func (s *SlogLogger) Info(ctx context.Context, msg string, args ...any) {
	s.l.InfoContext(ctx, msg, args...)
}

// Warn logs a warning-level message via slog.
func (s *SlogLogger) Warn(ctx context.Context, msg string, args ...any) {
	s.l.WarnContext(ctx, msg, args...)
}

// Error logs an error-level message via slog.
func (s *SlogLogger) Error(ctx context.Context, msg string, args ...any) {
	s.l.ErrorContext(ctx, msg, args...)
}

// With returns a child logger that includes the supplied key–value pairs
// as structured fields for all subsequent log entries.
func (s *SlogLogger) With(args ...any) Logger {
	return &SlogLogger{l: s.l.With(args...)}
}
