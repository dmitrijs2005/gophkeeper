package files

import (
	"context"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
)

// Repository describes CRUD and workflow operations for File records.
// Implementations are typically backed by a local SQLite database.
type Repository interface {
	// CreateOrUpdate inserts or updates a file record for an entry.
	CreateOrUpdate(ctx context.Context, entry *models.File) error

	// DeleteByEntryID removes (or marks deleted) the file record linked to the entry.
	DeleteByEntryID(ctx context.Context, id string) error

	// GetByEntryID returns the file record for a given entry ID.
	GetByEntryID(ctx context.Context, id string) (*models.File, error)

	// GetAllPendingUpload returns files that are staged locally and still need
	// to be uploaded to remote storage (e.g., UploadStatus="pending").
	GetAllPendingUpload(ctx context.Context) ([]*models.File, error)

	// MarkUploaded marks the file for the given entry as uploaded (e.g., set
	// UploadStatus="completed" and clear any temporary local path if desired).
	MarkUploaded(ctx context.Context, id string) error
}
