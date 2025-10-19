// Package refreshtokens provides a PostgreSQL-backed repository for managing
// refresh tokens used in the server's authentication flow.
package refreshtokens

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
)

// PostgresRepository implements CRUD operations for refresh tokens over dbx.DBTX
// (satisfied by *sql.DB or *sql.Tx).
type PostgresRepository struct {
	db dbx.DBTX
}

// NewPostgresRepository constructs a repository bound to the given DBTX.
func NewPostgresRepository(db dbx.DBTX) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Create inserts a new refresh token for userID with an expiry time of now+validity.
func (r *PostgresRepository) Create(ctx context.Context, userID string, token string, validity time.Duration) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`
	if _, err := r.db.ExecContext(ctx, query, userID, token, time.Now().Add(validity)); err != nil {
		return fmt.Errorf("error performing sql request: %v", err)
	}
	return nil
}

// Find returns the refresh token row for the given token string.
// If not found, it returns common.ErrorNotFound.
func (r *PostgresRepository) Find(ctx context.Context, token string) (*models.RefreshToken, error) {
	query := `
		SELECT user_id, expires_at
		FROM refresh_tokens
		WHERE token = $1
	`
	refreshToken := &models.RefreshToken{}
	if err := r.db.QueryRowContext(ctx, query, token).Scan(&refreshToken.UserID, &refreshToken.Expires); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrorNotFound
		}
		return nil, fmt.Errorf("db error: %w", err)
	}
	return refreshToken, nil
}

// Delete removes a refresh token by its token string.
func (r *PostgresRepository) Delete(ctx context.Context, token string) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE token = $1
	`
	if _, err := r.db.ExecContext(ctx, query, token); err != nil {
		return fmt.Errorf("db error: %w", err)
	}
	return nil
}
