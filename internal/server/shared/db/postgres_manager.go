package db

import (
	"context"
	"database/sql"

	migrations "github.com/dmitrijs2005/gophkeeper/internal/migrations/server"
	"github.com/dmitrijs2005/gophkeeper/internal/server/entries"
	"github.com/dmitrijs2005/gophkeeper/internal/server/refreshtokens"
	"github.com/dmitrijs2005/gophkeeper/internal/server/users"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type PostgresRepositoryManager struct {
	db            *sql.DB
	users         users.Repository
	refreshTokens refreshtokens.Repository
	entries       entries.Repository
}

func (m *PostgresRepositoryManager) Conn() *sql.DB {
	return m.db
}

func (m *PostgresRepositoryManager) Users() users.Repository {
	return m.users
}

func (m *PostgresRepositoryManager) RefreshTokens() refreshtokens.Repository {
	return m.refreshTokens
}

func (m *PostgresRepositoryManager) Entries() entries.Repository {
	return m.entries
}

func (m *PostgresRepositoryManager) RunMigrations(ctx context.Context) error {
	goose.SetBaseFS(migrations.Migrations) // Вот здесь передаём embed FS!

	if err := goose.UpContext(ctx, m.db, "."); err != nil {
		return err
	}

	return nil
}

func NewPostgresRepositoryManager(dsn string) (RepositoryManager, error) {

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	users, err := users.NewPostgresRepository(db)
	if err != nil {
		return nil, err
	}

	refreshTokens, err := refreshtokens.NewPostgresRepository(db)
	if err != nil {
		return nil, err
	}

	entries, err := entries.NewPostgresRepository(db)
	if err != nil {
		return nil, err
	}

	m := &PostgresRepositoryManager{
		db:            db,
		users:         users,
		refreshTokens: refreshTokens,
		entries:       entries,
	}

	err = m.RunMigrations(context.Background())
	if err != nil {
		return nil, err
	}

	return m, nil
}
