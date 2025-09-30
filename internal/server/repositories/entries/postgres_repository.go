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
		`INSERT INTO entries (user_id, title, type, encrypted_data, nonce)
		VALUES ($1, $2, $3, $4, $5)
		 RETURNING id
		 `

	err := r.db.QueryRowContext(ctx, query, entry.UserID, entry.Title, entry.Type, entry.EncryptedData, entry.Nonce).Scan(&entry.ID)

	if err != nil {
		return nil, fmt.Errorf("error performing sql request: %v", err)
	}

	return entry, nil
}
