// Package refreshtokens declares the server-side repository contract for
// managing refresh tokens in persistent storage.
package refreshtokens

import (
	"context"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
)

// Repository defines operations for issuing, retrieving, and revoking refresh tokens.
type Repository interface {
	// Create stores a new refresh token for userID with an expiry of now+validity.
	Create(ctx context.Context, userID string, token string, validity time.Duration) error

	// Find looks up a refresh token by its opaque token string and returns its metadata.
	// Implementations should return a not-found error when the token is absent.
	Find(ctx context.Context, token string) (*models.RefreshToken, error)

	// Delete removes a refresh token by its token string. Deleting a non-existent
	// token should not be considered an error.
	Delete(ctx context.Context, token string) error
}
