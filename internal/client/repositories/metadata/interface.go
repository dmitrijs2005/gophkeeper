package metadata

import (
	"context"
)

type Repository interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context) (map[string][]byte, error)
	Clear(ctx context.Context) error
}
