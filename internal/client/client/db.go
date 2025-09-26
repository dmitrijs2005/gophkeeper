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

type Repositories struct {
	Metadata metadata.Repository
	Entry    entries.Repository
	//Vault    vault.VaultRepository
	//DB       *sql.DB
}

func RunMigrations(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(migrations.Migrations)

	// Set the database dialect
	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Fatal("failed to set goose dialect:", err)
	}

	return goose.UpContext(ctx, db, ".")
}

func InitDatabase(ctx context.Context, dsn string) (*Repositories, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	if err := RunMigrations(ctx, db); err != nil {
		return nil, err
	}

	repos := &Repositories{
		Metadata: metadata.NewSQLiteMetadataRepository(db),
		Entry:    entries.NewSQLiteRepository(db),
		//Vault:    vault.NewSQLiteVaultRepository(db),
		//DB:       db,
	}
	return repos, nil
}
