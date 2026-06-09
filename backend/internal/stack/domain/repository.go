package domain

import (
	"context"
	"errors"

	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

var (
	ErrStackAlreadyExists = errors.New("stack already exists")
	ErrStackNotFound      = errors.New("stack not found")
	ErrInvalidStack       = errors.New("invalid stack")
)

type Repository interface {
	// queries
	GetByUUID(ctx context.Context, uuid string) (Stack, error)
	List(ctx context.Context, userExternalID string) ([]Stack, error)
	ListPublic(ctx context.Context, userExternalID string) ([]Stack, error)
	ListAllPublic(ctx context.Context) ([]Stack, error)

	// commands
	Create(ctx context.Context, stack Stack) error
	Update(ctx context.Context, modified Stack) (Stack, error)
	Delete(ctx context.Context, uuid string) error
	AddPaper(ctx context.Context, stackUUID string, paper paperDomain.Paper) error
	RemovePaper(ctx context.Context, stackUUID string, paperUUID string) error
}
