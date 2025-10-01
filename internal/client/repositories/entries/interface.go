package entries

import (
	"context"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
)

type Repository interface {
	Insert(ctx context.Context, entry *models.Entry) error
	GetAll(ctx context.Context) ([]models.Entry, error)
	DeleteByID(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*models.Entry, error)
	GetAllPending(ctx context.Context) ([]*models.Entry, error)
}
