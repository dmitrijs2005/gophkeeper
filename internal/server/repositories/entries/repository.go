package entries

import (
	"context"

	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
)

type Repository interface {
	Create(ctx context.Context, entry *models.Entry) (*models.Entry, error)
}
