package entries

import (
	"context"
	"database/sql"
	"fmt"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) (*PostgresRepository, error) {
	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) Create(ctx context.Context, entry *Entry) (*Entry, error) {

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
