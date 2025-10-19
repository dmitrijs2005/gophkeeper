package metadata

import "context"

// Repository describes CRUD-style operations over a namespaced key–value store.
// Implementations are typically backed by a local database.
type Repository interface {
	// Get returns the raw value for key. If the key is missing, the
	// implementation may return (nil, nil) or a not-found error.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores value under key, creating or replacing the existing value.
	Set(ctx context.Context, key string, value []byte) error

	// Delete removes the value for key. Deleting a non-existent key is not an error.
	Delete(ctx context.Context, key string) error

	// List returns all key–value pairs. Order is implementation-defined.
	List(ctx context.Context) (map[string][]byte, error)

	// Clear removes all keys from the store.
	Clear(ctx context.Context) error
}
