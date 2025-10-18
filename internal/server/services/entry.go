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

type EntryService struct {
	db          *sql.DB
	repomanager repomanager.RepositoryManager
	config      *sc.Config
}

func NewEntryService(db *sql.DB, repomanager repomanager.RepositoryManager, config *sc.Config) *EntryService {
	return &EntryService{
		db:          db,
		repomanager: repomanager,
		config:      config,
	}
}

func GetRandomStorageKey() string {
	d := time.Now()
	return fmt.Sprintf("users/%d/%d/%d/%v", d.Year(), d.Month(), d.Day(), uuid.New())
}

func (s *EntryService) getPresignClient() (*s3.PresignClient, error) {
	cfg, err := loadDefaultAWSConfig(context.Background(),
		config.WithRegion(s.config.S3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			s.config.S3RootUser,     // MINIO_ROOT_USER
			s.config.S3RootPassword, // MINIO_ROOT_PASSWORD
			"",
		)))
	if err != nil {
		return nil, err
	}

	client := newS3ClientFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(s.config.S3BaseEndpoint)
	})

	return newS3PresignClient(client), nil
}

func (s *EntryService) GetPresignedPutUrl(ctx context.Context) (string, string, error) {

	presignClient, err := s.getPresignClient()
	if err != nil {
		return "", "", err
	}

	bucket := s.config.S3Bucket
	key := GetRandomStorageKey()

	// Presigned PUT
	req, err := presignPutObject(presignClient, ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(15*time.Minute))

	if err != nil {
		return "", "", err
	}

	return key, req.URL, nil
}

func (s *EntryService) GetPresignedGetUrl(ctx context.Context, key string) (string, error) {

	presignClient, err := s.getPresignClient()
	if err != nil {
		return "", err
	}

	bucket := s.config.S3Bucket

	// Presigned GET
	reg, err := presignGetObject(presignClient, ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", err
	}

	return reg.URL, nil
}

func (s *EntryService) Sync(ctx context.Context, userID string, pendingEntries []*models.Entry,
	pendingFiles []*models.File, maxVersion int64) ([]*models.Entry, []*models.Entry, []*models.File, []*models.FileUploadTask, int64, error) {

	userRepo := s.repomanager.Users(s.db)
	entryRepo := s.repomanager.Entries(s.db)
	fileRepo := s.repomanager.Files(s.db)

	var processedEntries, otherUpdatedEntries []*models.Entry
	var maxServerVersion int64

	otherUpdatedEntries, err := entryRepo.SelectUpdated(ctx, userID, maxVersion)
	if err != nil {
		return nil, nil, nil, nil, 0, err
	}

	otherUpdatedFiles, err := fileRepo.SelectUpdated(ctx, userID, maxVersion)
	if err != nil {
		return nil, nil, nil, nil, 0, err
	}

	var uploadTasks []*models.FileUploadTask
	var newFiles []models.File
	for _, f := range pendingFiles {
		storageKey, url, err := s.GetPresignedPutUrl(ctx)
		if err != nil {
			return nil, nil, nil, nil, 0, err
		}

		f := models.File{
			EntryID:          f.EntryID,
			UserID:           userID,
			Version:          f.Version,
			EncryptedFileKey: f.EncryptedFileKey,
			Nonce:            f.Nonce,
			StorageKey:       storageKey,
			UploadStatus:     "pending",
		}

		newFiles = append(newFiles, f)
		uploadTasks = append(uploadTasks, &models.FileUploadTask{URL: url, EntryID: f.EntryID})
	}

	err = dbx.WithTx(ctx, s.db, nil, func(ctx context.Context, tx dbx.DBTX) error {

		for _, e := range pendingEntries {

			version, err := userRepo.IncrementCurrentVersion(ctx, userID)
			if err != nil {
				return err
			}

			e.Version = version
			maxServerVersion = version

			e.UserID = userID

			err = entryRepo.CreateOrUpdate(ctx, e)
			if err != nil {
				return err
			}

			processedEntries = append(processedEntries, e)

		}

		for _, f := range newFiles {
			err = fileRepo.CreateOrUpdate(ctx, &f)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, nil, nil, nil, 0, fmt.Errorf("error creating entries: %v", err)
	}

	return processedEntries, otherUpdatedEntries, otherUpdatedFiles, uploadTasks, maxServerVersion, nil

}

func (s *EntryService) MarkUploaded(ctx context.Context, id string) error {

	fileRepo := s.repomanager.Files(s.db)

	err := fileRepo.MarkUploaded(ctx, id)
	if err != nil {
		return fmt.Errorf("error updating file: %v", err)
	}

	return nil
}

func (s *EntryService) GetPresignedGetURL(ctx context.Context, id string) (string, error) {

	fileRepo := s.repomanager.Files(s.db)

	f, err := fileRepo.GetByEntryID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("error getting file: %v", err)
	}

	url, err := s.GetPresignedGetUrl(ctx, f.StorageKey)

	return url, err
}
