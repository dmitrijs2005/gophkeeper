// Package entries declares the server-side repository contract for working
// with entry records in persistent storage.
package entries

import (
	"context"

	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
)

// Repository defines persistence operations for encrypted entry records.
type Repository interface {
	// CreateOrUpdate inserts a new entry or updates an existing one by ID
	// (scoped to its owner), maintaining the entry's version.
	CreateOrUpdate(ctx context.Context, entry *models.Entry) error

	// SelectUpdated returns all entries for the given user whose version is
	// strictly greater than minVersion (used for incremental sync).
	SelectUpdated(ctx context.Context, userID string, minVersion int64) ([]*models.Entry, error)
}
