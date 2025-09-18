package service

import (
	"context"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
)

type Service interface {
	Close() error
	Register(ctx context.Context, username string, password []byte) error
	GetSalt(ctx context.Context, username string) ([]byte, error)
	Login(ctx context.Context, username string, key []byte) error
	AddEntry(ctx context.Context, entryType models.EntryType, title string, сypherText []byte, nonce []byte) error
}
