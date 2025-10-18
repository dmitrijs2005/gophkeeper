package files

import (
	"context"
	"fmt"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
)

type SQLiteRepository struct {
	db dbx.DBTX
}

func NewSQLiteRepository(db dbx.DBTX) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

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
	_, err := r.db.ExecContext(ctx, query, e.EntryID, e.EncryptedFileKey, e.Nonce, e.LocalPath, e.UploadStatus, e.Deleted)
	if err != nil {
		return fmt.Errorf("failed to upsert file: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) DeleteByEntryID(ctx context.Context, id string) error {

	query := `update files set deleted=1 where entry_id=? and deleted=0`
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

	query := `select entry_id, encrypted_file_key, nonce, local_path, upload_status, deleted from files where entry_id=?`
	row := r.db.QueryRowContext(ctx, query, id)

	e := &models.File{}
	err := row.Scan(&e.EntryID, &e.EncryptedFileKey, &e.Nonce, &e.LocalPath, &e.UploadStatus, &e.Deleted)

	if err != nil {
		return nil, fmt.Errorf("wrong rows affected count: %w", err)
	}

	return e, nil

}

func (r *SQLiteRepository) GetAllPendingUpload(ctx context.Context) ([]*models.File, error) {

	query := `select entry_id, encrypted_file_key, nonce from files where upload_status='pending'`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error selecting files: %w", err)
	}
	defer rows.Close()

	var result []*models.File

	for rows.Next() {
		var item = &models.File{}
		err := rows.Scan(&item.EntryID, &item.EncryptedFileKey, &item.Nonce)
		if err != nil {
			return nil, err
		}
		result = append(result, item)

	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil

}

func (r *SQLiteRepository) MarkUploaded(ctx context.Context, id string) error {

	query := `update files set upload_status='completed' where entry_id=?`
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
