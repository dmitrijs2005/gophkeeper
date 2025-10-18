// Package services contains server-side business logic. This file implements
// EntryService, which coordinates sync of entries/files with the database and
// generates presigned S3 URLs for client uploads/downloads.
package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
	sc "github.com/dmitrijs2005/gophkeeper/internal/server/config"
	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/repomanager"
	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	// seams for unit tests
	loadDefaultAWSConfig = config.LoadDefaultConfig

	newS3ClientFromConfig = func(cfg aws.Config, optFns ...func(*s3.Options)) *s3.Client {
		return s3.NewFromConfig(cfg, optFns...)
	}
	newS3PresignClient = func(c *s3.Client) *s3.PresignClient {
		return s3.NewPresignClient(c)
	}
	presignPutObject = func(pc *s3.PresignClient, ctx context.Context, in *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
		return pc.PresignPutObject(ctx, in, optFns...)
	}
	presignGetObject = func(pc *s3.PresignClient, ctx context.Context, in *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
		return pc.PresignGetObject(ctx, in, optFns...)
	}
)

// EntryService implements server-side entry/file synchronization and presigned
// URL generation against an S3-compatible backend.
type EntryService struct {
	db          *sql.DB
	repomanager repomanager.RepositoryManager
	config      *sc.Config
}

// NewEntryService wires the service with a DB handle, repository manager, and config.
func NewEntryService(db *sql.DB, repomanager repomanager.RepositoryManager, config *sc.Config) *EntryService {
	return &EntryService{db: db, repomanager: repomanager, config: config}
}

// GetRandomStorageKey produces a time-bucketed object-storage key for new uploads.
func GetRandomStorageKey() string {
	d := time.Now()
	return fmt.Sprintf("users/%d/%d/%d/%v", d.Year(), d.Month(), d.Day(), uuid.New())
}

// getPresignClient builds an S3 presign client using config-provided endpoint,
// region, and static credentials (e.g., MinIO).
func (s *EntryService) getPresignClient() (*s3.PresignClient, error) {
	cfg, err := loadDefaultAWSConfig(context.Background(),
		config.WithRegion(s.config.S3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			s.config.S3RootUser, s.config.S3RootPassword, "",
		)))
	if err != nil {
		return nil, err
	}
	client := newS3ClientFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(s.config.S3BaseEndpoint)
	})
	return newS3PresignClient(client), nil
}

// GetPresignedPutUrl returns (storageKey, url) for a client to PUT an encrypted file.
// The URL is short-lived and suitable for direct upload from the client.
func (s *EntryService) GetPresignedPutUrl(ctx context.Context) (string, string, error) {
	presignClient, err := s.getPresignClient()
	if err != nil {
		return "", "", err
	}
	bucket := s.config.S3Bucket
	key := GetRandomStorageKey()

	req, err := presignPutObject(presignClient, ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", "", err
	}
	return key, req.URL, nil
}

// GetPresignedGetUrl returns a short-lived URL to GET an object by storage key.
func (s *EntryService) GetPresignedGetUrl(ctx context.Context, key string) (string, error) {
	presignClient, err := s.getPresignClient()
	if err != nil {
		return "", err
	}
	bucket := s.config.S3Bucket
	reg, err := presignGetObject(presignClient, ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", err
	}
	return reg.URL, nil
}

// Sync merges client-submitted pending entries/files with server state,
// returns processed (server-accepted) entries, server-side updates since
// client maxVersion, new files created on the server, upload tasks for
// the client, and the new global max version.
//
// Workflow (simplified):
//  1. Fetch server updates (entries/files) newer than client's maxVersion.
//  2. For each pending file, generate a storage key + presigned PUT URL.
//  3. In a transaction:
//     - For each pending entry, increment user's global version and upsert.
//     - Upsert new file metadata (pending state).
//  4. Return processed entries, server updates, file upload tasks, and max version.
func (s *EntryService) Sync(
	ctx context.Context,
	userID string,
	pendingEntries []*models.Entry,
	pendingFiles []*models.File,
	maxVersion int64,
) ([]*models.Entry, []*models.Entry, []*models.File, []*models.FileUploadTask, int64, error) {

	userRepo := s.repomanager.Users(s.db)
	entryRepo := s.repomanager.Entries(s.db)
	fileRepo := s.repomanager.Files(s.db)

	otherUpdatedEntries, err := entryRepo.SelectUpdated(ctx, userID, maxVersion)
	if err != nil {
		return nil, nil, nil, nil, 0, err
	}
	otherUpdatedFiles, err := fileRepo.SelectUpdated(ctx, userID, maxVersion)
	if err != nil {
		return nil, nil, nil, nil, 0, err
	}

	var (
		processedEntries []*models.Entry
		uploadTasks      []*models.FileUploadTask
		newFiles         []models.File
		maxServerVersion int64
	)

	// Prepare file records + presigned PUTs
	for _, f := range pendingFiles {
		storageKey, url, err := s.GetPresignedPutUrl(ctx)
		if err != nil {
			return nil, nil, nil, nil, 0, err
		}
		newFiles = append(newFiles, models.File{
			EntryID:          f.EntryID,
			UserID:           userID,
			Version:          f.Version,
			EncryptedFileKey: f.EncryptedFileKey,
			Nonce:            f.Nonce,
			StorageKey:       storageKey,
			UploadStatus:     "pending",
		})
		uploadTasks = append(uploadTasks, &models.FileUploadTask{URL: url, EntryID: f.EntryID})
	}

	// Persist entries and file rows transactionally.
	if err := dbx.WithTx(ctx, s.db, nil, func(ctx context.Context, tx dbx.DBTX) error {
		for _, e := range pendingEntries {
			version, err := userRepo.IncrementCurrentVersion(ctx, userID)
			if err != nil {
				return err
			}
			e.Version = version
			maxServerVersion = version
			e.UserID = userID
			if err := entryRepo.CreateOrUpdate(ctx, e); err != nil {
				return err
			}
			processedEntries = append(processedEntries, e)
		}
		for _, f := range newFiles {
			if err := fileRepo.CreateOrUpdate(ctx, &f); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, nil, nil, nil, 0, fmt.Errorf("error creating entries: %v", err)
	}

	return processedEntries, otherUpdatedEntries, otherUpdatedFiles, uploadTasks, maxServerVersion, nil
}

// MarkUploaded marks the file for the given entry as uploaded (completed).
func (s *EntryService) MarkUploaded(ctx context.Context, id string) error {
	fileRepo := s.repomanager.Files(s.db)
	if err := fileRepo.MarkUploaded(ctx, id); err != nil {
		return fmt.Errorf("error updating file: %v", err)
	}
	return nil
}

// GetPresignedGetURL returns a presigned GET URL for the file associated
// with the given entry ID after verifying ownership and loading storage key.
func (s *EntryService) GetPresignedGetURL(ctx context.Context, id string) (string, error) {
	fileRepo := s.repomanager.Files(s.db)
	f, err := fileRepo.GetByEntryID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("error getting file: %v", err)
	}
	url, err := s.GetPresignedGetUrl(ctx, f.StorageKey)
	return url, err
}
