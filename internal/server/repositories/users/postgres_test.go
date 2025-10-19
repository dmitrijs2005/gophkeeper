package users

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dmitrijs2005/gophkeeper/internal/common"
	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
)

func newRepoWithMock(t *testing.T) (*PostgresRepository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	return NewPostgresRepository(db), mock, db
}

func TestCreate_Success(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := `(?s)^INSERT\s+INTO\s+users\s*\(username,\s*salt,\s*master_key_verifier\)\s*VALUES\s*\(\$1,\s*\$2,\s*\$3\)\s*RETURNING\s+id\s*$`

	rows := sqlmock.NewRows([]string{"id"}).AddRow("42")
	mock.ExpectQuery(q).
		WithArgs("alice", []byte("salt"), []byte("verifier")).
		WillReturnRows(rows)

	u := &models.User{UserName: "alice", Salt: []byte("salt"), Verifier: []byte("verifier")}
	got, err := repo.Create(context.Background(), u)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if got.ID != "42" || got.UserName != "alice" {
		t.Fatalf("unexpected user: %+v", got)
	}
}

func TestCreate_DBError(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := `(?s)^INSERT\s+INTO\s+users\s*\(username,\s*salt,\s*master_key_verifier\)\s*VALUES\s*\(\$1,\s*\$2,\s*\$3\)\s*RETURNING\s+id\s*$`

	mock.ExpectQuery(q).
		WithArgs("alice", []byte("salt"), []byte("verifier")).
		WillReturnError(errors.New("db down"))

	_, err := repo.Create(context.Background(), &models.User{UserName: "alice", Salt: []byte("salt"), Verifier: []byte("verifier")})
	if err == nil || !regexp.MustCompile(`db error: .*db down`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped db error, got %v", err)
	}
}

func TestGetUserByLogin_Found(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := `(?s)^SELECT\s+ID,\s*username,\s*master_key_verifier,\s*salt\s+FROM\s+users\s+WHERE\s+username\s*=\s*\$1\s*$`

	rows := sqlmock.NewRows([]string{"id", "username", "master_key_verifier", "salt"}).
		AddRow("u-1", "alice", []byte("ver"), []byte("salt"))
	mock.ExpectQuery(q).
		WithArgs("alice").
		WillReturnRows(rows)

	got, err := repo.GetUserByLogin(context.Background(), "alice")
	if err != nil {
		t.Fatalf("GetUserByLogin error: %v", err)
	}
	if got.ID != "u-1" || got.UserName != "alice" {
		t.Fatalf("unexpected user: %+v", got)
	}
}

func TestGetUserByLogin_NotFound(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := `(?s)^SELECT\s+ID,\s*username,\s*master_key_verifier,\s*salt\s+FROM\s+users\s+WHERE\s+username\s*=\s*\$1\s*$`

	mock.ExpectQuery(q).
		WithArgs("ghost").
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetUserByLogin(context.Background(), "ghost")
	if !errors.Is(err, common.ErrorNotFound) {
		t.Fatalf("want common.ErrorNotFound, got %v", err)
	}
}

func TestGetUserByLogin_DBError(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := `(?s)^SELECT\s+ID,\s*username,\s*master_key_verifier,\s*salt\s+FROM\s+users\s+WHERE\s+username\s*=\s*\$1\s*$`

	mock.ExpectQuery(q).
		WithArgs("alice").
		WillReturnError(errors.New("db err"))

	_, err := repo.GetUserByLogin(context.Background(), "alice")
	if err == nil || !regexp.MustCompile(`db error: .*db err`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped db error, got %v", err)
	}
}

func TestIncrementCurrentVersion_Success(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := `(?s)^UPDATE\s+users\s+set\s+current_version\s*=\s*current_version\s*\+\s*1\s+WHERE\s+id\s*=\s*\$1\s+RETURNING\s+current_version\s*$`

	rows := sqlmock.NewRows([]string{"current_version"}).AddRow(int64(7))
	mock.ExpectQuery(q).
		WithArgs("u-1").
		WillReturnRows(rows)

	got, err := repo.IncrementCurrentVersion(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("IncrementCurrentVersion error: %v", err)
	}
	if got != 7 {
		t.Fatalf("unexpected version: %d", got)
	}
}

func TestIncrementCurrentVersion_DBError(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := `(?s)^UPDATE\s+users\s+set\s+current_version\s*=\s*current_version\s*\+\s*1\s+WHERE\s+id\s*=\s*\$1\s+RETURNING\s+current_version\s*$`

	mock.ExpectQuery(q).
		WithArgs("u-1").
		WillReturnError(errors.New("db err"))

	_, err := repo.IncrementCurrentVersion(context.Background(), "u-1")
	if err == nil || !regexp.MustCompile(`db error: .*db err`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped db error, got %v", err)
	}
}
