package domain

import (
	"context"
)

type Repository interface {
	// queries

	// commands
	Save(ctx context.Context, stack Stack) error
}
