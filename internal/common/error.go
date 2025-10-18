// Package common defines shared constants and sentinel errors used across
// client and server layers of GophKeeper. Callers should use errors.Is to
// match these values.
package common

import "errors"

var (
	// Repository-level errors.
	ErrorNotFound = errors.New("not found")

	// Service-level errors (generic/internal flow control).
	ErrorInternal      = errors.New("internal error")
	ErrorUnauthorized  = errors.New("unauthorized")
	ErrVersionConflict = errors.New("version conflict")

	// Validation / item-specific errors.
	ErrorIncorrectMetadata = errors.New("incorrect metadata")

	// Auth errors (invalid or malformed token).
	ErrInvalidToken = errors.New("invalid token")

	// Token lifecycle errors.
	ErrTokenExpired        = errors.New("token expired")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
)
