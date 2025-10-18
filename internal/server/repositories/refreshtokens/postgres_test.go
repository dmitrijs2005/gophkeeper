package refreshtokens

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dmitrijs2005/gophkeeper/internal/common"
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

	q := `(?s)^INSERT\s+INTO\s+refresh_tokens\b.*VALUES\s*\(\$1,\s*\$2,\s*\$3\)\s*$`

	mock.ExpectExec(q).
		WithArgs("u1", "tok123", sqlmock.AnyArg()). // expires_at = time.Now().Add(validity)
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(context.Background(), "u1", "tok123", 30*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreate_DBError(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := `(?s)^INSERT\s+INTO\s+refresh_tokens\b.*VALUES\s*\(\$1,\s*\$2,\s*\$3\)\s*$`

	mock.ExpectExec(q).
		WithArgs("u1", "tok123", sqlmock.AnyArg()).
		WillReturnError(errors.New("db down"))

	err := repo.Create(context.Background(), "u1", "tok123", time.Hour)
	if err == nil || !regexp.MustCompile(`error performing sql request: .*db down`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped db error, got %v", err)
	}
}

func TestFind_Found(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := `(?s)^SELECT\s+user_id,\s*expires_at\s+FROM\s+refresh_tokens\s+WHERE\s+token\s*=\s*\$1\s*$`

	expires := time.Now().Add(10 * time.Minute)
	rows := sqlmock.NewRows([]string{"user_id", "expires_at"}).
		AddRow("u1", expires)

	mock.ExpectQuery(q).
		WithArgs("tok123").
		WillReturnRows(rows)

	got, err := repo.Find(context.Background(), "tok123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.UserID != "u1" || !got.Expires.Equal(expires) {
		t.Fatalf("unexpected row: %+v", got)
	}
}

func TestFind_NotFound(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := `(?s)^SELECT\s+user_id,\s*expires_at\s+FROM\s+refresh_tokens\s+WHERE\s+token\s*=\s*\$1\s*$`

	mock.ExpectQuery(q).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	_, err := repo.Find(context.Background(), "missing")
	if !errors.Is(err, common.ErrorNotFound) {
		t.Fatalf("want common.ErrorNotFound, got %v", err)
	}
}

func TestFind_DBError(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := `(?s)^SELECT\s+user_id,\s*expires_at\s+FROM\s+refresh_tokens\s+WHERE\s+token\s*=\s*\$1\s*$`

	mock.ExpectQuery(q).
		WithArgs("tok123").
		WillReturnError(errors.New("db err"))

	_, err := repo.Find(context.Background(), "tok123")
	if err == nil || !regexp.MustCompile(`db error: .*db err`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped db error, got %v", err)
	}
}

func TestDelete_Success(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := `(?s)^DELETE\s+FROM\s+refresh_tokens\s+WHERE\s+token\s*=\s*\$1\s*$`

	mock.ExpectExec(q).
		WithArgs("tok123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Delete(context.Background(), "tok123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDelete_DBError(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := `(?s)^DELETE\s+FROM\s+refresh_tokens\s+WHERE\s+token\s*=\s*\$1\s*$`

	mock.ExpectExec(q).
		WithArgs("tok123").
		WillReturnError(errors.New("db err"))

	err := repo.Delete(context.Background(), "tok123")
	if err == nil || !regexp.MustCompile(`db error: .*db err`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped db error, got %v", err)
	}
}
