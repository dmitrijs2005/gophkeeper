package entries

import (
	"context"

	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
)

type Repository interface {
	CreateOrUpdate(ctx context.Context, entry *models.Entry) error
	SelectUpdated(ctx context.Context, userID string, minVersion int64) ([]*models.Entry, error)
}
