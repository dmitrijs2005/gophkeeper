// Package metadata provides a SQLite-backed implementation of the client-side
// metadata.Repository — a simple key–value store for auxiliary data such as
// sync cursors and feature flags.
package metadata

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
)

// SQLiteRepository implements Repository over a dbx.DBTX (*sql.DB or *sql.Tx).
type SQLiteRepository struct {
	db dbx.DBTX
}

// NewSQLiteRepository constructs a metadata repository bound to the given DBTX.
func NewSQLiteRepository(db dbx.DBTX) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// Get returns the value stored under key. If the key does not exist,
// it returns (nil, nil).
func (r *SQLiteRepository) Get(ctx context.Context, key string) ([]byte, error) {
	var value []byte
	err := r.db.QueryRowContext(ctx, `SELECT value FROM metadata WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata[%s]: %w", key, err)
	}
	return value, nil
}

// Set creates or replaces the value stored under key.
func (r *SQLiteRepository) Set(ctx context.Context, key string, value []byte) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO metadata (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	if err != nil {
		return fmt.Errorf("failed to set metadata[%s]: %w", key, err)
	}
	return nil
}

// Delete removes the given key from the store. Deleting a non-existent key
// is treated as success.
func (r *SQLiteRepository) Delete(ctx context.Context, key string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM metadata WHERE key = ?`, key)
	if err != nil {
		return fmt.Errorf("failed to delete metadata[%s]: %w", key, err)
	}
	return nil
}

// Clear removes all key–value pairs from the store.
func (r *SQLiteRepository) Clear(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM metadata`)
	if err != nil {
		return fmt.Errorf("failed to clear metadata: %w", err)
	}
	return nil
}

// List returns all metadata as a map keyed by string. The order of iteration
// over the returned map is undefined.
func (r *SQLiteRepository) List(ctx context.Context) (map[string][]byte, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT key, value FROM metadata`)
	if err != nil {
		return nil, fmt.Errorf("failed to list metadata: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]byte)
	for rows.Next() {
		var key string
		var value []byte
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan metadata row: %w", err)
		}
		result[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate metadata rows: %w", err)
	}
	return result, nil
}
