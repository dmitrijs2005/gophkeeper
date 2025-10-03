package repomanager

import (
	"context"
	"database/sql"

	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/entries"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/refreshtokens"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/users"
)

type RepositoryManager interface {
	RunMigrations(context.Context, *sql.DB) error
	Users(db dbx.DBTX) users.Repository
	RefreshTokens(db dbx.DBTX) refreshtokens.Repository
	Entries(db dbx.DBTX) entries.Repository
}
