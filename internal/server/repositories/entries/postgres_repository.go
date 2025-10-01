package entries

import (
	"context"
	"fmt"

	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
	"github.com/dmitrijs2005/gophkeeper/internal/server/shared/tx"
)

type PostgresRepository struct {
	db tx.DBTX
}

func NewPostgresRepository(db tx.DBTX) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, entry *models.Entry) (*models.Entry, error) {

	query :=
		`INSERT INTO entries (id, user_id, overview, nonce_overview, details, nonce_details)
		VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id
		 `

	err := r.db.QueryRowContext(ctx, query, entry.ID, entry.UserID, entry.Overview, entry.NonceOverview, entry.Details, entry.NonceDetails).Scan(&entry.ID)

	if err != nil {
		return nil, fmt.Errorf("error performing sql request: %v", err)
	}

	return entry, nil
}
