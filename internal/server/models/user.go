package models

import "time"

// type User struct {
// 	ID                string    `db:"id"`
// 	UserName          string    `db:"username"`
// 	Salt              string    `db:"salt"`
// 	MasterKeyVerifier []byte    `db:"master_key_verifier"`
// 	CreatedAt         time.Time `db:"created_at"`
// }

type User struct {
	ID        string
	UserName  string
	Salt      []byte
	Verifier  []byte
	CreatedAt time.Time
}
