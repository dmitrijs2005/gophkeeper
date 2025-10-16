package client

import (
	"context"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
)

type Client interface {
	Close() error
	Register(ctx context.Context, username string, salt []byte, key []byte) error
	GetSalt(ctx context.Context, username string) ([]byte, error)
	Login(ctx context.Context, username string, key []byte) error
	Ping(ctx context.Context) error
	Sync(ctx context.Context, entries []*models.Entry, files []*models.File, maxVersion int64) ([]*models.Entry, []*models.Entry, []*models.File, []*models.FileUploadTask, int64, error)
	MarkUploaded(ctx context.Context, entryID string) error
}
