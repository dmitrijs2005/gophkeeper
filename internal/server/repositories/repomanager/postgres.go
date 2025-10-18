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

type PostgresRepositoryManager struct {
}

func (m *PostgresRepositoryManager) Users(db dbx.DBTX) users.Repository {
	return users.NewPostgresRepository(db)
}

func (m *PostgresRepositoryManager) RefreshTokens(db dbx.DBTX) refreshtokens.Repository {
	return refreshtokens.NewPostgresRepository(db)
}

func (m *PostgresRepositoryManager) Entries(db dbx.DBTX) entries.Repository {
	return entries.NewPostgresRepository(db)
}

func (m *PostgresRepositoryManager) Files(db dbx.DBTX) files.Repository {
	return files.NewPostgresRepository(db)
}

var gooseUpContext = func(ctx context.Context, db *sql.DB, dir string, opts ...goose.OptionsFunc) error {
	return goose.UpContext(ctx, db, dir, opts...)
}

func (m *PostgresRepositoryManager) RunMigrations(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(migrations.Migrations)
	goose.SetDialect("pgx")
	if err := gooseUpContext(ctx, db, "."); err != nil {
		return err
	}
	return nil
}

func NewPostgresRepositoryManager(db *sql.DB) (RepositoryManager, error) {

	m := &PostgresRepositoryManager{}

	return m, nil
}
