package metadata

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
)

type SQLiteRepository struct {
	db dbx.DBTX
}

func NewSQLiteRepository(db dbx.DBTX) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

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

func (r *SQLiteRepository) Delete(ctx context.Context, key string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM metadata WHERE key = ?`, key)
	if err != nil {
		return fmt.Errorf("failed to delete metadata[%s]: %w", key, err)
	}
	return nil
}

func (r *SQLiteRepository) Clear(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM metadata`)
	if err != nil {
		return fmt.Errorf("failed to clear metadata: %w", err)
	}
	return nil
}

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
