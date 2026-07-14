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
	ErrInvalidSearch      = errors.New("invalid search options")
)

type Repository interface {
	// queries
	GetByUUID(ctx context.Context, uuid string) (Stack, error)
	List(ctx context.Context, userExternalID string) ([]Stack, error)
	ListPublic(ctx context.Context, userExternalID string) ([]Stack, error)
	Search(ctx context.Context, options SearchOptions) (SearchResult, error)
	CountPublic(ctx context.Context) (int, error)

	// commands
	Create(ctx context.Context, stack Stack) error
	Update(ctx context.Context, modified Stack) (Stack, error)
	Delete(ctx context.Context, uuid string) error
	AddPaper(ctx context.Context, stackUUID string, paper paperDomain.Paper) error
	RemovePaper(ctx context.Context, stackUUID string, paperUUID string) error
}

type SearchOptions struct {
	Query    string
	SortBy   string
	Desc     bool
	Page     int
	PageSize int
}

type SearchResult struct {
	Items    []Stack
	Total    int
	Page     int
	PageSize int
	HasNext  bool
}
