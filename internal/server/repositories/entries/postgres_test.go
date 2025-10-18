package entries

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

func TestCreateOrUpdate_SuccessRowsAffected1(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := regexp.MustCompile(`INSERT INTO entries .* ON CONFLICT .* DO UPDATE SET .* WHERE entries\.user_id = EXCLUDED\.user_id;`)

	mock.ExpectExec(q.String()).
		WithArgs(
			"e1", "u1",
			[]byte("ov"), []byte("no"),
			[]byte("det"), []byte("nd"),
			int64(3),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.CreateOrUpdate(context.Background(), &models.Entry{
		ID:            "e1",
		UserID:        "u1",
		Overview:      []byte("ov"),
		NonceOverview: []byte("no"),
		Details:       []byte("det"),
		NonceDetails:  []byte("nd"),
		Version:       3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateOrUpdate_VersionConflictRowsAffected0(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := regexp.MustCompile(`INSERT INTO entries .* ON CONFLICT .* DO UPDATE SET .* WHERE entries\.user_id = EXCLUDED\.user_id;`)

	mock.ExpectExec(q.String()).
		WithArgs("e1", "u1", []byte("o"), []byte("no"), []byte("d"), []byte("nd"), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.CreateOrUpdate(context.Background(), &models.Entry{
		ID:            "e1",
		UserID:        "u1",
		Overview:      []byte("o"),
		NonceOverview: []byte("no"),
		Details:       []byte("d"),
		NonceDetails:  []byte("nd"),
		Version:       1,
	})
	if !errors.Is(err, common.ErrVersionConflict) {
		t.Fatalf("want ErrVersionConflict, got %v", err)
	}
}

func TestCreateOrUpdate_DBExecError(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := regexp.MustCompile(`INSERT INTO entries .* ON CONFLICT .* DO UPDATE SET .* WHERE entries\.user_id = EXCLUDED\.user_id;`)

	mock.ExpectExec(q.String()).
		WithArgs("e1", "u1", []byte("o"), []byte("no"), []byte("d"), []byte("nd"), int64(1)).
		WillReturnError(errors.New("db is down"))

	err := repo.CreateOrUpdate(context.Background(), &models.Entry{
		ID:            "e1",
		UserID:        "u1",
		Overview:      []byte("o"),
		NonceOverview: []byte("no"),
		Details:       []byte("d"),
		NonceDetails:  []byte("nd"),
		Version:       1,
	})
	if err == nil || !regexp.MustCompile(`db error: .*db is down`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped db error, got %v", err)
	}
}

func TestCreateOrUpdate_RowsAffectedError(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := regexp.MustCompile(`INSERT INTO entries .* ON CONFLICT .* DO UPDATE SET .* WHERE entries\.user_id = EXCLUDED\.user_id;`)

	mock.ExpectExec(q.String()).
		WithArgs("e1", "u1", []byte("o"), []byte("no"), []byte("d"), []byte("nd"), int64(1)).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("rows-err")))

	err := repo.CreateOrUpdate(context.Background(), &models.Entry{
		ID:            "e1",
		UserID:        "u1",
		Overview:      []byte("o"),
		NonceOverview: []byte("no"),
		Details:       []byte("d"),
		NonceDetails:  []byte("nd"),
		Version:       1,
	})
	if err == nil || !regexp.MustCompile(`rows affected error: .*rows-err`).MatchString(err.Error()) {
		t.Fatalf("expected rows affected error, got %v", err)
	}
}

func TestCreateOrUpdate_UnexpectedRowsAffectedGt1(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := regexp.MustCompile(`INSERT INTO entries .* ON CONFLICT .* DO UPDATE SET .* WHERE entries\.user_id = EXCLUDED\.user_id;`)

	mock.ExpectExec(q.String()).
		WithArgs("e1", "u1", []byte("o"), []byte("no"), []byte("d"), []byte("nd"), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.CreateOrUpdate(context.Background(), &models.Entry{
		ID:            "e1",
		UserID:        "u1",
		Overview:      []byte("o"),
		NonceOverview: []byte("no"),
		Details:       []byte("d"),
		NonceDetails:  []byte("nd"),
		Version:       1,
	})
	if err == nil || !regexp.MustCompile(`unexpected rows affected: 2`).MatchString(err.Error()) {
		t.Fatalf("expected unexpected rows affected error, got %v", err)
	}
}

func TestSelectUpdated_Success(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := regexp.MustCompile(`SELECT id, overview, nonce_overview, details, nonce_details, deleted, version from entries\s+WHERE user_id=\$1 and version>\$2`)

	rows := sqlmock.NewRows([]string{
		"id", "overview", "nonce_overview", "details", "nonce_details", "deleted", "version",
	}).AddRow(
		"e1", []byte("ov1"), []byte("no1"), []byte("d1"), []byte("nd1"), false, int64(2),
	).AddRow(
		"e2", []byte("ov2"), []byte("no2"), []byte("d2"), []byte("nd2"), true, int64(5),
	)

	mock.ExpectQuery(q.String()).
		WithArgs("u1", int64(1)).
		WillReturnRows(rows)

	got, err := repo.SelectUpdated(context.Background(), "u1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 rows, got %d", len(got))
	}
	if got[0].ID != "e1" || got[0].Version != 2 || got[0].Deleted {
		t.Fatalf("unexpected first row: %+v", got[0])
	}
	if got[1].ID != "e2" || got[1].Version != 5 || !got[1].Deleted {
		t.Fatalf("unexpected second row: %+v", got[1])
	}
}

func TestSelectUpdated_QueryError(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := regexp.MustCompile(`SELECT id, overview, nonce_overview, details, nonce_details, deleted, version from entries\s+WHERE user_id=\$1 and version>\$2`)

	mock.ExpectQuery(q.String()).
		WithArgs("u1", int64(10)).
		WillReturnError(errors.New("db err"))

	_, err := repo.SelectUpdated(context.Background(), "u1", 10)
	if err == nil || !regexp.MustCompile(`failed to select entries: .*db err`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped select error, got %v", err)
	}
}

func TestSelectUpdated_ScanRowError(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := regexp.MustCompile(`SELECT id, overview, nonce_overview, details, nonce_details, deleted, version from entries\s+WHERE user_id=\$1 and version>\$2`)

	rows := sqlmock.NewRows([]string{
		"id", "overview", "nonce_overview", "details", "nonce_details", "deleted", "version",
	}).AddRow(
		"e1", []byte("ov1"), []byte("no1"), []byte("d1"), []byte("nd1"), "not-bool", int64(2),
	)

	mock.ExpectQuery(q.String()).
		WithArgs("u1", int64(1)).
		WillReturnRows(rows)

	_, err := repo.SelectUpdated(context.Background(), "u1", 1)
	if err == nil {
		t.Fatalf("expected scan error, got nil")
	}
}

func TestSelectUpdated_RowsErr(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := regexp.MustCompile(`SELECT id, overview, nonce_overview, details, nonce_details, deleted, version from entries\s+WHERE user_id=\$1 and version>\$2`)

	rows := sqlmock.NewRows([]string{
		"id", "overview", "nonce_overview", "details", "nonce_details", "deleted", "version",
	}).
		AddRow("e1", []byte("ov1"), []byte("no1"), []byte("d1"), []byte("nd1"), false, int64(2)).
		AddRow("e2", []byte("ov2"), []byte("no2"), []byte("d2"), []byte("nd2"), true, int64(3)).
		RowError(1, errors.New("row-err"))

	mock.ExpectQuery(q.String()).
		WithArgs("u1", int64(1)).
		WillReturnRows(rows)

	_, err := repo.SelectUpdated(context.Background(), "u1", 1)
	if err == nil || err.Error() != "row-err" {
		t.Fatalf("expected rows.Err 'row-err', got %v", err)
	}
}
