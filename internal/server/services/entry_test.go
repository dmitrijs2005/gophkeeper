package services

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
	"github.com/dmitrijs2005/gophkeeper/internal/server/config"
	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/entries"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/files"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/repomanager"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/users"
)

// -------- test fakes --------

type fakeUsersRepo struct {
	users.Repository
	incVer int64
	err    error
}

func (f *fakeUsersRepo) IncrementCurrentVersion(ctx context.Context, userID string) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	f.incVer++
	return f.incVer, nil
}

type fakeEntriesRepo struct {
	entries.Repository
	selUpdated []*models.Entry
	selErr     error

	createErr error

	created []*models.Entry
}

func (f *fakeEntriesRepo) SelectUpdated(ctx context.Context, userID string, minVersion int64) ([]*models.Entry, error) {
	return f.selUpdated, f.selErr
}

func (f *fakeEntriesRepo) CreateOrUpdate(ctx context.Context, e *models.Entry) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, e)
	return nil
}

type fakeFilesRepo struct {
	files.Repository
	selUpdated []*models.File
	selErr     error

	markErr error

	getByID *models.File
	getErr  error
}

func (f *fakeFilesRepo) SelectUpdated(ctx context.Context, userID string, minVersion int64) ([]*models.File, error) {
	return f.selUpdated, f.selErr
}
func (f *fakeFilesRepo) CreateOrUpdate(ctx context.Context, file *models.File) error {
	return nil
}
func (f *fakeFilesRepo) MarkUploaded(ctx context.Context, id string) error {
	return f.markErr
}
func (f *fakeFilesRepo) GetByEntryID(ctx context.Context, id string) (*models.File, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.getByID, nil
}

type fakeRepoManager struct {
	repomanager.RepositoryManager
	u *fakeUsersRepo
	e *fakeEntriesRepo
	f *fakeFilesRepo
}

func (m *fakeRepoManager) Users(dbx dbx.DBTX) users.Repository     { return m.u }
func (m *fakeRepoManager) Entries(dbx dbx.DBTX) entries.Repository { return m.e }
func (m *fakeRepoManager) Files(dbx dbx.DBTX) files.Repository     { return m.f }

// -------- helpers --------

func newSQLMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	return db, mock
}

func newService(t *testing.T, db *sql.DB, m *fakeRepoManager) *EntryService {
	t.Helper()
	cfg := &config.Config{
		S3Region:       "us-east-1",
		S3RootUser:     "x",
		S3RootPassword: "y",
		S3BaseEndpoint: "http://127.0.0.1:9000",
		S3Bucket:       "bucket",
		SecretKey:      "k",
	}
	return NewEntryService(db, m, cfg)
}

// -------- tests --------

