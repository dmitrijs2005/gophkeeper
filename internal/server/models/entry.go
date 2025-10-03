package models

import "time"

type Entry struct {
	ID            string
	UserID        string
	Overview      []byte
	NonceOverview []byte
	Details       []byte
	NonceDetails  []byte
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Deleted       bool
	Version       int64
}
