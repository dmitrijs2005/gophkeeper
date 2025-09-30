package models

import "time"

type Entry struct {
	ID            string
	UserID        string
	Title         string
	Type          string
	EncryptedData []byte
	Nonce         []byte
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
