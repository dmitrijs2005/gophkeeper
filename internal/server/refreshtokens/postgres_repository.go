package refreshtokens

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) (*PostgresRepository, error) {
	return &PostgresRepository{db: db}, nil
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
