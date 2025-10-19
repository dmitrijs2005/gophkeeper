package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	sc "github.com/dmitrijs2005/gophkeeper/internal/server/config"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/repomanager"
)

type noopRepoMgr2 struct{ repomanager.RepositoryManager }

func newSvcForPresignWrappers(t *testing.T) (*EntryService, *sql.DB) {
	t.Helper()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New err: %v", err)
	}
	cfg := &sc.Config{
		S3Region:       "us-east-1",
		S3RootUser:     "minioadmin",
		S3RootPassword: "minioadmin",
		S3BaseEndpoint: "http://127.0.0.1:9000",
		S3Bucket:       "gophkeeper",
		SecretKey:      "k",
	}
	return NewEntryService(db, &noopRepoMgr2{}, cfg), db
}

func TestGetPresignedPutUrl_ErrorFromClientFactory(t *testing.T) {
	svc, db := newSvcForPresignWrappers(t)
	defer db.Close()

	orig := loadDefaultAWSConfig
	defer func() { loadDefaultAWSConfig = orig }()
	loadDefaultAWSConfig = func(ctx context.Context, optFns ...func(*awsconfig.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, errors.New("load-fail")
	}

	_, _, err := svc.GetPresignedPutUrl(context.Background())
	if err == nil || err.Error() != "load-fail" {
		t.Fatalf("want load-fail, got %v", err)
	}
}

func TestGetPresignedGetUrl_ErrorFromClientFactory(t *testing.T) {
	svc, db := newSvcForPresignWrappers(t)
	defer db.Close()

	orig := loadDefaultAWSConfig
	defer func() { loadDefaultAWSConfig = orig }()
	loadDefaultAWSConfig = func(ctx context.Context, optFns ...func(*awsconfig.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, errors.New("load-fail")
	}

	_, err := svc.GetPresignedGetUrl(context.Background(), "any-key")
	if err == nil || err.Error() != "load-fail" {
		t.Fatalf("want load-fail, got %v", err)
	}
}
