package users

import (
	"context"
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/server/auth"
	"github.com/dmitrijs2005/gophkeeper/internal/server/config"
	"github.com/dmitrijs2005/gophkeeper/internal/server/refreshtokens"
	"github.com/dmitrijs2005/gophkeeper/internal/shared"
)

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

func (s *Service) GetSalt(ctx context.Context, userName string) ([]byte, error) {

	user, err := s.repo.GetUserByLogin(ctx, userName)
	if err != nil {
		return nil, fmt.Errorf("error selecting user0: %v", err)
	}

	return user.Salt, nil
}

func (s *Service) Login(ctx context.Context, userName string, verifierCandidate []byte) (string, string, error) {

	user, err := s.repo.GetUserByLogin(ctx, userName)
	if err != nil {
		return "", "", fmt.Errorf("error selecting user1: %v", err)
	}

	if subtle.ConstantTimeCompare(user.Verifier, verifierCandidate) != 1 {

		return "", "", fmt.Errorf("error selecting user2: %v", err)
	}

	// everything is ok, username + verifier match
	accessToken, err := auth.GenerateToken(user.ID, s.jwtSecret, s.accessTokenValidityDuration)
	if err != nil {
		return "", "", fmt.Errorf("error selecting user3: %v", err)
	}

	refreshtoken, err := shared.MakeRandHexString(32)
	if err != nil {
		return "", "", fmt.Errorf("error selecting user4: %v", err)
	}

	err = s.refreshTokenRepo.Create(ctx, user.ID, refreshtoken, s.refreshTokenValidityDuration)
	if err != nil {
		return "", "", fmt.Errorf("error selecting user5: %v", err)
	}

	return accessToken, refreshtoken, nil
}
