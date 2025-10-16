package files

import (
	"context"
	"fmt"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
)

type PostgresRepository struct {
	db dbx.DBTX
}

func NewPostgresRepository(db dbx.DBTX) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateOrUpdate(ctx context.Context, file *models.File) error {

	query :=
		`INSERT INTO files (entry_id, user_id, version, encrypted_file_key, nonce, upload_status, storage_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (entry_id)
		DO UPDATE SET 
			user_id = EXCLUDED.user_id, 
			version = EXCLUDED.version,
			encrypted_file_key = EXCLUDED.encrypted_file_key, 
			nonce = EXCLUDED.nonce, 
			upload_status = EXCLUDED.upload_status 
			WHERE files.entry_id = EXCLUDED.entry_id;
		 `

	res, err := r.db.ExecContext(ctx, query, file.EntryID, file.UserID, file.Version, file.EncryptedFileKey, file.Nonce, file.UploadStatus, file.StorageKey)
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected error: %w", err)
	}

	switch n {
	case 1:
		return nil
	case 0:
		return common.ErrVersionConflict
	default:
		return fmt.Errorf("unexpected rows affected: %d", n)
	}

}

func (r *PostgresRepository) SelectUpdated(ctx context.Context, userID string, minVersion int64) ([]*models.File, error) {
	query := ` SELECT entry_id, user_id, version, encrypted_file_key, nonce, upload_status from files 
		WHERE user_id=$1 and version>$2
		`
	rows, err := r.db.QueryContext(ctx, query, userID, minVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to select files: %w", err)
	}

	var result []*models.File

	defer rows.Close()
	for rows.Next() {
		var item = models.File{}
		err := rows.Scan(&item.EntryID, &item.UserID, &item.Version, &item.EncryptedFileKey, &item.Nonce, &item.UploadStatus)
		if err != nil {
			return nil, err
		}
		result = append(result, &item)

	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *PostgresRepository) MarkUploaded(ctx context.Context, id string) error {

	query := `update files set upload_status='completed' where entry_id=$1`
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

func (r *PostgresRepository) GetByEntryID(ctx context.Context, id string) (*models.File, error) {
	query := ` SELECT entry_id, user_id, storage_key from files 
		WHERE entry_id=$1
		`

	result := &models.File{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&result.EntryID, &result.UserID, &result.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to select files: %w", err)
	}

	return result, nil
}
