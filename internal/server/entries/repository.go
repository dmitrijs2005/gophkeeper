package entries

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, entry *Entry) (*Entry, error)
}
