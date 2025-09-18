package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dmitrijs2005/gophkeeper/internal/shared"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) (*PostgresRepository, error) {
	return &PostgresRepository{db: db}, nil
}

// func (r *PostgresRepository) FindUserByLogin(ctx context.Context, login string) (models.User, error) {

// 	s := "select id, login, password, salt from users where login=$1"

// 	var user models.User

// 	_, err := common.RetryWithResult(ctx, func() (*sql.Row, error) {
// 		r := r.db.QueryRowContext(ctx, s, login)
// 		err := r.Scan(&user.ID, &user.Login, &user.Password, &user.Salt)

// 		if err != nil {
// 			if errors.Is(err, sql.ErrNoRows) {
// 				return nil, common.ErrorNotFound
// 			}
// 			return nil, err
// 		}

// 		return r, nil
// 	})

// 	return user, err
// }

func (r *PostgresRepository) Create(ctx context.Context, user *User) (*User, error) {

	query :=
		`INSERT INTO users (username, salt, master_key_verifier)
         VALUES ($1, $2, $3)
		 RETURNING id
		 `

	err := r.db.QueryRowContext(ctx, query,
		user.UserName, user.Salt, user.Verifier).Scan(&user.ID)

	if err != nil {
		return nil, fmt.Errorf("error performing sql request: %v", err)
	}

	return user, nil
}

func (r *PostgresRepository) GetUserByLogin(ctx context.Context, userName string) (*User, error) {
	query :=
		`SELECT ID, username, master_key_verifier, salt FROM users
		 WHERE username = $1
		 `

	user := &User{}
	err := r.db.QueryRowContext(ctx, query, userName).Scan(&user.ID, &user.UserName, &user.Verifier, &user.Salt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, shared.ErrorNotFound
		}
		return nil, fmt.Errorf("error performing sql request: %v", err)
	}

	return user, nil
}
