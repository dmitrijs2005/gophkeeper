// Package models defines server-side data models persisted in the database.
package models

import "time"

// User represents an account registered in the system.
// Passwords are never stored; instead, a salt and verifier are kept.
type User struct {
	// ID is the unique identifier of the user.
	ID string
	// UserName is the user's login (typically an email).
	UserName string
	// Salt is the per-user random salt used for deriving the master key.
	Salt []byte
	// Verifier is a one-way value derived from the master key (not the password).
	Verifier []byte
	// CreatedAt is the account creation timestamp (UTC).
	CreatedAt time.Time
}
