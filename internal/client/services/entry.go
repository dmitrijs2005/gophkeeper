package services

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/dmitrijs2005/gophkeeper/internal/client/client"
	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
	"github.com/dmitrijs2005/gophkeeper/internal/client/repositories/entries"
	"github.com/dmitrijs2005/gophkeeper/internal/client/repositories/files"
	"github.com/dmitrijs2005/gophkeeper/internal/client/repositories/metadata"
	"github.com/dmitrijs2005/gophkeeper/internal/cryptox"
	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
	"github.com/dmitrijs2005/gophkeeper/internal/netx"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type EntryService interface {
	Sync(ctx context.Context) error
	List(ctx context.Context, masterKey []byte) ([]models.ViewOverview, error)
	Add(ctx context.Context, envelope models.Envelope, file *models.File, masterKey []byte) error
	DeleteByID(ctx context.Context, id string) error
	Get(ctx context.Context, id string, masterKey []byte) (*models.Envelope, error)
	GetPresignedGetUrl(ctx context.Context, id string) (string, error)
	GetFile(ctx context.Context, id string) (*models.File, error)
}

type entryService struct {
	client client.Client
	db     *sql.DB
}

func NewEntryService(client client.Client, db *sql.DB) EntryService {
	return &entryService{client: client, db: db}
}

func (s *entryService) getMetadataRepo(db dbx.DBTX) metadata.Repository {
	return metadata.NewSQLiteRepository(db)
}

func (s *entryService) getEntryRepo(db dbx.DBTX) entries.Repository {
	return entries.NewSQLiteRepository(db)
}

func (s *entryService) getFileRepo(db dbx.DBTX) files.Repository {
	return files.NewSQLiteRepository(db)
}

