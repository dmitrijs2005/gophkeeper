package users

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetUserByLogin(ctx context.Context, login string) (*User, error)
}
