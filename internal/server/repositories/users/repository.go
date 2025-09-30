package users

import (
	"context"

	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
)

type Repository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
}
