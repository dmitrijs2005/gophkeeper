package client

import (
	"context"
	"database/sql"
	"log"

	"github.com/dmitrijs2005/gophkeeper/internal/client/migrations"
	"github.com/dmitrijs2005/gophkeeper/internal/client/repositories/entries"
	"github.com/dmitrijs2005/gophkeeper/internal/client/repositories/metadata"
	"github.com/pressly/goose/v3"
)

// Repositories groups concrete repository implementations backed by the
// local SQLite database.
type Repositories struct {
	Metadata metadata.Repository
	Entry    entries.Repository
}

// RunMigrations registers the embedded migration files and applies all
// pending goose migrations to the provided database connection.
//
// The function configures goose to use the "sqlite3" dialect to match the
// modernc.org/sqlite driver and executes migrations from the embedded FS.
func RunMigrations(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(migrations.Migrations)

	// Set the database dialect for goose.
	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Fatal("failed to set goose dialect:", err)
	}

	return goose.UpContext(ctx, db, ".")
}

// InitDatabase opens (or creates) an SQLite database at the given DSN and
// runs schema migrations. On success it returns the ready-to-use *sql.DB.
func InitDatabase(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := RunMigrations(ctx, db); err != nil {
		return nil, err
	}
	return db, nil
}
