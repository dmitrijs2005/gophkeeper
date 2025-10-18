// Package services implements application services for the GophKeeper client.
// This file defines EntryService: listing, retrieval, creation (with optional
// file staging), synchronization with the server, and file upload orchestration.
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

// EntryService coordinates entry CRUD, local encryption, sync, and file I/O.
type EntryService interface {
	// Sync performs a bidirectional synchronization with the backend.
	Sync(ctx context.Context) error

	// List returns decrypted overviews for display using the provided master key.
	List(ctx context.Context, masterKey []byte) ([]models.ViewOverview, error)

	// Add encrypts and stores an envelope (and optional staged file) locally.
	Add(ctx context.Context, envelope models.Envelope, file *models.File, masterKey []byte) error

	// DeleteByID marks an entry as deleted (implementation-defined).
	DeleteByID(ctx context.Context, id string) error

	// Get returns and decrypts a single entry envelope by id.
	Get(ctx context.Context, id string, masterKey []byte) (*models.Envelope, error)

	// GetPresignedGetUrl requests a presigned URL for downloading a file.
	GetPresignedGetUrl(ctx context.Context, id string) (string, error)

	// GetFile loads file metadata for an entry.
	GetFile(ctx context.Context, id string) (*models.File, error)
}

// entryService is the concrete EntryService backed by repositories and a Client.
type entryService struct {
	client client.Client
	db     *sql.DB
}

// NewEntryService constructs an EntryService bound to the given API client and DB.
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

// Add encrypts the envelope overview and details with masterKey, creates a new
// local Entry (with a generated id), and optionally stores file metadata as a
// pending upload in the same transaction.
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

	e := &models.Entry{
		Id:            uuid.NewString(),
		Overview:      oCipherText,
		NonceOverview: oNonce,
		Details:       cipherText,
		NonceDetails:  nonce,
	}

	err = dbx.WithTx(ctx, s.db, nil, func(ctx context.Context, tx dbx.DBTX) error {
		entryRepo := s.getEntryRepo(tx)
		if err := entryRepo.CreateOrUpdate(ctx, e); err != nil {
			return fmt.Errorf("error tx: %w", err)
		}
		if file != nil {
			fileRepo := s.getFileRepo(tx)
			file.EntryID = e.Id
			file.Deleted = false
			file.UploadStatus = "pending"
			if err := fileRepo.CreateOrUpdate(ctx, file); err != nil {
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

// List enumerates non-deleted entries and decrypts their Overview structures.
func (s *entryService) List(ctx context.Context, masterKey []byte) ([]models.ViewOverview, error) {
	entryRepo := s.getEntryRepo(s.db)
	rows, err := entryRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	result := make([]models.ViewOverview, 0, len(rows))
	for _, row := range rows {
		var x models.Overview
		if err := cryptox.DecryptEntry(row.Overview, row.NonceOverview, masterKey, &x); err != nil {
			log.Printf("error decryption entry: %v", err)
		}
		result = append(result, models.ViewOverview{Id: row.Id, Type: string(x.Type), Title: x.Title})
	}
	return result, nil
}

// DeleteByID soft-deletes an entry.
func (s *entryService) DeleteByID(ctx context.Context, id string) error {
	if err := s.getEntryRepo(s.db).DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("error deleting entry: %w", err)
	}
	return nil
}

// Get fetches and decrypts a single entry envelope using masterKey.
func (s *entryService) Get(ctx context.Context, id string, masterKey []byte) (*models.Envelope, error) {
	entry, err := s.getEntryRepo(s.db).GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving entry: %w", err)
	}
	var envelope *models.Envelope
	if err := cryptox.DecryptEntry(entry.Details, entry.NonceDetails, masterKey, &envelope); err != nil {
		return nil, fmt.Errorf("error decrypting entry: %w", err)
	}
	return envelope, nil
}

// uploadPendingFiles uploads staged ciphertexts to the server using presigned
// URLs, marks them uploaded both locally and remotely, and removes temp files.
// Uploads are done concurrently with a small semaphore for backpressure.
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
			if err := netx.UploadToS3PresignedURL(task.URL, data); err != nil {
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

// Sync reconciles local pending changes with the server, uploads any staged
// files when requested, and applies server changes in a transaction.
//
// Flow:
//  1. Read current_version from metadata (defaults to 0).
//  2. Collect pending entries/files.
//  3. Call client.Sync(entries, files, currentVersion).
//  4. Upload files for any returned upload tasks.
//  5. In a TX, persist processed/new entries and files, and update current_version.
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
		currentVersion = 0
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

	if err := s.uploadPendingFiles(ctx, uploadTasks); err != nil {
		return fmt.Errorf("error uploading files: %w", err)
	}

	if err := dbx.WithTx(ctx, s.db, nil, func(ctx context.Context, tx dbx.DBTX) error {
		metadataRepoTx := s.getMetadataRepo(tx)
		entryRepoTx := s.getEntryRepo(tx)
		fileRepoTx := s.getFileRepo(tx)

		if err := metadataRepoTx.Set(ctx, "current_version", fmt.Appendf(nil, "%v", max_version)); err != nil {
			return err
		}
		for _, e := range processedEntries {
			if err := entryRepoTx.CreateOrUpdate(ctx, e); err != nil {
				return err
			}
		}
		for _, e := range newEntries {
			if err := entryRepoTx.CreateOrUpdate(ctx, e); err != nil {
				return err
			}
		}
		for _, f := range newFiles {
			if err := fileRepoTx.CreateOrUpdate(ctx, f); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error tx: %w", err)
	}
	return nil
}

// GetPresignedGetUrl fetches a presigned GET URL for the entry's file.
func (s *entryService) GetPresignedGetUrl(ctx context.Context, id string) (string, error) {
	url, err := s.client.GetPresignedGetURL(ctx, id)
	if err != nil {
		return "", fmt.Errorf("error get presigned url: %w", err)
	}
	return url, nil
}

// GetFile returns file metadata for the given entry id.
func (s *entryService) GetFile(ctx context.Context, id string) (*models.File, error) {
	return s.getFileRepo(s.db).GetByEntryID(ctx, id)
}
