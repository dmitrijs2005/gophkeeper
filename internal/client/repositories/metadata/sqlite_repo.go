package metadata

import (
	"context"
	"database/sql"
	"fmt"
)

type SQLiteMetadataRepository struct {
	db *sql.DB
}

func NewSQLiteMetadataRepository(db *sql.DB) *SQLiteMetadataRepository {
	return &SQLiteMetadataRepository{db: db}
}

func (r *SQLiteMetadataRepository) Get(ctx context.Context, key string) ([]byte, error) {
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

func (r *SQLiteMetadataRepository) Set(ctx context.Context, key string, value []byte) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO metadata (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	if err != nil {
		return fmt.Errorf("failed to set metadata[%s]: %w", key, err)
	}
	return nil
}

func (r *SQLiteMetadataRepository) Delete(ctx context.Context, key string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM metadata WHERE key = ?`, key)
	if err != nil {
		return fmt.Errorf("failed to delete metadata[%s]: %w", key, err)
	}
	return nil
}

func (r *SQLiteMetadataRepository) List(ctx context.Context) (map[string][]byte, error) {
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
