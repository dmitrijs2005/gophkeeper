// Package users declares the server-side repository contract for working with
// user accounts and sync-related metadata.
package users

import (
	"context"

	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
)

// Repository defines persistence operations for User entities.
type Repository interface {
	// Create inserts a new user and returns it with a populated ID.
	Create(ctx context.Context, user *models.User) (*models.User, error)

	// GetUserByLogin fetches a user by login (username). Should return a
	// not-found error when the user does not exist.
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)

	// IncrementCurrentVersion atomically increments and returns the user's
	// current_version counter used for synchronization.
	IncrementCurrentVersion(ctx context.Context, userID string) (int64, error)
}
