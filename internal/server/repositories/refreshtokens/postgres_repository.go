package refreshtokens

import (
	"context"
	"fmt"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/server/shared/tx"
)

type PostgresRepository struct {
	db tx.DBTX
}

func NewPostgresRepository(db tx.DBTX) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, userID string, token string, validity time.Duration) error {

	query :=
		`INSERT INTO refresh_tokens (user_id, token, expires_at)
         VALUES ($1, $2, $3)
		 `

	_, err := r.db.ExecContext(ctx, query, userID, token, time.Now().Add(validity))

	if err != nil {
		return fmt.Errorf("error performing sql request: %v", err)
	}

	return nil
}
