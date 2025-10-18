// Package services contains application services for the GophKeeper client.
// This file defines the authentication service: online/offline login, register,
// liveness probe, and housekeeping of local (offline) auth metadata.
package services

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dmitrijs2005/gophkeeper/internal/client/client"
	"github.com/dmitrijs2005/gophkeeper/internal/client/repositories/metadata"
	"github.com/dmitrijs2005/gophkeeper/internal/common"
	"github.com/dmitrijs2005/gophkeeper/internal/cryptox"
	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
)

// AuthService defines authentication operations for the CLI.
//
// Contract:
//   - OnlineLogin: authenticate against the server and persist offline auth data.
//   - OfflineLogin: derive and verify credentials against locally cached data.
//   - Register: create a new user on the server.
//   - Ping: check server liveness.
//   - Close: release underlying client resources.
//   - ClearOfflineData: wipe locally cached auth metadata.
//
// All methods must honor context cancellation/timeouts.
type AuthService interface {
	OfflineLogin(ctx context.Context, username string, password []byte) ([]byte, error)
	OnlineLogin(ctx context.Context, username string, password []byte) ([]byte, error)
	Register(ctx context.Context, username string, password []byte) error
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	ClearOfflineData(ctx context.Context) error
}

// authService is the concrete AuthService backed by a remote Client
// and a local SQL database for offline metadata.
type authService struct {
	client client.Client
	db     *sql.DB
}

// NewAuthService constructs an AuthService bound to the given API client and DB.
func NewAuthService(client client.Client, db *sql.DB) AuthService {
	return &authService{client: client, db: db}
}

func (a *authService) getMetadataRepo() metadata.Repository {
	return metadata.NewSQLiteRepository(a.db)
}

// OfflineLogin derives a master key from (password,salt) stored locally
// and verifies it against the locally cached verifier. Returns the master key
// on success. If local data is missing, returns client.ErrLocalDataNotAvailable;
// if verification fails, returns client.ErrUnauthorized.
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

	masterKeyCandidate := cryptox.DeriveMasterKey(password, savedSalt)
	verifierCandidate := cryptox.MakeVerifier(masterKeyCandidate)

	if subtle.ConstantTimeCompare(savedVerifier, verifierCandidate) == 0 {
		return nil, client.ErrUnauthorized
	}
	return masterKeyCandidate, nil
}

// OnlineLogin authenticates against the server, saves offline metadata
// (username, salt, verifier), and returns the derived master key.
func (a *authService) OnlineLogin(ctx context.Context, userName string, password []byte) ([]byte, error) {
	salt, err := a.client.GetSalt(ctx, userName)
	if err != nil {
		return nil, fmt.Errorf("get salt error: %w", err)
	}

	masterKeyCandidate := cryptox.DeriveMasterKey(password, salt)
	verifierCandidate := cryptox.MakeVerifier(masterKeyCandidate)

	if err := a.client.Login(ctx, userName, verifierCandidate); err != nil {
		return nil, fmt.Errorf("login error: %w", err)
	}

	if err := a.saveOfflineData(ctx, userName, salt, verifierCandidate); err != nil {
		return nil, fmt.Errorf("offline data saving error: %w", err)
	}
	return masterKeyCandidate, nil
}

// saveOfflineData persists minimal auth metadata required for offline login:
// username, salt, and verifier, in a single transaction.
func (a *authService) saveOfflineData(ctx context.Context, userName string, salt []byte, varifier []byte) error {
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

// Register creates a new account on the server. It generates a random salt,
// derives a master key from the provided password, computes a verifier,
// and sends salt/verifier to the server.
func (a *authService) Register(ctx context.Context, username string, password []byte) error {
	salt := common.GenerateRandByteArray(32)
	key := cryptox.DeriveMasterKey(password, salt)
	verifier := cryptox.MakeVerifier(key)

	if err := a.client.Register(ctx, username, salt, verifier); err != nil {
		return err
	}
	return nil
}

// Ping proxies a liveness check to the underlying client.
func (a *authService) Ping(ctx context.Context) error {
	return a.client.Ping(ctx)
}

// Close releases resources held by the underlying client.
func (a *authService) Close(ctx context.Context) error {
	return a.client.Close()
}

// ClearOfflineData wipes locally cached auth metadata (e.g., on logout).
func (a *authService) ClearOfflineData(ctx context.Context) error {
	metadataRepo := a.getMetadataRepo()
	return metadataRepo.Clear(ctx)
}
