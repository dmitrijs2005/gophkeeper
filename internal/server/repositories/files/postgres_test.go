package files

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

func TestCreateOrUpdate_Success(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := `(?s)^INSERT\s+INTO\s+files\b.*ON\s+CONFLICT\s*\(entry_id\)\s*DO\s+UPDATE\s+SET\b.*WHERE\s+files\.entry_id\s*=\s*EXCLUDED\.entry_id;?$`

	mock.ExpectExec(q).
		WithArgs("e1", "u1", int64(3), []byte("fk"), []byte("n"), "pending", "skey").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.CreateOrUpdate(context.Background(), &models.File{
		EntryID:          "e1",
		UserID:           "u1",
		Version:          3,
		EncryptedFileKey: []byte("fk"),
		Nonce:            []byte("n"),
		UploadStatus:     "pending",
		StorageKey:       "skey",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateOrUpdate_VersionConflict(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := `(?s)^INSERT\s+INTO\s+files\b.*ON\s+CONFLICT\s*\(entry_id\)\s*DO\s+UPDATE\s+SET\b.*WHERE\s+files\.entry_id\s*=\s*EXCLUDED\.entry_id;?$`

	mock.ExpectExec(q).
		WithArgs("e1", "u1", int64(1), []byte("fk"), []byte("n"), "pending", "skey").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.CreateOrUpdate(context.Background(), &models.File{
		EntryID:          "e1",
		UserID:           "u1",
		Version:          1,
		EncryptedFileKey: []byte("fk"),
		Nonce:            []byte("n"),
		UploadStatus:     "pending",
		StorageKey:       "skey",
	})
	if !errors.Is(err, common.ErrVersionConflict) {
		t.Fatalf("want ErrVersionConflict, got %v", err)
	}
}

func TestCreateOrUpdate_DBError(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := `(?s)^INSERT\s+INTO\s+files\b.*ON\s+CONFLICT\s*\(entry_id\)\s*DO\s+UPDATE\s+SET\b.*WHERE\s+files\.entry_id\s*=\s*EXCLUDED\.entry_id;?$`

	mock.ExpectExec(q).
		WithArgs("e1", "u1", int64(1), []byte("fk"), []byte("n"), "pending", "skey").
		WillReturnError(errors.New("db down"))

	err := repo.CreateOrUpdate(context.Background(), &models.File{
		EntryID:          "e1",
		UserID:           "u1",
		Version:          1,
		EncryptedFileKey: []byte("fk"),
		Nonce:            []byte("n"),
		UploadStatus:     "pending",
		StorageKey:       "skey",
	})
	if err == nil || !regexp.MustCompile(`db error: .*db down`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped db error, got %v", err)
	}
}

func TestCreateOrUpdate_RowsAffectedErr(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := `(?s)^INSERT\s+INTO\s+files\b.*ON\s+CONFLICT\s*\(entry_id\)\s*DO\s+UPDATE\s+SET\b.*WHERE\s+files\.entry_id\s*=\s*EXCLUDED\.entry_id;?$`

	mock.ExpectExec(q).
		WithArgs("e1", "u1", int64(1), []byte("fk"), []byte("n"), "pending", "skey").
		WillReturnResult(sqlmock.NewErrorResult(errors.New("rows-err")))

	err := repo.CreateOrUpdate(context.Background(), &models.File{
		EntryID:          "e1",
		UserID:           "u1",
		Version:          1,
		EncryptedFileKey: []byte("fk"),
		Nonce:            []byte("n"),
		UploadStatus:     "pending",
		StorageKey:       "skey",
	})
	if err == nil || !regexp.MustCompile(`rows affected error: .*rows-err`).MatchString(err.Error()) {
		t.Fatalf("expected rows affected error, got %v", err)
	}
}

func TestCreateOrUpdate_UnexpectedRowsAffected(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := `(?s)^INSERT\s+INTO\s+files\b.*ON\s+CONFLICT\s*\(entry_id\)\s*DO\s+UPDATE\s+SET\b.*WHERE\s+files\.entry_id\s*=\s*EXCLUDED\.entry_id;?$`

	mock.ExpectExec(q).
		WithArgs("e1", "u1", int64(1), []byte("fk"), []byte("n"), "pending", "skey").
		WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.CreateOrUpdate(context.Background(), &models.File{
		EntryID:          "e1",
		UserID:           "u1",
		Version:          1,
		EncryptedFileKey: []byte("fk"),
		Nonce:            []byte("n"),
		UploadStatus:     "pending",
		StorageKey:       "skey",
	})
	if err == nil || !regexp.MustCompile(`unexpected rows affected: 2`).MatchString(err.Error()) {
		t.Fatalf("expected unexpected rows affected error, got %v", err)
	}
}

func TestSelectUpdated_Success(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := regexp.MustCompile(`SELECT entry_id, user_id, version, encrypted_file_key, nonce, upload_status from files\s+WHERE user_id=\$1 and version>\$2`)
	rows := sqlmock.NewRows([]string{"entry_id", "user_id", "version", "encrypted_file_key", "nonce", "upload_status"}).
		AddRow("e1", "u1", int64(2), []byte("fk1"), []byte("n1"), "pending").
		AddRow("e2", "u1", int64(5), []byte("fk2"), []byte("n2"), "completed")

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
	if got[0].EntryID != "e1" || got[0].UploadStatus != "pending" || got[0].Version != 2 {
		t.Fatalf("bad row[0]: %+v", got[0])
	}
	if got[1].EntryID != "e2" || got[1].UploadStatus != "completed" || got[1].Version != 5 {
		t.Fatalf("bad row[1]: %+v", got[1])
	}
}

func TestSelectUpdated_QueryErr(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := regexp.MustCompile(`SELECT entry_id, user_id, version, encrypted_file_key, nonce, upload_status from files\s+WHERE user_id=\$1 and version>\$2`)
	mock.ExpectQuery(q.String()).
		WithArgs("u1", int64(9)).
		WillReturnError(errors.New("db err"))

	_, err := repo.SelectUpdated(context.Background(), "u1", 9)
	if err == nil || !regexp.MustCompile(`failed to select files: .*db err`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped select error, got %v", err)
	}
}

func TestSelectUpdated_ScanErr(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := regexp.MustCompile(`SELECT entry_id, user_id, version, encrypted_file_key, nonce, upload_status from files\s+WHERE user_id=\$1 and version>\$2`)
	rows := sqlmock.NewRows([]string{"entry_id", "user_id", "version", "encrypted_file_key", "nonce", "upload_status"}).
		AddRow("e1", "u1", "not-int", []byte("fk"), []byte("n"), "pending")

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

	q := regexp.MustCompile(`SELECT entry_id, user_id, version, encrypted_file_key, nonce, upload_status from files\s+WHERE user_id=\$1 and version>\$2`)
	rows := sqlmock.NewRows([]string{"entry_id", "user_id", "version", "encrypted_file_key", "nonce", "upload_status"}).
		AddRow("e1", "u1", int64(2), []byte("fk1"), []byte("n1"), "pending").
		AddRow("e2", "u1", int64(3), []byte("fk2"), []byte("n2"), "completed").
		RowError(1, errors.New("row-err"))

	mock.ExpectQuery(q.String()).
		WithArgs("u1", int64(1)).
		WillReturnRows(rows)

	_, err := repo.SelectUpdated(context.Background(), "u1", 1)
	if err == nil || err.Error() != "row-err" {
		t.Fatalf("expected rows.Err 'row-err', got %v", err)
	}
}

func TestMarkUploaded_OK(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := regexp.MustCompile(`update files set upload_status='completed' where entry_id=\$1`)
	mock.ExpectExec(q.String()).
		WithArgs("e1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.MarkUploaded(context.Background(), "e1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarkUploaded_DBErr(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := regexp.MustCompile(`update files set upload_status='completed' where entry_id=\$1`)
	mock.ExpectExec(q.String()).
		WithArgs("e1").
		WillReturnError(errors.New("db err"))

	err := repo.MarkUploaded(context.Background(), "e1")
	if err == nil || !regexp.MustCompile(`failed to delete entry: .*db err`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped db error, got %v", err)
	}
}

func TestMarkUploaded_RowsAffectedErr(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := regexp.MustCompile(`update files set upload_status='completed' where entry_id=\$1`)
	mock.ExpectExec(q.String()).
		WithArgs("e1").
		WillReturnResult(sqlmock.NewErrorResult(errors.New("rows-err")))

	err := repo.MarkUploaded(context.Background(), "e1")
	if err == nil || !regexp.MustCompile(`failed to get rows affected: .*rows-err`).MatchString(err.Error()) {
		t.Fatalf("expected rows affected error, got %v", err)
	}
}

func TestMarkUploaded_WrongRowsAffectedCount(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := regexp.MustCompile(`update files set upload_status='completed' where entry_id=\$1`)
	mock.ExpectExec(q.String()).
		WithArgs("no-such-id").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.MarkUploaded(context.Background(), "no-such-id")
	if err == nil || !regexp.MustCompile(`wrong rows affected count`).MatchString(err.Error()) {
		t.Fatalf("expected wrong rows affected count error, got %v", err)
	}
}

func TestGetByEntryID_OK(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := regexp.MustCompile(`SELECT entry_id, user_id, storage_key from files\s+WHERE entry_id=\$1`)
	rows := sqlmock.NewRows([]string{"entry_id", "user_id", "storage_key"}).
		AddRow("e1", "u1", "skey")

	mock.ExpectQuery(q.String()).
		WithArgs("e1").
		WillReturnRows(rows)

	got, err := repo.GetByEntryID(context.Background(), "e1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.EntryID != "e1" || got.UserID != "u1" || got.StorageKey != "skey" {
		t.Fatalf("unexpected row: %+v", got)
	}
}

func TestGetByEntryID_QueryErr(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	q := regexp.MustCompile(`SELECT entry_id, user_id, storage_key from files\s+WHERE entry_id=\$1`)
	mock.ExpectQuery(q.String()).
		WithArgs("bad").
		WillReturnError(errors.New("db err"))

	_, err := repo.GetByEntryID(context.Background(), "bad")
	if err == nil || !regexp.MustCompile(`failed to select files: .*db err`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped select error, got %v", err)
	}
}