func TestSync_Success_NoFiles(t *testing.T) {
	db, mock := newSQLMockDB(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	u := &fakeUsersRepo{incVer: 0}
	e := &fakeEntriesRepo{
		selUpdated: []*models.Entry{
			{ID: "o1", Version: 10},
		},
	}
	f := &fakeFilesRepo{
		selUpdated: []*models.File{
			{EntryID: "f1", Version: 11},
		},
	}
	m := &fakeRepoManager{u: u, e: e, f: f}

	s := newService(t, db, m)

	ctx := context.Background()
	pendingEntries := []*models.Entry{
		{ID: "p1", Overview: []byte("o"), NonceOverview: []byte("no"), Details: []byte("d"), NonceDetails: []byte("nd")},
		{ID: "p2"},
	}
	pendingFiles := []*models.File{}
	processed, otherEntries, otherFiles, uploadTasks, maxVer, err := s.Sync(ctx, "user-1", pendingEntries, pendingFiles, 1)
	if err != nil {
		t.Fatalf("Sync error: %v", err)
	}

	if maxVer != 2 {
		t.Fatalf("unexpected maxVersion: %d", maxVer)
	}
	if len(processed) != 2 || processed[0].ID != "p1" || processed[1].ID != "p2" {
		t.Fatalf("unexpected processed entries: %+v", processed)
	}
	if processed[0].Version != 1 || processed[1].Version != 2 {
		t.Fatalf("unexpected versions: %d, %d", processed[0].Version, processed[1].Version)
	}
	if len(otherEntries) != 1 || otherEntries[0].ID != "o1" {
		t.Fatalf("unexpected other entries: %+v", otherEntries)
	}
	if len(otherFiles) != 1 || otherFiles[0].EntryID != "f1" {
		t.Fatalf("unexpected other files: %+v", otherFiles)
	}
	if len(uploadTasks) != 0 {
		t.Fatalf("unexpected upload tasks: %+v", uploadTasks)
	}

	if len(e.created) != 2 || e.created[0].ID != "p1" || e.created[1].ID != "p2" {
		t.Fatalf("Entries.CreateOrUpdate calls: %+v", e.created)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSync_ErrorsBeforeTx(t *testing.T) {
	db, _ := newSQLMockDB(t)
	defer db.Close()

	e1 := &fakeEntriesRepo{selErr: errBoom{}}
	m1 := &fakeRepoManager{u: &fakeUsersRepo{}, e: e1, f: &fakeFilesRepo{}}
	s1 := newService(t, db, m1)
	_, _, _, _, _, err := s1.Sync(context.Background(), "u", nil, nil, 0)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("want entries select error, got %v", err)
	}

	e2 := &fakeEntriesRepo{selUpdated: nil}
	f2 := &fakeFilesRepo{selErr: errBoom{}}
	m2 := &fakeRepoManager{u: &fakeUsersRepo{}, e: e2, f: f2}
	s2 := newService(t, db, m2)
	_, _, _, _, _, err = s2.Sync(context.Background(), "u", nil, nil, 0)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("want files select error, got %v", err)
	}
}

func TestSync_ErrorsInsideTx(t *testing.T) {
	db, mock := newSQLMockDB(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	u := &fakeUsersRepo{err: errBoom{}}
	e := &fakeEntriesRepo{}
	m := &fakeRepoManager{u: u, e: e, f: &fakeFilesRepo{}}
	s := newService(t, db, m)

	_, _, _, _, _, err := s.Sync(context.Background(), "u", []*models.Entry{{ID: "p1"}}, nil, 0)
	if err == nil || !strings.Contains(err.Error(), "error creating entries:") {
		t.Fatalf("want wrapped tx error, got %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectRollback()

	u2 := &fakeUsersRepo{}
	e2 := &fakeEntriesRepo{createErr: errBoom{}}
	m2 := &fakeRepoManager{u: u2, e: e2, f: &fakeFilesRepo{}}
	s2 := newService(t, db, m2)

	_, _, _, _, _, err = s2.Sync(context.Background(), "u", []*models.Entry{{ID: "p1"}}, nil, 0)
	if err == nil || !strings.Contains(err.Error(), "error creating entries:") {
		t.Fatalf("want wrapped tx error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestMarkUploaded_OKAndError(t *testing.T) {
	db, _ := newSQLMockDB(t)
	defer db.Close()

	okFiles := &fakeFilesRepo{}
	errFiles := &fakeFilesRepo{markErr: errBoom{}}

	s1 := newService(t, db, &fakeRepoManager{u: &fakeUsersRepo{}, e: &fakeEntriesRepo{}, f: okFiles})
	if err := s1.MarkUploaded(context.Background(), "e1"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	s2 := newService(t, db, &fakeRepoManager{u: &fakeUsersRepo{}, e: &fakeEntriesRepo{}, f: errFiles})
	if err := s2.MarkUploaded(context.Background(), "e1"); err == nil || !strings.Contains(err.Error(), "error updating file:") {
		t.Fatalf("want wrapped error, got %v", err)
	}
}

func TestGetPresignedGetURL_ErrOnGetByEntryID(t *testing.T) {
	db, _ := newSQLMockDB(t)
	defer db.Close()

	filesRepo := &fakeFilesRepo{getErr: errBoom{}}
	s := newService(t, db, &fakeRepoManager{u: &fakeUsersRepo{}, e: &fakeEntriesRepo{}, f: filesRepo})

	_, err := s.GetPresignedGetURL(context.Background(), "e1")
	if err == nil || !strings.Contains(err.Error(), "error getting file:") {
		t.Fatalf("want wrapped error, got %v", err)
	}
}

func TestGetRandomStorageKey_Format(t *testing.T) {
	k := GetRandomStorageKey()
	// users/YYYY/M/D/UUID
	re := regexp.MustCompile(`^users/\d{4}/\d{1,2}/\d{1,2}/[0-9a-fA-F-]+$`)
	if !re.MatchString(k) {
		t.Fatalf("unexpected format: %q", k)
	}
}

type errBoom struct{}

func (errBoom) Error() string { return "boom" }
