package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	sc "github.com/dmitrijs2005/gophkeeper/internal/server/config"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/repomanager"
)

type noopRepoMgr struct{ repomanager.RepositoryManager }

func newSvcForPresign(t *testing.T) (*EntryService, *sql.DB) {
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
	return NewEntryService(db, &noopRepoMgr{}, cfg), db
}

func Test_getPresignClient_SuccessAndError(t *testing.T) {
	svc, db := newSvcForPresign(t)
	defer db.Close()

	origLoad := loadDefaultAWSConfig
	origNewS3 := newS3ClientFromConfig
	origNewPre := newS3PresignClient
	t.Cleanup(func() {
		loadDefaultAWSConfig = origLoad
		newS3ClientFromConfig = origNewS3
		newS3PresignClient = origNewPre
	})

	loadDefaultAWSConfig = func(ctx context.Context, optFns ...func(*awsconfig.LoadOptions) error) (aws.Config, error) {
		if len(optFns) == 0 {
			t.Fatalf("expected config options")
		}
		var lo awsconfig.LoadOptions
		for _, fn := range optFns {
			if err := fn(&lo); err != nil {
				t.Fatalf("load options fn error: %v", err)
			}
		}
		if lo.Region != "us-east-1" {
			t.Fatalf("region not applied: %q", lo.Region)
		}
		return aws.Config{}, nil
	}

	var capturedBaseEndpoint string
	newS3ClientFromConfig = func(cfg aws.Config, optFns ...func(*s3.Options)) *s3.Client {
		var opts s3.Options
		for _, fn := range optFns {
			fn(&opts)
		}
		if opts.BaseEndpoint == nil {
			t.Fatalf("BaseEndpoint not set")
		}
		capturedBaseEndpoint = *opts.BaseEndpoint
		return &s3.Client{}
	}

	newS3PresignClient = func(c *s3.Client) *s3.PresignClient {
		if c == nil {
			t.Fatalf("nil client passed to presign")
		}
		return &s3.PresignClient{}
	}

	pc, err := svc.getPresignClient()
	if err != nil {
		t.Fatalf("getPresignClient err: %v", err)
	}
	if pc == nil {
		t.Fatalf("nil presign client")
	}
	if capturedBaseEndpoint != "http://127.0.0.1:9000" {
		t.Fatalf("BaseEndpoint mismatch: %q", capturedBaseEndpoint)
	}

	loadDefaultAWSConfig = func(ctx context.Context, optFns ...func(*awsconfig.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, errors.New("load-fail")
	}

	pc, err = svc.getPresignClient()
	if err == nil || err.Error() != "load-fail" {
		t.Fatalf("expected load-fail, got %v (pc=%v)", err, pc)
	}
}
