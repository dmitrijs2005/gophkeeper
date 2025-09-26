package models

import "time"

type Entry struct {
	Id            string
	Version       int64
	Deleted       bool
	Overview      []byte
	NonceOverview []byte
	Details       []byte
	NonceDetails  []byte
	UpdatedAr     time.Time
}
