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
)

type AuthService interface {
	OfflineLogin(ctx context.Context, username string, password []byte) ([]byte, error)
	OnlineLogin(ctx context.Context, username string, password []byte) ([]byte, error)
	Register(ctx context.Context, username string, password []byte) error
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
}

type authService struct {
	client       client.Client
	metadataRepo metadata.Repository
}

func NewAuthService(client client.Client, metadataRepo metadata.Repository) AuthService {
	return &authService{client: client, metadataRepo: metadataRepo}
}

func (a *authService) getMetadataKey(ctx context.Context, key string) ([]byte, error) {
	return a.metadataRepo.Get(ctx, key)
}

func (a *authService) OfflineLogin(ctx context.Context, username string, password []byte) ([]byte, error) {

	savedUsername, err := a.getMetadataKey(ctx, "username")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, client.ErrLocalDataNotAvailable
		}
	}

	if string(savedUsername) != username {
		return nil, client.ErrUnauthorized
	}

	savedSalt, err := a.getMetadataKey(ctx, "salt")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, client.ErrLocalDataNotAvailable
		}
	}

	savedVerifier, err := a.getMetadataKey(ctx, "verifier")
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

	return masterKeyCandidate, nil
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
