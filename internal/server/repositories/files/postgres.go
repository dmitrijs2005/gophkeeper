package files

import (
	"context"
	"fmt"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
)

// PostgresRepository implements file storage over a dbx.DBTX (*sql.DB or *sql.Tx).
type PostgresRepository struct {
	db dbx.DBTX
}

// NewPostgresRepository constructs a repository bound to the given DBTX.
func NewPostgresRepository(db dbx.DBTX) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreateOrUpdate upserts a file record by entry_id. On conflict, server-side
// fields are updated. Returns ErrVersionConflict when no row is affected
// due to a version or ownership constraint.
func (r *PostgresRepository) CreateOrUpdate(ctx context.Context, file *models.File) error {
	query := `
		INSERT INTO files (entry_id, user_id, version, encrypted_file_key, nonce, upload_status, storage_key)
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
	res, err := r.db.ExecContext(ctx, query,
		file.EntryID, file.UserID, file.Version, file.EncryptedFileKey, file.Nonce, file.UploadStatus, file.StorageKey)
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

// SelectUpdated returns all files for userID with version > minVersion.
func (r *PostgresRepository) SelectUpdated(ctx context.Context, userID string, minVersion int64) ([]*models.File, error) {
	query := ` SELECT entry_id, user_id, version, encrypted_file_key, nonce, upload_status from files 
		WHERE user_id=$1 and version>$2
		`
	rows, err := r.db.QueryContext(ctx, query, userID, minVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to select files: %w", err)
	}
	defer rows.Close()

	var result []*models.File
	for rows.Next() {
		var item models.File
		if err := rows.Scan(&item.EntryID, &item.UserID, &item.Version, &item.EncryptedFileKey, &item.Nonce, &item.UploadStatus); err != nil {
			return nil, err
		}
		result = append(result, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// MarkUploaded marks the file for entry id as uploaded (upload_status='completed').
// Exactly one row must be affected.
func (r *PostgresRepository) MarkUploaded(ctx context.Context, id string) error {
	query := `update files set upload_status='completed' where entry_id=$1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark uploaded: %w", err)
	}
	ra, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if ra != 1 {
		return fmt.Errorf("wrong rows affected count: %d", ra)
	}
	return nil
}

// GetByEntryID returns a minimal file row (entry_id, user_id, storage_key)
// used to authorize and build presigned URLs.
func (r *PostgresRepository) GetByEntryID(ctx context.Context, id string) (*models.File, error) {
	query := ` SELECT entry_id, user_id, storage_key from files 
		WHERE entry_id=$1
		`

	result := &models.File{}
	if err := r.db.QueryRowContext(ctx, query, id).Scan(&result.EntryID, &result.UserID, &result.StorageKey); err != nil {
		return nil, fmt.Errorf("failed to select files: %w", err)
	}
	return result, nil
}
