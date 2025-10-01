package entries

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Insert(ctx context.Context, entry *models.Entry) error {

	query := ` INSERT INTO entries (id, overview, nonce_overview, details, nonce_details)
			values (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, entry.Id, entry.Overview, entry.NonceOverview, entry.Details, entry.NonceDetails)
	if err != nil {
		return fmt.Errorf("failed to insert entry: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) GetAll(ctx context.Context) ([]models.Entry, error) {

	query := ` select id, overview, nonce_overview from entries where deleted=0`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to select entries: %w", err)
	}

	var result []models.Entry

	defer rows.Close()
	for rows.Next() {
		var item = models.Entry{}
		err := rows.Scan(&item.Id, &item.Overview, &item.NonceOverview)
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

func (r *SQLiteRepository) DeleteByID(ctx context.Context, id string) error {

	query := `update entries set deleted=1 where id=?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

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

func (r *SQLiteRepository) GetByID(ctx context.Context, id string) (*models.Entry, error) {

	query := `select details, nonce_details from entries where deleted=0 and id=?`
	row := r.db.QueryRowContext(ctx, query, id)

	e := &models.Entry{}
	err := row.Scan(&e.Details, &e.NonceDetails)

	if err != nil {
		return nil, fmt.Errorf("wrong rows affected count: %w", err)
	}

	return e, nil

}

func (r *SQLiteRepository) GetAllPending(ctx context.Context) ([]*models.Entry, error) {
	query := `select id, overview, nonce_overview, details, nonce_details from entries where pending=1`
	rows, err := r.db.QueryContext(ctx, query)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	defer rows.Close()

	pendingEntries := []*models.Entry{}

	for rows.Next() {
		var entry = &models.Entry{}
		err := rows.Scan(&entry.Id, &entry.Overview, &entry.NonceOverview, &entry.Details, &entry.NonceDetails)
		if err != nil {
			return nil, err
		}
		pendingEntries = append(pendingEntries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pendingEntries, nil

}
