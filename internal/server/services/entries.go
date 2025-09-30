package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sc "github.com/dmitrijs2005/gophkeeper/internal/server/config"
	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/repomanager"
	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

// func (s *EntryService) Create(ctx context.Context, userID string, title string, entryType string, cypherText []byte, nonce []byte) (*models.Entry, error) {

// 	entry := &models.Entry{
// 		UserID:        userID,
// 		Title:         title,
// 		Type:          entryType,
// 		EncryptedData: cypherText,
// 		Nonce:         nonce,
// 	}

// 	repo := s.repomanager.

// 	user, err := s.repo.Create(ctx, entry)
// 	if err != nil {
// 		return nil, fmt.Errorf("error creating entry: %v", err)
// 	}

// 	return user, nil
// }

func GetRandomStorageKey() string {
	d := time.Now()
	return fmt.Sprintf("users/%d/%d/%d/%v", d.Year(), d.Month(), d.Day(), uuid.New())
}

func (s *EntryService) getPresignClient() (*s3.PresignClient, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(s.config.S3Region), // обязательный параметр
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			s.config.S3RootUser,     // MINIO_ROOT_USER
			s.config.S3RootPassword, // MINIO_ROOT_PASSWORD
			"",                      // токен (не нужен)
		)))
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(s.config.S3BaseEndpoint)
	})

	return s3.NewPresignClient(client), nil
}

func (s *EntryService) GetPresignedPutUrl(ctx context.Context) (string, string, error) {

	presignClient, err := s.getPresignClient()
	if err != nil {
		return "", "", err
	}

	bucket := s.config.S3Bucket
	key := GetRandomStorageKey()

	// Presigned PUT
	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
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
	reg, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", err
	}

	return reg.URL, nil
}

func (s *EntryService) Sync(ctx context.Context, entries []*models.Entry) error {

	// for a, b := range entries {

	// }

	// user, err := s.repo.Create(ctx, entry)
	// if err != nil {
	// 	return nil, fmt.Errorf("error creating entry: %v", err)
	// }

	return nil

}
