// Package logging defines a minimal structured-logging interface used across
// the project. Implementations can wrap slog, zap, zerolog, etc.
package logging

import "context"

// Logger is a context-aware, structured logger.
//
// The variadic args are interpreted as key–value pairs, e.g.:
//
//	log.Info(ctx, "starting server", "addr", addr, "mode", mode)
type Logger interface {
	// Info logs an informational message.
	Info(ctx context.Context, msg string, args ...any)

	// Warn logs a warning message for unusual but non-fatal conditions.
	Warn(ctx context.Context, msg string, args ...any)

	// Error logs an error message for failures.
	Error(ctx context.Context, msg string, args ...any)

	// With returns a child logger that always includes the given key–value pairs.
	With(args ...any) Logger
}
