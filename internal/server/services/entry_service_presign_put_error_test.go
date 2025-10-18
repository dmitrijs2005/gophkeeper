package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	sc "github.com/dmitrijs2005/gophkeeper/internal/server/config"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/repomanager"
)

type noopRM struct{ repomanager.RepositoryManager }

func newEntrySvc(t *testing.T) (*EntryService, *sql.DB) {
	t.Helper()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	cfg := &sc.Config{
		S3Region:                     "us-east-1",
		S3RootUser:                   "minioadmin",
		S3RootPassword:               "minioadmin",
		S3BaseEndpoint:               "http://127.0.0.1:9000",
		S3Bucket:                     "gophkeeper",
		SecretKey:                    "k",
		AccessTokenValidityDuration:  time.Minute,
		RefreshTokenValidityDuration: time.Hour,
	}
	return NewEntryService(db, &noopRM{}, cfg), db
}

func TestGetPresignedPutUrl_ErrorFromPresign(t *testing.T) {
	svc, db := newEntrySvc(t)
	defer db.Close()

	origLoad, origNewS3, origNewPre := loadDefaultAWSConfig, newS3ClientFromConfig, newS3PresignClient
	t.Cleanup(func() {
		loadDefaultAWSConfig = origLoad
		newS3ClientFromConfig = origNewS3
		newS3PresignClient = origNewPre
		presignPutObject = func(pc *s3.PresignClient, ctx context.Context, in *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
			return pc.PresignPutObject(ctx, in, optFns...)
		}
	})

	loadDefaultAWSConfig = func(ctx context.Context, optFns ...func(*awsconfig.LoadOptions) error) (aws.Config, error) {
		var lo awsconfig.LoadOptions
		for _, fn := range optFns {
			_ = fn(&lo)
		}
		lo.Region = "us-east-1"
		return aws.Config{}, nil
	}
	newS3ClientFromConfig = func(cfg aws.Config, optFns ...func(*s3.Options)) *s3.Client { return &s3.Client{} }
	newS3PresignClient = func(c *s3.Client) *s3.PresignClient { return &s3.PresignClient{} }

	presignPutObject = func(pc *s3.PresignClient, ctx context.Context, in *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
		return nil, errors.New("presign-put-fail")
	}

	_, _, err := svc.GetPresignedPutUrl(context.Background())
	if err == nil || err.Error() != "presign-put-fail" {
		t.Fatalf("want presign-put-fail, got %v", err)
	}
}
