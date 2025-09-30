package repomanager

import (
	"context"
	"database/sql"

	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/entries"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/refreshtokens"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/users"
	"github.com/dmitrijs2005/gophkeeper/internal/server/shared/tx"
)

type RepositoryManager interface {
	RunMigrations(context.Context, *sql.DB) error
	Users(db tx.DBTX) users.Repository
	RefreshTokens(db tx.DBTX) refreshtokens.Repository
	Entries(db tx.DBTX) entries.Repository
}
