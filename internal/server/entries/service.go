package entries

import (
	"context"
	"fmt"
	"time"

	sc "github.com/dmitrijs2005/gophkeeper/internal/server/config"
	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Service struct {
	repo   Repository
	config *sc.Config
}

func NewService(repo Repository, config *sc.Config) *Service {
	return &Service{
		repo:   repo,
		config: config,
	}
}

func (s *Service) Create(ctx context.Context, userID string, title string, entryType string, cypherText []byte, nonce []byte) (*Entry, error) {

	entry := &Entry{
		UserID:        userID,
		Title:         title,
		Type:          entryType,
		EncryptedData: cypherText,
		Nonce:         nonce,
	}

	user, err := s.repo.Create(ctx, entry)
	if err != nil {
		return nil, fmt.Errorf("error creating entry: %v", err)
	}

	return user, nil
}

func GetRandomStorageKey() string {
	d := time.Now()
	return fmt.Sprintf("users/%d/%d/%d/%v", d.Year(), d.Month(), d.Day(), uuid.New())
}

func (s *Service) getPresignClient() (*s3.PresignClient, error) {
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

func (s *Service) GetPresignedPutUrl(ctx context.Context) (string, string, error) {

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

func (s *Service) GetPresignedGetUrl(ctx context.Context, key string) (string, error) {

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
