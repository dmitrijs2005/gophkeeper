// Package entries provides PostgreSQL-backed repositories for server-side
// entry persistence and sync queries.
package entries

import (
	"context"
	"fmt"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
)

// PostgresRepository implements entry storage over a dbx.DBTX (*sql.DB or *sql.Tx).
type PostgresRepository struct {
	db dbx.DBTX
}

// NewPostgresRepository constructs a repository bound to the given DBTX.
func NewPostgresRepository(db dbx.DBTX) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreateOrUpdate upserts an entry by ID for a specific user. If a conflicting
// row exists for another user, no row is updated and ErrVersionConflict is returned.
// Returns an error for DB failures or unexpected rows affected.
func (r *PostgresRepository) CreateOrUpdate(ctx context.Context, entry *models.Entry) error {
	query := `
		INSERT INTO entries (id, user_id, overview, nonce_overview, details, nonce_details, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id)
		DO UPDATE SET 
			overview = EXCLUDED.overview, 
			nonce_overview = EXCLUDED.nonce_overview, 
			details = EXCLUDED.details, 
			nonce_details = EXCLUDED.nonce_details, 
			version = EXCLUDED.version
			WHERE entries.user_id = EXCLUDED.user_id;
	`
	res, err := r.db.ExecContext(ctx, query,
		entry.ID, entry.UserID, entry.Overview, entry.NonceOverview, entry.Details, entry.NonceDetails, entry.Version)
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

// SelectUpdated returns all entries for userID with version > minVersion.
// The result contains encrypted fields and sync metadata.
func (r *PostgresRepository) SelectUpdated(ctx context.Context, userID string, minVersion int64) ([]*models.Entry, error) {
	query := ` SELECT id, overview, nonce_overview, details, nonce_details, deleted, version from entries 
		WHERE user_id=$1 and version>$2
		`
	rows, err := r.db.QueryContext(ctx, query, userID, minVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to select entries: %w", err)
	}
	defer rows.Close()

	var result []*models.Entry
	for rows.Next() {
		var item models.Entry
		if err := rows.Scan(
			&item.ID, &item.Overview, &item.NonceOverview, &item.Details, &item.NonceDetails,
			&item.Deleted, &item.Version,
		); err != nil {
			return nil, err
		}
		result = append(result, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
