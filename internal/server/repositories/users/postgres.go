// Package users provides PostgreSQL-backed repositories for server-side user
// accounts and related sync metadata.
package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
)

// PostgresRepository implements user persistence over dbx.DBTX (*sql.DB or *sql.Tx).
type PostgresRepository struct {
	db dbx.DBTX
}

// NewPostgresRepository constructs a user repository bound to the given DBTX.
func NewPostgresRepository(db dbx.DBTX) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Create inserts a new user row and returns the populated user (with ID).
func (r *PostgresRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (username, salt, master_key_verifier)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	if err := r.db.QueryRowContext(ctx, query, user.UserName, user.Salt, user.Verifier).Scan(&user.ID); err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}
	return user, nil
}

// GetUserByLogin fetches a user by username. Returns common.ErrorNotFound if missing.
func (r *PostgresRepository) GetUserByLogin(ctx context.Context, userName string) (*models.User, error) {
	query :=
		`SELECT ID, username, master_key_verifier, salt FROM users
		 WHERE username = $1
		 `

	u := &models.User{}
	if err := r.db.QueryRowContext(ctx, query, userName).Scan(&u.ID, &u.UserName, &u.Verifier, &u.Salt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrorNotFound
		}
		return nil, fmt.Errorf("db error: %w", err)
	}
	return u, nil
}

// IncrementCurrentVersion atomically increments and returns the user's current_version.
// This is used to produce a new global version for sync operations.
func (r *PostgresRepository) IncrementCurrentVersion(ctx context.Context, userID string) (int64, error) {
	query :=
		`UPDATE users set current_version = current_version + 1
		 WHERE id = $1
		 RETURNING current_version
		 `

	var maxVersion int64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&maxVersion)

	if err != nil {
		return 0, fmt.Errorf("db error: %w", err)
	}

	return maxVersion, nil
}
