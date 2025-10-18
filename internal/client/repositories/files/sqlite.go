// Package files provides a SQLite-backed implementation of the client-side
// files.Repository for persisting per-entry file metadata (keys, nonces,
// local staging path, upload status, soft-delete flag).
package files

import (
	"context"
	"fmt"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
)

// SQLiteRepository implements Repository over a dbx.DBTX (*sql.DB or *sql.Tx).
type SQLiteRepository struct {
	db dbx.DBTX
}

// NewSQLiteRepository constructs a repository bound to the given DBTX.
func NewSQLiteRepository(db dbx.DBTX) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// CreateOrUpdate upserts a file record by entry_id.
// On conflict, all tracked columns are updated.
func (r *SQLiteRepository) CreateOrUpdate(ctx context.Context, e *models.File) error {
	query := ` INSERT INTO files (entry_id, encrypted_file_key, nonce, local_path, upload_status, deleted)
			values (?, ?, ?, ?, ?, ?)
			ON CONFLICT(entry_id) DO UPDATE SET entry_id = excluded.entry_id, 
				encrypted_file_key = excluded.encrypted_file_key, 
				nonce = excluded.nonce, 
				local_path = excluded.local_path,
				upload_status = excluded.upload_status,
				deleted = excluded.deleted
	`
	if _, err := r.db.ExecContext(ctx, query, e.EntryID, e.EncryptedFileKey, e.Nonce, e.LocalPath, e.UploadStatus, e.Deleted); err != nil {
		return fmt.Errorf("failed to upsert file: %w", err)
	}
	return nil
}

// DeleteByEntryID performs a soft delete for the file record linked to id.
// Exactly one row must be affected.
func (r *SQLiteRepository) DeleteByEntryID(ctx context.Context, id string) error {
	query := `update files set deleted=1 where entry_id=? and deleted=0`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if ra != 1 {
		return fmt.Errorf("unexpected rows affected: %d", ra)
	}
	return nil
}

// GetByEntryID returns a file record for the given entry id.
func (r *SQLiteRepository) GetByEntryID(ctx context.Context, id string) (*models.File, error) {
	query := `select entry_id, encrypted_file_key, nonce, local_path, upload_status, deleted from files where entry_id=?`
	row := r.db.QueryRowContext(ctx, query, id)

	e := &models.File{}
	if err := row.Scan(&e.EntryID, &e.EncryptedFileKey, &e.Nonce, &e.LocalPath, &e.UploadStatus, &e.Deleted); err != nil {
		return nil, fmt.Errorf("query row scan failed: %w", err)
	}
	return e, nil
}

// GetAllPendingUpload returns files whose upload_status indicates a pending upload.
func (r *SQLiteRepository) GetAllPendingUpload(ctx context.Context) ([]*models.File, error) {
	query := `select entry_id, encrypted_file_key, nonce from files where upload_status='pending'`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error selecting files: %w", err)
	}
	defer rows.Close()

	var result []*models.File
	for rows.Next() {
		item := &models.File{}
		if err := rows.Scan(&item.EntryID, &item.EncryptedFileKey, &item.Nonce); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// MarkUploaded sets upload_status='completed' for the file of the given entry id.
// Exactly one row must be affected.
func (r *SQLiteRepository) MarkUploaded(ctx context.Context, id string) error {
	query := `update files set upload_status='completed' where entry_id=?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark uploaded: %w", err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if ra != 1 {
		return fmt.Errorf("unexpected rows affected: %d", ra)
	}
	return nil
}
