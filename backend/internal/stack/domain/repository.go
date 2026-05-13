package domain

import (
	"context"
	"errors"

	domainUser "github.com/paperstacks.io/paperstacks/internal/auth/domain"
)

var (
	ErrStackAlreadyExists = errors.New("stack already exists")
)

type Repository interface {
	// queries
	List(ctx context.Context, user domainUser.User) ([]Stack, error)

	//commands
	Create(ctx context.Context, stack Stack) error
	Update(ctx context.Context, modified Stack) (Stack, error)
	Delete(ctx context.Context, uuid string) error
}
