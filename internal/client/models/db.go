// Package models defines client-side data models used by the GophKeeper CLI.
package models

import "time"

// Entry is a versioned envelope persisted locally and synced with the server.
// Encrypted fields store AEAD ciphertext alongside their nonces.
type Entry struct {
	// Id is a globally unique identifier for the entry.
	Id string

	// Version is the monotonic, server-assigned version used for sync/merge.
	Version int64

	// Deleted marks the entry as a tombstone (kept for conflict-free sync).
	Deleted bool

	// Overview contains encrypted, short summary bytes (human preview).
	Overview []byte
	// NonceOverview is the AEAD nonce for Overview.
	NonceOverview []byte

	// Details contains encrypted, full payload bytes (type-specific).
	Details []byte
	// NonceDetails is the AEAD nonce for Details.
	NonceDetails []byte

	// UpdatedAt is the last modification time in UTC.
	UpdatedAt time.Time

	// IsFile indicates that this entry represents a binary file payload.
	IsFile bool
}
