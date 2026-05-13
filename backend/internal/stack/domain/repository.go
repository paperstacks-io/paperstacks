package domain

import (
	"context"

	domainUser "github.com/paperstacks.io/paperstacks/internal/auth/domain"
)

type Repository interface {
	// queries
	Create(ctx context.Context, stack Stack) error
	Update(ctx context.Context, modified Stack) (Stack, error)
	Delete(ctx context.Context, uuid string) error

	//commands
	List(ctx context.Context, user domainUser.User) ([]Stack, error)
}
