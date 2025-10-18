package entries

import (
	"context"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
)

// Repository describes CRUD and query operations for Entry objects.
// Implementations are typically backed by a local SQLite database.
type Repository interface {
	// CreateOrUpdate inserts a new entry or updates an existing one by Id.
	CreateOrUpdate(ctx context.Context, entry *models.Entry) error

	// GetAll returns all entries, including deleted ones if the implementation
	// uses tombstones for synchronization.
	GetAll(ctx context.Context) ([]models.Entry, error)

	// DeleteByID marks an entry as deleted or removes it (implementation-defined).
	DeleteByID(ctx context.Context, id string) error

	// GetByID returns an entry by its identifier.
	GetByID(ctx context.Context, id string) (*models.Entry, error)

	// GetAllPending returns entries that have local changes not yet synchronized
	// with the server (e.g., new/updated/deleted since last sync).
	GetAllPending(ctx context.Context) ([]*models.Entry, error)
}
