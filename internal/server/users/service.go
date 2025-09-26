package users

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
	"github.com/dmitrijs2005/gophkeeper/internal/server/auth"
	"github.com/dmitrijs2005/gophkeeper/internal/server/config"
	"github.com/dmitrijs2005/gophkeeper/internal/server/refreshtokens"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type Service struct {
	repo                         Repository
	refreshTokenRepo             refreshtokens.Repository
	jwtSecret                    []byte
	accessTokenValidityDuration  time.Duration
	refreshTokenValidityDuration time.Duration
}

func NewService(repo Repository, refreshTokenRepo refreshtokens.Repository, cfg *config.Config) *Service {
	return &Service{
		repo:                         repo,
		refreshTokenRepo:             refreshTokenRepo,
		jwtSecret:                    []byte(cfg.SecretKey),
		accessTokenValidityDuration:  cfg.AccessTokenValidityDuration,
		refreshTokenValidityDuration: cfg.RefreshTokenValidityDuration,
	}
}

func (s *Service) Register(ctx context.Context, username string, salt, verifier []byte) (*User, error) {

	user := &User{
		UserName: username,
		Salt:     salt,
		Verifier: verifier,
	}

	user, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %v", err)
	}

	return user, nil
}

func (s *Service) getRandomSalt() []byte {
	return common.GenerateRandByteArray(32)
}

func (s *Service) GetSalt(ctx context.Context, userName string) ([]byte, error) {

	user, err := s.repo.GetUserByLogin(ctx, userName)
	if err != nil {
		if errors.Is(err, common.ErrorNotFound) {
			// if user not found, returning random salt
			return s.getRandomSalt(), nil
		}
		return nil, common.ErrorInternal
	}

	return user.Salt, nil
}

func (s *Service) generateAccessToken(ctx context.Context, user *User) (string, error) {
	token, err := auth.GenerateToken(user.ID, s.jwtSecret, s.accessTokenValidityDuration)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *Service) generateRefreshToken() (string, error) {
	token, err := common.MakeRandHexString(32)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *Service) checkVerifier(verifier []byte, verifierCandidate []byte) bool {
	return subtle.ConstantTimeCompare(verifier, verifierCandidate) == 1
}

func (s *Service) Login(ctx context.Context, userName string, verifierCandidate []byte) (*TokenPair, error) {

	user, err := s.repo.GetUserByLogin(ctx, userName)
	if err != nil {
		if errors.Is(err, common.ErrorNotFound) {
			return nil, common.ErrorUnauthorized
		}
		return nil, common.ErrorInternal
	}

	if !s.checkVerifier(user.Verifier, verifierCandidate) {
		return nil, common.ErrorUnauthorized
	}

	accessToken, err := s.generateAccessToken(ctx, user)
	if err != nil {
		return nil, common.ErrorInternal
	}

	refreshtoken, err := s.generateRefreshToken()
	if err != nil {
		return nil, common.ErrorInternal
	}

	err = s.refreshTokenRepo.Create(ctx, user.ID, refreshtoken, s.refreshTokenValidityDuration)
	if err != nil {
		return nil, common.ErrorInternal
	}

	return &TokenPair{AccessToken: accessToken, RefreshToken: refreshtoken}, nil
}
