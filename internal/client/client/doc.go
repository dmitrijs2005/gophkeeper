// Package client contains client-side building blocks for GophKeeper.
//
// # Overview
//
// The package provides:
//  1. A transport-agnostic API contract (see the Client interface) to talk
//     to the GophKeeper backend: Register/GetSalt/Login, Ping, Sync,
//     MarkUploaded, and presigned URL helpers.
//  2. A concrete gRPC implementation (see GRPCClient) that manages a
//     connection, injects an access token via an interceptor, transparently
//     refreshes expired tokens, and maps gRPC status codes to sentinel errors.
//  3. Local persistence bootstrap utilities (InitDatabase, RunMigrations) for
//     the CLI, wiring an SQLite database and applying embedded goose migrations.
//
// # Error Handling
//
// Common conditions are exposed as sentinel errors that callers can match with
// errors.Is: ErrUnavailable, ErrUnauthorized, ErrLocalDataNotAvailable.
//
// Concurrency & Contexts
//
// Implementations should be safe for concurrent use unless stated otherwise.
// All operations accept context.Context and must honor cancellation/timeouts.
//
// See Also
//
//   - Interface:  Client
//   - gRPC impl:  GRPCClient
//   - DB helpers: InitDatabase, RunMigrations
//   - Errors:     ErrUnavailable, ErrUnauthorized, ErrLocalDataNotAvailable
package client
