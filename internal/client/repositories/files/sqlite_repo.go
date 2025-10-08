package files

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) CreateOrUpdate(ctx context.Context, e *models.File) error {

	query := ` INSERT INTO files (id, entry_id, encrypted_file_key, nonce, local_path, upload_status, deleted)
			values (?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET entry_id = excluded.entry_id, 
				encrypted_file_key = excluded.encrypted_file_key, 
				nonce = excluded.nonce, 
				local_path = excluded.local_path,
				upload_status = excluded.upload_status
				deleted = excluded.deleted
	`
	_, err := r.db.ExecContext(ctx, query, e.ID, e.EntryID, e.EncryptedFileKey, e.Nonce, e.LocalPath, e.UploadStatus, e.Deleted)
	if err != nil {
		return fmt.Errorf("failed to upsert entry: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) DeleteByEntryID(ctx context.Context, id string) error {

	query := `update files set deleted=1 where entry_id=?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected != 1 {
		return fmt.Errorf("wrong rows affected count: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) GetByEntryID(ctx context.Context, id string) (*models.File, error) {

	query := `select id, entry_id, encrypted_file_key, nonce, local_path, upload_status, deleted from files where entry_id=?`
	row := r.db.QueryRowContext(ctx, query, id)

	e := &models.File{}
	err := row.Scan(&e.ID, &e.EntryID, e.EncryptedFileKey, e.Nonce, e.LocalPath, e.UploadStatus, e.Deleted)

	if err != nil {
		return nil, fmt.Errorf("wrong rows affected count: %w", err)
	}

	return e, nil

}
