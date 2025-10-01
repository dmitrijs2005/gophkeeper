package repomanager

import (
	"context"
	"database/sql"

	"github.com/dmitrijs2005/gophkeeper/internal/server/migrations"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/entries"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/refreshtokens"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/users"
	"github.com/dmitrijs2005/gophkeeper/internal/server/shared/tx"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type PostgresRepositoryManager struct {
}

func (m *PostgresRepositoryManager) Users(db tx.DBTX) users.Repository {
	return users.NewPostgresRepository(db)
}

func (m *PostgresRepositoryManager) RefreshTokens(db tx.DBTX) refreshtokens.Repository {
	return refreshtokens.NewPostgresRepository(db)
}

func (m *PostgresRepositoryManager) Entries(db tx.DBTX) entries.Repository {
	return entries.NewPostgresRepository(db)
}

func (m *PostgresRepositoryManager) RunMigrations(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(migrations.Migrations)
	goose.SetDialect("pgx")

	if err := goose.UpContext(ctx, db, "."); err != nil {
		return err
	}

	return nil
}

func NewPostgresRepositoryManager(db *sql.DB) (RepositoryManager, error) {

	m := &PostgresRepositoryManager{}

	return m, nil
}
