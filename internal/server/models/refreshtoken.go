// Package models defines server-side data models persisted in the database.
package models

import "time"

// RefreshToken represents a long-lived token used to mint new access tokens.
// It is stored server-side and validated during refresh flows.
type RefreshToken struct {
	// ID is the unique identifier of this refresh token record.
	ID string
	// UserID is the owner of the token.
	UserID string
	// Token is the opaque refresh token string.
	Token string
	// Expires is the time after which the token is no longer valid (UTC).
	Expires time.Time
	// CreatedAt is when the token record was created (UTC).
	CreatedAt time.Time
}
