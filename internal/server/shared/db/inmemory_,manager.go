package db

import (
	"context"
	"database/sql"

	"github.com/dmitrijs2005/gophkeeper/internal/server/entries"
	"github.com/dmitrijs2005/gophkeeper/internal/server/refreshtokens"
	"github.com/dmitrijs2005/gophkeeper/internal/server/users"
)

type InMemoryRepositoryManager struct {
	users         users.Repository
	refreshTokens refreshtokens.Repository
	entries       entries.Repository
}

func (m InMemoryRepositoryManager) Conn() *sql.DB {
	return nil
}

func (m InMemoryRepositoryManager) RunMigrations(ctx context.Context) error {
	return nil
}

func (m InMemoryRepositoryManager) Users() users.Repository {
	return m.users
}

func (m InMemoryRepositoryManager) RefreshTokens() refreshtokens.Repository {
	return m.refreshTokens
}

func (m InMemoryRepositoryManager) Entries() entries.Repository {
	return m.entries
}

func NewInMemoryRepositoryManager() RepositoryManager {
	return InMemoryRepositoryManager{}
}
