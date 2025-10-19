// Package models defines server-side data models persisted in the database.
package models

import "time"

// Entry is the canonical server-side representation of a vault entry.
// Encrypted fields store AEAD ciphertext alongside their nonces.
type Entry struct {
	// ID is the globally unique identifier of the entry.
	ID string
	// UserID is the owner of this entry.
	UserID string

	// Overview holds a short, encrypted summary (ciphertext bytes).
	Overview []byte
	// NonceOverview is the AEAD nonce for Overview.
	NonceOverview []byte

	// Details holds the full, encrypted payload (ciphertext bytes).
	Details []byte
	// NonceDetails is the AEAD nonce for Details.
	NonceDetails []byte

	// CreatedAt is the insertion timestamp (UTC).
	CreatedAt time.Time
	// UpdatedAt is the last modification timestamp (UTC).
	UpdatedAt time.Time

	// Deleted marks the entry as a soft-deleted tombstone for sync/GC.
	Deleted bool
	// Version is the server-assigned, monotonically increasing version used for sync.
	Version int64
}
