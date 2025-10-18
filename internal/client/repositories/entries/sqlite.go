package entries

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
)

// SQLiteRepository implements Repository using a DBTX (either *sql.DB or *sql.Tx).
type SQLiteRepository struct {
	db dbx.DBTX
}

// NewSQLiteRepository returns a new SQLiteRepository bound to the given DBTX.
func NewSQLiteRepository(db dbx.DBTX) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// CreateOrUpdate upserts an entry by id. On conflict, selected columns are updated.
func (r *SQLiteRepository) CreateOrUpdate(ctx context.Context, e *models.Entry) error {
	query := ` INSERT INTO entries (id, overview, nonce_overview, details, nonce_details, deleted)
			values (?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET overview = excluded.overview, 
				nonce_overview = excluded.nonce_overview, 
				details = excluded.details, 
				nonce_details = excluded.nonce_details,
				deleted = excluded.deleted
	`
	_, err := r.db.ExecContext(ctx, query,
		e.Id, e.Overview, e.NonceOverview, e.Details, e.NonceDetails, e.Deleted)
	if err != nil {
		return fmt.Errorf("failed to upsert entry: %w", err)
	}
	return nil
}

// GetAll lists all non-deleted entries, returning only overview fields.
func (r *SQLiteRepository) GetAll(ctx context.Context) ([]models.Entry, error) {
	query := `select id, overview, nonce_overview from entries where deleted=0`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to select entries: %w", err)
	}
	defer rows.Close()

	var result []models.Entry
	for rows.Next() {
		var item models.Entry
		if err := rows.Scan(&item.Id, &item.Overview, &item.NonceOverview); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteByID marks an entry as deleted (soft delete). It expects exactly one row to be affected.
func (r *SQLiteRepository) DeleteByID(ctx context.Context, id string) error {
	query := `update entries set deleted=1 where id=? and deleted=0`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if ra != 1 {
		return fmt.Errorf("wrong rows affected count: %d", ra)
	}
	return nil
}

// GetByID returns details for a single non-deleted entry.
func (r *SQLiteRepository) GetByID(ctx context.Context, id string) (*models.Entry, error) {
	query := `select details, nonce_details from entries where deleted=0 and id=?`
	row := r.db.QueryRowContext(ctx, query, id)

	e := &models.Entry{}
	if err := row.Scan(&e.Details, &e.NonceDetails); err != nil {
		return nil, fmt.Errorf("query row scan failed: %w", err)
	}
	return e, nil
}

// GetAllPending returns entries flagged as pending=1 (awaiting sync).
func (r *SQLiteRepository) GetAllPending(ctx context.Context) ([]*models.Entry, error) {
	query := `select id, overview, nonce_overview, details, nonce_details from entries where pending=1`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	defer rows.Close()

	var pending []*models.Entry
	for rows.Next() {
		entry := &models.Entry{}
		if err := rows.Scan(&entry.Id, &entry.Overview, &entry.NonceOverview, &entry.Details, &entry.NonceDetails); err != nil {
			return nil, err
		}
		pending = append(pending, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return pending, nil
}
