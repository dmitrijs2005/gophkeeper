// Package repomanager defines an abstraction over concrete repository sets
// used by the server. It centralizes construction of per-boundary repositories
// (users, refresh tokens, entries, files) and exposes a migrations hook.
package repomanager

import (
	"context"
	"database/sql"

	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/entries"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/files"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/refreshtokens"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/users"
)

// RepositoryManager is a factory/registry for server-side repositories.
//
// Implementations typically bind to a specific database (e.g., PostgreSQL),
// provide a RunMigrations method to ensure schema readiness, and vend
// repository instances bound to either *sql.DB or *sql.Tx via the dbx.DBTX
// interface so callers can operate inside or outside transactions.
type RepositoryManager interface {
	// RunMigrations ensures the database schema is up-to-date.
	RunMigrations(context.Context, *sql.DB) error

	// Users returns a users.Repository bound to the provided DBTX.
	Users(db dbx.DBTX) users.Repository
	// RefreshTokens returns a refreshtokens.Repository bound to the provided DBTX.
	RefreshTokens(db dbx.DBTX) refreshtokens.Repository
	// Entries returns an entries.Repository bound to the provided DBTX.
	Entries(db dbx.DBTX) entries.Repository
	// Files returns a files.Repository bound to the provided DBTX.
	Files(db dbx.DBTX) files.Repository
}
