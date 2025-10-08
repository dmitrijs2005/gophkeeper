package services

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/dmitrijs2005/gophkeeper/internal/client/client"
	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
	"github.com/dmitrijs2005/gophkeeper/internal/client/repositories/entries"
	"github.com/dmitrijs2005/gophkeeper/internal/client/repositories/files"
	"github.com/dmitrijs2005/gophkeeper/internal/client/repositories/metadata"
	"github.com/dmitrijs2005/gophkeeper/internal/client/utils"
	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
	"github.com/google/uuid"
)

type EntryService interface {
	Sync(ctx context.Context) error
	List(ctx context.Context, masterKey []byte) ([]models.ViewOverview, error)
	Add(ctx context.Context, envelope models.Envelope, file *models.File, masterKey []byte) error
	DeleteByID(ctx context.Context, id string) error
	Get(ctx context.Context, id string, masterKey []byte) (*models.Envelope, error)
}

type entryService struct {
	client client.Client
	db     *sql.DB
}

func NewEntryService(client client.Client, db *sql.DB) EntryService {
	return &entryService{client: client, db: db}
}

func (s *entryService) getMetadataRepo() metadata.Repository {
	return metadata.NewSQLiteRepository(s.db)
}

func (s *entryService) getEntryRepo() entries.Repository {
	return entries.NewSQLiteRepository(s.db)
}

func (s *entryService) getFileRepo() files.Repository {
	return files.NewSQLiteRepository(s.db)
}

func (s *entryService) Add(ctx context.Context, envelope models.Envelope, file *models.File, masterKey []byte) error {

	fmt.Println("masterKey", masterKey)

	overview := envelope.Overview()
	oCipherText, oNonce, err := utils.EncryptEntry(overview, masterKey)

	if err != nil {
		return fmt.Errorf("encryption error1: %w", err)
	}

	cipherText, nonce, err := utils.EncryptEntry(envelope, masterKey)

	if err != nil {
		return fmt.Errorf("encryption error2: %w", err)
	}

	e := &models.Entry{Id: uuid.NewString(),
		Overview:      oCipherText,
		NonceOverview: oNonce,
		Details:       cipherText,
		NonceDetails:  nonce,
	}

	entryRepo := s.getEntryRepo()

	err = dbx.WithTx(ctx, s.db, nil, func(ctx context.Context, tx dbx.DBTX) error {

		err := entryRepo.CreateOrUpdate(ctx, e)
		if err != nil {
			return fmt.Errorf("error tx: %w", err)
		}

		if file != nil {

			fileRepo := s.getFileRepo()

			file.ID = uuid.NewString()
			file.EntryID = e.Id
			file.Deleted = false
			file.UploadStatus = "preupload"

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

	entryRepo := s.getEntryRepo()

	rows, err := entryRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	var result = make([]models.ViewOverview, 0, len(rows))

	for _, row := range rows {

		var x models.Overview
		err = utils.DecryptEntry(row.Overview, row.NonceOverview, masterKey, &x)
		if err != nil {
			log.Printf("error decryption entry: %v", err)
		}

		item := models.ViewOverview{Id: row.Id, Type: string(x.Type), Title: x.Title}
		result = append(result, item)
	}

	return result, nil
}

func (s *entryService) DeleteByID(ctx context.Context, id string) error {

	entryRepo := s.getEntryRepo()
	err := entryRepo.DeleteByID(ctx, id)
	if err != nil {
		return fmt.Errorf("error deleting entry: %w", err)
	}
	return nil
}

func (s *entryService) Get(ctx context.Context, id string, masterKey []byte) (*models.Envelope, error) {

	entryRepo := s.getEntryRepo()
	entry, err := entryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving entry: %w", err)
	}

	var envelope *models.Envelope
	err = utils.DecryptEntry(entry.Details, entry.NonceDetails, masterKey, &envelope)

	if err != nil {
		return nil, fmt.Errorf("error decrypting entry: %w", err)
	}

	return envelope, err
}

func (s *entryService) Sync(ctx context.Context) error {

	metadataRepo := s.getMetadataRepo()
	entryRepo := s.getEntryRepo()

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

	processed, new, max_version, err := s.client.Sync(ctx, entries, currentVersion)
	if err != nil {
		return fmt.Errorf("error client sync: %w", err)
	}

	err = dbx.WithTx(ctx, s.db, nil, func(ctx context.Context, tx dbx.DBTX) error {

		err := metadataRepo.Set(ctx, "current_version", fmt.Appendf(nil, "%v", max_version))
		if err != nil {
			return err
		}

		for _, e := range processed {
			err := entryRepo.CreateOrUpdate(ctx, e)
			if err != nil {
				return err
			}
		}

		for _, e := range new {
			err := entryRepo.CreateOrUpdate(ctx, e)
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
