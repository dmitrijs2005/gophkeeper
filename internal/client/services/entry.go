package services

import (
	"context"
	"fmt"
	"log"

	"github.com/dmitrijs2005/gophkeeper/internal/client/client"
	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
	"github.com/dmitrijs2005/gophkeeper/internal/client/repositories/entries"
	"github.com/dmitrijs2005/gophkeeper/internal/client/repositories/metadata"
	"github.com/dmitrijs2005/gophkeeper/internal/client/utils"
	"github.com/google/uuid"
)

type EntryService interface {
	Sync(ctx context.Context) error
	List(ctx context.Context, masterKey []byte) ([]models.ViewOverview, error)
	Add(ctx context.Context, envelope models.Envelope, masterKey []byte) error
	DeleteByID(ctx context.Context, id string) error
	Get(ctx context.Context, id string, masterKey []byte) (*models.Envelope, error)
}

type entryService struct {
	client       client.Client
	entryRepo    entries.Repository
	metadataRepo metadata.Repository
}

func NewEntryService(client client.Client, entryRepo entries.Repository) EntryService {
	return &entryService{client: client, entryRepo: entryRepo}
}

func (s *entryService) Add(ctx context.Context, envelope models.Envelope, masterKey []byte) error {

	overview := envelope.Overview()
	oCipherText, oNonce, err := utils.EncryptEntry(overview, masterKey)

	if err != nil {
		return fmt.Errorf("encryption error: %w", err)
	}

	cipherText, nonce, err := utils.EncryptEntry(envelope, masterKey)

	if err != nil {
		return fmt.Errorf("encryption error: %w", err)
	}

	e := &models.Entry{Id: uuid.NewString(),
		Overview:      oCipherText,
		NonceOverview: oNonce,
		Details:       cipherText,
		NonceDetails:  nonce,
	}

	err = s.entryRepo.Insert(ctx, e)
	if err != nil {
		return fmt.Errorf("saving error: %w", err)
	}

	return nil

}

func (s *entryService) List(ctx context.Context, masterKey []byte) ([]models.ViewOverview, error) {

	rows, err := s.entryRepo.GetAll(ctx)
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

	err := s.entryRepo.DeleteByID(ctx, id)
	if err != nil {
		return fmt.Errorf("error deleting entry: %w", err)
	}
	return nil
}

func (s *entryService) Get(ctx context.Context, id string, masterKey []byte) (*models.Envelope, error) {

	entry, err := s.entryRepo.GetByID(ctx, id)
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

	entries, err := s.entryRepo.GetAllPending(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving entries: %w", err)
	}

	err = s.client.Sync(ctx, entries, int64(1))

	return err
}
