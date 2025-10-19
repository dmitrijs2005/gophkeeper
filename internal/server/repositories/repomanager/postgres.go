// Package repomanager provides a concrete RepositoryManager for PostgreSQL,
// wiring together repository constructors and database migrations (via goose).
package repomanager

import (
	"context"
	"database/sql"

	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
	"github.com/dmitrijs2005/gophkeeper/internal/server/migrations"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/entries"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/files"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/refreshtokens"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/users"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// PostgresRepositoryManager vends PostgreSQL-backed repository implementations
// and exposes a schema migration hook.
type PostgresRepositoryManager struct{}

// Users returns a users.Repository bound to the provided DBTX.
func (m *PostgresRepositoryManager) Users(db dbx.DBTX) users.Repository {
	return users.NewPostgresRepository(db)
}

// RefreshTokens returns a refreshtokens.Repository bound to the provided DBTX.
func (m *PostgresRepositoryManager) RefreshTokens(db dbx.DBTX) refreshtokens.Repository {
	return refreshtokens.NewPostgresRepository(db)
}

// Entries returns an entries.Repository bound to the provided DBTX.
func (m *PostgresRepositoryManager) Entries(db dbx.DBTX) entries.Repository {
	return entries.NewPostgresRepository(db)
}

// Files returns a files.Repository bound to the provided DBTX.
func (m *PostgresRepositoryManager) Files(db dbx.DBTX) files.Repository {
	return files.NewPostgresRepository(db)
}

// gooseUpContext is a seam for testing goose.UpContext.
var gooseUpContext = func(ctx context.Context, db *sql.DB, dir string, opts ...goose.OptionsFunc) error {
	return goose.UpContext(ctx, db, dir, opts...)
}

// RunMigrations sets up goose with the embedded migrations and runs them
// against the provided database connection.
func (m *PostgresRepositoryManager) RunMigrations(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(migrations.Migrations)
	goose.SetDialect("pgx")
	if err := gooseUpContext(ctx, db, "."); err != nil {
		return err
	}
	return nil
}

// NewPostgresRepositoryManager constructs a PostgreSQL-backed RepositoryManager.
func NewPostgresRepositoryManager(db *sql.DB) (RepositoryManager, error) {
	return &PostgresRepositoryManager{}, nil
}
