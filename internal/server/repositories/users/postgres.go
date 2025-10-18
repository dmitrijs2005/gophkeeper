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

type PostgresRepository struct {
	db dbx.DBTX
}

func NewPostgresRepository(db dbx.DBTX) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {

	query :=
		`INSERT INTO users (username, salt, master_key_verifier)
         VALUES ($1, $2, $3)
		 RETURNING id
		 `

	err := r.db.QueryRowContext(ctx, query,
		user.UserName, user.Salt, user.Verifier).Scan(&user.ID)

	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}

	return user, nil
}

func (r *PostgresRepository) GetUserByLogin(ctx context.Context, userName string) (*models.User, error) {
	query :=
		`SELECT ID, username, master_key_verifier, salt FROM users
		 WHERE username = $1
		 `

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, userName).Scan(&user.ID, &user.UserName, &user.Verifier, &user.Salt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrorNotFound
		}
		return nil, fmt.Errorf("db error: %w", err)
	}

	return user, nil
}

func (r *PostgresRepository) IncrementCurrentVersion(ctx context.Context, userID string) (int64, error) {
	query :=
		`UPDATE users set current_version = current_version + 1
		 WHERE id = $1
		 RETURNING current_version
		 `

	var maxVerson int64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&maxVerson)

	if err != nil {
		return 0, fmt.Errorf("db error: %w", err)
	}

	return maxVerson, nil
}