func (s *entryService) Add(ctx context.Context, envelope models.Envelope, file *models.File, masterKey []byte) error {

	overview := envelope.Overview()
	oCipherText, oNonce, err := cryptox.EncryptEntry(overview, masterKey)

	if err != nil {
		return fmt.Errorf("encryption error1: %w", err)
	}

	cipherText, nonce, err := cryptox.EncryptEntry(envelope, masterKey)

	if err != nil {
		return fmt.Errorf("encryption error2: %w", err)
	}

	e := &models.Entry{Id: uuid.NewString(),
		Overview:      oCipherText,
		NonceOverview: oNonce,
		Details:       cipherText,
		NonceDetails:  nonce,
	}

	err = dbx.WithTx(ctx, s.db, nil, func(ctx context.Context, tx dbx.DBTX) error {

		entryRepo := s.getEntryRepo(tx)

		err := entryRepo.CreateOrUpdate(ctx, e)
		if err != nil {
			return fmt.Errorf("error tx: %w", err)
		}

		if file != nil {

			fileRepo := s.getFileRepo(tx)

			file.EntryID = e.Id
			file.Deleted = false
			file.UploadStatus = "pending"

			err := fileRepo.CreateOrUpdate(ctx, file)
			if err != nil {
				return fmt.Errorf("error tx: %w", err)
			}

		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error tx: %w", err)
	}

	return nil

}

func (s *entryService) List(ctx context.Context, masterKey []byte) ([]models.ViewOverview, error) {

	entryRepo := s.getEntryRepo(s.db)

	rows, err := entryRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	var result = make([]models.ViewOverview, 0, len(rows))

	for _, row := range rows {

		var x models.Overview
		err = cryptox.DecryptEntry(row.Overview, row.NonceOverview, masterKey, &x)
		if err != nil {
			log.Printf("error decryption entry: %v", err)
		}

		item := models.ViewOverview{Id: row.Id, Type: string(x.Type), Title: x.Title}
		result = append(result, item)
	}

	return result, nil
}

func (s *entryService) DeleteByID(ctx context.Context, id string) error {

	entryRepo := s.getEntryRepo(s.db)
	err := entryRepo.DeleteByID(ctx, id)
	if err != nil {
		return fmt.Errorf("error deleting entry: %w", err)
	}
	return nil
}

func (s *entryService) Get(ctx context.Context, id string, masterKey []byte) (*models.Envelope, error) {

	entryRepo := s.getEntryRepo(s.db)
	entry, err := entryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving entry: %w", err)
	}

	var envelope *models.Envelope
	err = cryptox.DecryptEntry(entry.Details, entry.NonceDetails, masterKey, &envelope)

	if err != nil {
		return nil, fmt.Errorf("error decrypting entry: %w", err)
	}

	return envelope, err
}

func (s *entryService) uploadPendingFiles(ctx context.Context, uploadTasks []*models.FileUploadTask) error {

	fileRepo := s.getFileRepo(s.db)
	grp, ctx := errgroup.WithContext(ctx)

	sem := make(chan struct{}, 4)
	var wg sync.WaitGroup

	for _, task := range uploadTasks {
		task := task
		wg.Add(1)
		sem <- struct{}{}

		grp.Go(func() error {
			defer wg.Done()
			defer func() { <-sem }()

			file, err := fileRepo.GetByEntryID(ctx, task.EntryID)
			if err != nil {
				return err
			}

			if file.LocalPath == "" {
				return err
			}

			data, err := os.ReadFile(file.LocalPath)
			if err != nil {
				return err
			}

			err = netx.UploadToS3PresignedURL(task.URL, data)
			if err != nil {
				return err
			}

			if err := fileRepo.MarkUploaded(ctx, task.EntryID); err != nil {
				return err
			}

			if err := s.client.MarkUploaded(ctx, task.EntryID); err != nil {
				return err
			}

			_ = os.Remove(file.LocalPath)

			return nil

		})
	}

	if err := grp.Wait(); err != nil {
		return err
	}

	return nil

}

func (s *entryService) Sync(ctx context.Context) error {

	metadataRepo := s.getMetadataRepo(s.db)
	entryRepo := s.getEntryRepo(s.db)
	fileRepo := s.getFileRepo(s.db)

	value, err := metadataRepo.Get(ctx, "current_version")
	if err != nil {
		return fmt.Errorf("error retrieving current version: %w", err)
	}

	sValue := string(bytes.TrimSpace(value))
	var currentVersion int64
	if sValue == "" {
		currentVersion = int64(0)
	} else {
		currentVersion, err = strconv.ParseInt(sValue, 10, 64)
		if err != nil {
			return fmt.Errorf("parse current_version: %w", err)
		}
	}

	entries, err := entryRepo.GetAllPending(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving entries: %w", err)
	}

	files, err := fileRepo.GetAllPendingUpload(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving files: %w", err)
	}

	processedEntries, newEntries, newFiles, uploadTasks, max_version, err := s.client.Sync(ctx, entries, files, currentVersion)
	if err != nil {
		return fmt.Errorf("error client sync: %w", err)
	}

	err = s.uploadPendingFiles(ctx, uploadTasks)
	if err != nil {
		return fmt.Errorf("error uploading files: %w", err)
	}

	err = dbx.WithTx(ctx, s.db, nil, func(ctx context.Context, tx dbx.DBTX) error {

		metadataRepoTx := s.getMetadataRepo(tx)
		entryRepoTx := s.getEntryRepo(tx)
		fileRepoTx := s.getFileRepo(tx)

		err := metadataRepoTx.Set(ctx, "current_version", fmt.Appendf(nil, "%v", max_version))
		if err != nil {
			return err
		}

		for _, e := range processedEntries {
			err := entryRepoTx.CreateOrUpdate(ctx, e)
			if err != nil {
				return err
			}
		}

		for _, e := range newEntries {
			err := entryRepoTx.CreateOrUpdate(ctx, e)
			if err != nil {
				return err
			}
		}

		for _, e := range newFiles {
			err := fileRepoTx.CreateOrUpdate(ctx, e)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error tx: %w", err)
	}

	return nil
}

func (s *entryService) GetPresignedGetUrl(ctx context.Context, id string) (string, error) {

	url, err := s.client.GetPresignedGetURL(ctx, id)
	if err != nil {
		return "", fmt.Errorf("error get presigned url: %w", err)
	}

	return url, err

}

func (s *entryService) GetFile(ctx context.Context, id string) (*models.File, error) {
	fileRepo := s.getFileRepo(s.db)
	return fileRepo.GetByEntryID(ctx, id)
}
