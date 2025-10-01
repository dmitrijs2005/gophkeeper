package refreshtokens

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
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

func (r *PostgresRepository) Find(ctx context.Context, token string) (*models.RefreshToken, error) {
	query :=
		`SELECT user_id, expires_at
			FROM refresh_tokens
		 WHERE token = $1
		 `
	refreshToken := &models.RefreshToken{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(&refreshToken.UserID, &refreshToken.Expires)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrorNotFound
		}
		return nil, fmt.Errorf("db error: %w", err)
	}

	return refreshToken, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, token string) error {
	query :=
		`DELETE FROM refresh_tokens
		 WHERE token = $1
		 `
	_, err := r.db.ExecContext(ctx, query, token)

	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	return nil
}
