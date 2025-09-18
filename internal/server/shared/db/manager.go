package db

import (
	"context"
	"database/sql"

	"github.com/dmitrijs2005/gophkeeper/internal/server/entries"
	"github.com/dmitrijs2005/gophkeeper/internal/server/refreshtokens"
	"github.com/dmitrijs2005/gophkeeper/internal/server/users"
)

type RepositoryManager interface {
	RunMigrations(context.Context) error
	Conn() *sql.DB
	Users() users.Repository
	RefreshTokens() refreshtokens.Repository
	Entries() entries.Repository
}
