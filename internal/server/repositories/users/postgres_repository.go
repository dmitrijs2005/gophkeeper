package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
