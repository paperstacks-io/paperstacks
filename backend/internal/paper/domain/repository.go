package domain

import (
	"context"
)

// Repository https://threedots.tech/post/repository-pattern-in-go/?utm_source=github.com
type Repository interface {
	Create(ctx context.Context, paper Paper) error
	ReadAll(_ context.Context) ([]Paper, error)
	Read(_ context.Context, id string) (Paper, error)
	Update(_ context.Context, id string, paper Paper) error
	Delete(_ context.Context, id string) error
}
