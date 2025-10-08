package files

import (
	"context"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
)

type Repository interface {
	CreateOrUpdate(ctx context.Context, entry *models.File) error
	DeleteByEntryID(ctx context.Context, id string) error
	GetByEntryID(ctx context.Context, id string) (*models.File, error)
}
