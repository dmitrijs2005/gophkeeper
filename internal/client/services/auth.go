package services

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"

	"github.com/dmitrijs2005/gophkeeper/internal/client/client"
	"github.com/dmitrijs2005/gophkeeper/internal/client/repositories/metadata"
	"github.com/dmitrijs2005/gophkeeper/internal/client/utils"
	"github.com/dmitrijs2005/gophkeeper/internal/common"
	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
)

type AuthService interface {
	OfflineLogin(ctx context.Context, username string, password []byte) ([]byte, error)
	OnlineLogin(ctx context.Context, username string, password []byte) ([]byte, error)
	Register(ctx context.Context, username string, password []byte) error
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
}

type authService struct {
	client client.Client
	db     *sql.DB
}

func NewAuthService(client client.Client, db *sql.DB) AuthService {
	return &authService{client: client, db: db}
}

func (a *authService) getMetadataRepo() metadata.Repository {
	return metadata.NewSQLiteRepository(a.db)
}

func (a *authService) OfflineLogin(ctx context.Context, username string, password []byte) ([]byte, error) {

	metadataRepo := a.getMetadataRepo()

	savedUsername, err := metadataRepo.Get(ctx, "username")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, client.ErrLocalDataNotAvailable
		}
	}

	if string(savedUsername) != username {
		return nil, client.ErrUnauthorized
	}

	savedSalt, err := metadataRepo.Get(ctx, "salt")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, client.ErrLocalDataNotAvailable
		}
	}

	savedVerifier, err := metadataRepo.Get(ctx, "verifier")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, client.ErrLocalDataNotAvailable
		}
	}

	// if salt and verifier are avalale, trying to log in locally
	masterKeyCandidate := utils.DeriveMasterKey(password, savedSalt)
	verifierCandidate := utils.MakeVerifier(masterKeyCandidate)

	if subtle.ConstantTimeCompare(savedVerifier, verifierCandidate) == 0 {
		return nil, client.ErrUnauthorized
	}

	return masterKeyCandidate, nil

}

func (a *authService) OnlineLogin(ctx context.Context, userName string, password []byte) ([]byte, error) {

	salt, err := a.client.GetSalt(ctx, userName)
	if err != nil {
		return nil, err
	}

	masterKeyCandidate := utils.DeriveMasterKey(password, salt)
	verifierCandidate := utils.MakeVerifier(masterKeyCandidate)

	err = a.client.Login(ctx, userName, verifierCandidate)

	if err != nil {
		return nil, err
	}

	if err := a.saveOfflineData(ctx, userName, salt, verifierCandidate); err != nil {
		return nil, err
	}

	return masterKeyCandidate, nil
}

func (a *authService) saveOfflineData(ctx context.Context,
	userName string,
	salt []byte,
	varifier []byte) error {

	metadataRepo := a.getMetadataRepo()

	return dbx.WithTx(ctx, a.db, nil, func(ctx context.Context, tx dbx.DBTX) error {
		if err := metadataRepo.Set(ctx, "username", []byte(userName)); err != nil {
			return err
		}

		if err := metadataRepo.Set(ctx, "salt", salt); err != nil {
			return err
		}

		if err := metadataRepo.Set(ctx, "verifier", varifier); err != nil {
			return err
		}
		return nil
	})

}

func (a *authService) Register(ctx context.Context, username string, password []byte) error {
	salt := common.GenerateRandByteArray(32)
	key := utils.DeriveMasterKey(password, salt)
	verifier := utils.MakeVerifier(key)

	err := a.client.Register(ctx, username, salt, verifier)

	if err != nil {
		return err
	}

	return nil

}

func (a *authService) Ping(ctx context.Context) error {
	return a.client.Ping(ctx)
}

func (a *authService) Close(ctx context.Context) error {
	return a.client.Close()
}
