package users

import "time"

type User struct {
	ID        string
	UserName  string
	Salt      []byte
	Verifier  []byte
	CreatedAt time.Time
}
