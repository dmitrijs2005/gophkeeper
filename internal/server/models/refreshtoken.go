package models

import "time"

type RefreshToken struct {
	ID        string
	UserID    string
	Token     string
	Expires   time.Time
	CreatedAt time.Time
}
