// Package services contains server-side business logic. This file implements
// UserService, which handles registration, login, and issuing/refreshing JWTs
// plus server-stored refresh tokens.
package services

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
	"github.com/dmitrijs2005/gophkeeper/internal/server/auth"
	"github.com/dmitrijs2005/gophkeeper/internal/server/config"
	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/repomanager"
)

// TokenPair bundles a short-lived access token and a long-lived refresh token.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// UserService provides authentication-related operations:
// - Register: create users
// - Login: verify credentials and mint tokens
// - RefreshToken: rotate refresh tokens and mint new access tokens
type UserService struct {
	db                           *sql.DB
	repomanager                  repomanager.RepositoryManager
	jwtSecret                    []byte
	accessTokenValidityDuration  time.Duration
	refreshTokenValidityDuration time.Duration
}

// NewUserService constructs a UserService using repositories and server config.
func NewUserService(db *sql.DB, m repomanager.RepositoryManager, cfg *config.Config) *UserService {
	return &UserService{
		db:                           db,
		repomanager:                  m,
		jwtSecret:                    []byte(cfg.SecretKey),
		accessTokenValidityDuration:  cfg.AccessTokenValidityDuration,
		refreshTokenValidityDuration: cfg.RefreshTokenValidityDuration,
	}
}

// RefreshToken validates a refresh token, rotates it transactionally, and
// returns a fresh TokenPair. Expired tokens yield ErrRefreshTokenExpired.
func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	repo := s.repomanager.RefreshTokens(s.db)

	token, err := repo.Find(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("error searching refresh token: %v", err)
	}
	if token.Expires.Before(time.Now()) {
		return nil, common.ErrRefreshTokenExpired
	}

	var pair *TokenPair
	if err := dbx.WithTx(ctx, s.db, nil, func(ctx context.Context, tx dbx.DBTX) error {
		repoTx := s.repomanager.RefreshTokens(tx)
		if err := repoTx.Delete(ctx, refreshToken); err != nil {
			return fmt.Errorf("error deleting refresh token: %v", err)
		}
		var genErr error
		pair, genErr = s.generateTokenPair(ctx, token.UserID, tx)
		return genErr
	}); err != nil {
		return nil, err
	}
	return pair, nil
}

// Register creates a new user with the given username, salt, and verifier.
func (s *UserService) Register(ctx context.Context, username string, salt, verifier []byte) (*models.User, error) {
	user := &models.User{UserName: username, Salt: salt, Verifier: verifier}
	repo := s.repomanager.Users(s.db)
	u, err := repo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %v", err)
	}
	return u, nil
}

// GetSalt returns the user's stored salt or a random salt if the user is absent,
// to avoid leaking existence through timing.
func (s *UserService) GetSalt(ctx context.Context, userName string) ([]byte, error) {
	repo := s.repomanager.Users(s.db)
	user, err := repo.GetUserByLogin(ctx, userName)
	if err != nil {
		if errors.Is(err, common.ErrorNotFound) {
			return s.getRandomSalt(), nil
		}
		return nil, common.ErrorInternal
	}
	return user.Salt, nil
}

// Login verifies the provided verifierCandidate against the stored verifier and,
// on success, returns a new TokenPair.
func (s *UserService) Login(ctx context.Context, userName string, verifierCandidate []byte) (*TokenPair, error) {
	repo := s.repomanager.Users(s.db)
	user, err := repo.GetUserByLogin(ctx, userName)
	if err != nil {
		if errors.Is(err, common.ErrorNotFound) {
			return nil, common.ErrorUnauthorized
		}
		return nil, common.ErrorInternal
	}
	if !s.checkVerifier(user.Verifier, verifierCandidate) {
		return nil, common.ErrorUnauthorized
	}
	return s.generateTokenPair(ctx, user.ID, s.db)
}

// --- helpers below ---

func (s *UserService) getRandomSalt() []byte { return common.GenerateRandByteArray(32) }

func (s *UserService) generateAccessToken(userID string) (string, error) {
	return auth.GenerateToken(userID, s.jwtSecret, s.accessTokenValidityDuration)
}

func (s *UserService) generateRefreshToken() (string, error) {
	return common.MakeRandHexString(32)
}

func (s *UserService) checkVerifier(verifier []byte, candidate []byte) bool {
	return subtle.ConstantTimeCompare(verifier, candidate) == 1
}

func (s *UserService) generateTokenPair(ctx context.Context, userID string, tx dbx.DBTX) (*TokenPair, error) {
	access, err := s.generateAccessToken(userID)
	if err != nil {
		return nil, common.ErrorInternal
	}
	refresh, err := s.generateRefreshToken()
	if err != nil {
		return nil, common.ErrorInternal
	}
	refreshRepo := s.repomanager.RefreshTokens(tx)
	if err := refreshRepo.Create(ctx, userID, refresh, s.refreshTokenValidityDuration); err != nil {
		return nil, common.ErrorInternal
	}
	return &TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}
