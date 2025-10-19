// Package files declares the server-side repository contract for working with
// file metadata stored in the database (sync state, keys/nonces, storage key).
package files

import (
	"context"

	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
)

// Repository defines persistence operations for server-side File records.
type Repository interface {
	// CreateOrUpdate inserts a new file row or updates an existing one by entry_id.
	CreateOrUpdate(ctx context.Context, file *models.File) error

	// SelectUpdated returns files for the given user with version strictly greater
	// than minVersion (used for incremental synchronization).
	SelectUpdated(ctx context.Context, userID string, minVersion int64) ([]*models.File, error)

	// MarkUploaded marks the file for the given entry as uploaded (e.g., sets status
	// to "completed").
	MarkUploaded(ctx context.Context, id string) error

	// GetByEntryID returns minimal file metadata for authorization and URL generation.
	GetByEntryID(ctx context.Context, id string) (*models.File, error)
}
