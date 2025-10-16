package files

import (
	"context"

	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
)

type Repository interface {
	CreateOrUpdate(ctx context.Context, file *models.File) error
	SelectUpdated(ctx context.Context, userID string, minVersion int64) ([]*models.File, error)
	MarkUploaded(ctx context.Context, id string) error
}
