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

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type UserService struct {
	db                           *sql.DB
	repomanager                  repomanager.RepositoryManager
	jwtSecret                    []byte
	accessTokenValidityDuration  time.Duration
	refreshTokenValidityDuration time.Duration
}

func NewUserService(db *sql.DB, m repomanager.RepositoryManager, cfg *config.Config) *UserService {
	return &UserService{
		db:                           db,
		repomanager:                  m,
		jwtSecret:                    []byte(cfg.SecretKey),
		accessTokenValidityDuration:  cfg.AccessTokenValidityDuration,
		refreshTokenValidityDuration: cfg.RefreshTokenValidityDuration,
	}
}

func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {

	repo := s.repomanager.RefreshTokens(s.db)

	token, err := repo.Find(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("error searching refresh token: %v", err)
	}

	if token.Expires.Before(time.Now()) {
		return nil, common.ErrRefreshTokenExpired
	}

	var tokenPair *TokenPair

	err = dbx.WithTx(ctx, s.db, nil, func(ctx context.Context, tx dbx.DBTX) error {
		err = repo.Delete(ctx, refreshToken)
		if err != nil {
			return fmt.Errorf("error deleting refresh token: %v", err)
		}

		tokenPair, err = s.generateTokenPair(ctx, token.UserID)
		if err != nil {
			return fmt.Errorf("error generating token pair: %v", err)
		}

		return err

	})

	if err != nil {
		return nil, err
	}

	return tokenPair, nil

}

func (s *UserService) Register(ctx context.Context, username string, salt, verifier []byte) (*models.User, error) {

	user := &models.User{
		UserName: username,
		Salt:     salt,
		Verifier: verifier,
	}

	repo := s.repomanager.Users(s.db)

	user, err := repo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %v", err)
	}

	return user, nil
}

func (s *UserService) getRandomSalt() []byte {
	return common.GenerateRandByteArray(32)
}

func (s *UserService) GetSalt(ctx context.Context, userName string) ([]byte, error) {

	repo := s.repomanager.Users(s.db)

	user, err := repo.GetUserByLogin(ctx, userName)
	if err != nil {
		if errors.Is(err, common.ErrorNotFound) {
			// if user not found, returning random salt
			return s.getRandomSalt(), nil
		}
		return nil, common.ErrorInternal
	}

	return user.Salt, nil
}

func (s *UserService) generateAccessToken(userID string) (string, error) {
	token, err := auth.GenerateToken(userID, s.jwtSecret, s.accessTokenValidityDuration)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *UserService) generateRefreshToken() (string, error) {
	token, err := common.MakeRandHexString(32)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *UserService) checkVerifier(verifier []byte, verifierCandidate []byte) bool {
	return subtle.ConstantTimeCompare(verifier, verifierCandidate) == 1
}

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

	return s.generateTokenPair(ctx, user.ID)
}

func (s *UserService) generateTokenPair(ctx context.Context, userID string) (*TokenPair, error) {
	accessToken, err := s.generateAccessToken(userID)
	if err != nil {
		return nil, common.ErrorInternal
	}

	refreshtoken, err := s.generateRefreshToken()
	if err != nil {
		return nil, common.ErrorInternal
	}

	refreshTokenRepo := s.repomanager.RefreshTokens(s.db)
	err = refreshTokenRepo.Create(ctx, userID, refreshtoken, s.refreshTokenValidityDuration)
	if err != nil {
		return nil, common.ErrorInternal
	}

	return &TokenPair{AccessToken: accessToken, RefreshToken: refreshtoken}, nil
}
