package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
	sc "github.com/dmitrijs2005/gophkeeper/internal/server/config"
	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
	entriesrepo "github.com/dmitrijs2005/gophkeeper/internal/server/repositories/entries"
	filesrepo "github.com/dmitrijs2005/gophkeeper/internal/server/repositories/files"
	refreshtokensrepo "github.com/dmitrijs2005/gophkeeper/internal/server/repositories/refreshtokens"
	usersrepo "github.com/dmitrijs2005/gophkeeper/internal/server/repositories/users"
)

type fakeUsersRepoSE struct{}

func (f *fakeUsersRepoSE) IncrementCurrentVersion(ctx context.Context, userID string) (int64, error) {
	return 1, nil
}
func (f *fakeUsersRepoSE) Create(context.Context, *models.User) (*models.User, error) {
	return nil, nil
}
func (f *fakeUsersRepoSE) GetUserByLogin(context.Context, string) (*models.User, error) {
	return nil, nil
}

type fakeEntriesRepoSE struct{}

func (f *fakeEntriesRepoSE) SelectUpdated(ctx context.Context, userID string, minVersion int64) ([]*models.Entry, error) {
	return nil, nil
}
func (f *fakeEntriesRepoSE) CreateOrUpdate(context.Context, *models.Entry) error { return nil }

type fakeFilesRepoSE struct{}

func (f *fakeFilesRepoSE) SelectUpdated(ctx context.Context, userID string, minVersion int64) ([]*models.File, error) {
	return nil, nil
}
func (f *fakeFilesRepoSE) CreateOrUpdate(context.Context, *models.File) error { return nil }
func (f *fakeFilesRepoSE) MarkUploaded(context.Context, string) error         { return nil }
func (f *fakeFilesRepoSE) GetByEntryID(context.Context, string) (*models.File, error) {
	return nil, nil
}

type fakeRepoMgrSE struct {
	u *fakeUsersRepoSE
	e *fakeEntriesRepoSE
	f *fakeFilesRepoSE
}

func (m *fakeRepoMgrSE) RunMigrations(context.Context, *sql.DB) error           { return nil }
func (m *fakeRepoMgrSE) Users(db dbx.DBTX) usersrepo.Repository                 { return m.u }
func (m *fakeRepoMgrSE) Entries(db dbx.DBTX) entriesrepo.Repository             { return m.e }
func (m *fakeRepoMgrSE) Files(db dbx.DBTX) filesrepo.Repository                 { return m.f }
func (m *fakeRepoMgrSE) RefreshTokens(db dbx.DBTX) refreshtokensrepo.Repository { return nil }

func TestSync_PresignPutError_NoTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New err: %v", err)
	}
	defer db.Close()

	cfg := &sc.Config{
		S3Region:       "us-east-1",
		S3RootUser:     "minioadmin",
		S3RootPassword: "minioadmin",
		S3BaseEndpoint: "http://127.0.0.1:9000",
		S3Bucket:       "gophkeeper",
		SecretKey:      "k",
	}
	svc := NewEntryService(db, &fakeRepoMgrSE{
		u: &fakeUsersRepoSE{}, e: &fakeEntriesRepoSE{}, f: &fakeFilesRepoSE{},
	}, cfg)

	orig := loadDefaultAWSConfig
	defer func() { loadDefaultAWSConfig = orig }()
	loadDefaultAWSConfig = func(ctx context.Context, optFns ...func(*awsconfig.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, errors.New("presign-fail")
	}

	_, _, _, _, _, err = svc.Sync(context.Background(), "u1",
		[]*models.Entry{{ID: "e1"}},
		[]*models.File{{EntryID: "e1", Version: 1}},
		0,
	)
	if err == nil || err.Error() != "presign-fail" {
		t.Fatalf("want presign-fail, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
