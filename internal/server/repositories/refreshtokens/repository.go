package refreshtokens

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, userID string, token string, validity time.Duration) error
	// Find(ctx context.Context, token string) (*RefreshToken, error)
	// Delete(ctx context.Context, token string) error
}
