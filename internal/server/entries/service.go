package entries

import (
	"context"
	"fmt"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, userID string, title string, entryType string, cypherText []byte, nonce []byte) (*Entry, error) {

	entry := &Entry{
		UserID:        userID,
		Title:         title,
		Type:          entryType,
		EncryptedData: cypherText,
		Nonce:         nonce,
	}

	user, err := s.repo.Create(ctx, entry)
	if err != nil {
		return nil, fmt.Errorf("error creating entry: %v", err)
	}

	return user, nil
}
