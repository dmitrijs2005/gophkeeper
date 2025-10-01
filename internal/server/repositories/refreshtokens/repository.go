package refreshtokens

import (
	"context"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
)

type Repository interface {
	Create(ctx context.Context, userID string, token string, validity time.Duration) error
	Find(ctx context.Context, token string) (*models.RefreshToken, error)
	Delete(ctx context.Context, token string) error
}
