package client

import (
	"context"
)

type Client interface {
	Close() error
	Register(ctx context.Context, username string, salt []byte, key []byte) error
	GetSalt(ctx context.Context, username string) ([]byte, error)
	Login(ctx context.Context, username string, key []byte) error
	//AddEntry(ctx context.Context, entryType models.EntryType, title string, сypherText []byte, nonce []byte) error
	GetPresignedPutURL(ctx context.Context) (string, string, error)
	GetPresignedGetURL(ctx context.Context, key string) (string, error)
}
