package repomanager

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/entries"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/files"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/refreshtokens"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/users"
	"github.com/pressly/goose/v3"
)

func newDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	return db, mock
}

func TestNewPostgresRepositoryManager_ReturnsInterface(t *testing.T) {
	db, _ := newDB(t)
	defer db.Close()

	m, err := NewPostgresRepositoryManager(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var _ RepositoryManager = m
}

func TestFactories_ReturnConcreteRepos(t *testing.T) {
	db, _ := newDB(t)
	defer db.Close()

	m := &PostgresRepositoryManager{}

	if u := m.Users(db); u == nil {
		t.Fatal("Users() nil")
	}
	if rt := m.RefreshTokens(db); rt == nil {
		t.Fatal("RefreshTokens() nil")
	}
	if en := m.Entries(db); en == nil {
		t.Fatal("Entries() nil")
	}
	if f := m.Files(db); f == nil {
		t.Fatal("Files() nil")
	}

	var _ users.Repository = m.Users(db)
	var _ refreshtokens.Repository = m.RefreshTokens(db)
	var _ entries.Repository = m.Entries(db)
	var _ files.Repository = m.Files(db)
}

func TestRunMigrations_Success(t *testing.T) {
	db, _ := newDB(t)
	defer db.Close()

	orig := gooseUpContext
	gooseUpContext = func(ctx context.Context, db *sql.DB, dir string, opts ...goose.OptionsFunc) error {
		if dir != "." {
			return errors.New("unexpected dir")
		}
		if len(opts) != 0 {
			return errors.New("unexpected opts")
		}
		return nil
	}
	defer func() { gooseUpContext = orig }()

	m := &PostgresRepositoryManager{}
	if err := m.RunMigrations(context.Background(), db); err != nil {
		t.Fatalf("RunMigrations error: %v", err)
	}
}

func TestRunMigrations_Error(t *testing.T) {
	db, _ := newDB(t)
	defer db.Close()

	orig := gooseUpContext
	gooseUpContext = func(ctx context.Context, db *sql.DB, dir string, opts ...goose.OptionsFunc) error {
		return errors.New("boom")
	}
	defer func() { gooseUpContext = orig }()

	m := &PostgresRepositoryManager{}
	if err := m.RunMigrations(context.Background(), db); err == nil || err.Error() != "boom" {
		t.Fatalf("expected boom, got %v", err)
	}
}
