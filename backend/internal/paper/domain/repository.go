package domain

import (
	"context"
	"errors"
)

var (
	ErrPaperNotFound      = errors.New("paper not found")
	ErrPaperAlreadyExists = errors.New("paper already exists")
	ErrInvalidPaper       = errors.New("invalid paper")
	ErrDOIMismatch        = errors.New("paper DOI does not match resource DOI")
	ErrInvalidSearch      = errors.New("invalid search options")
)

type Repository interface {
	// queries
	GetByUUID(ctx context.Context, uuid string) (Paper, error)
	GetByDOI(ctx context.Context, doi string) (Paper, error)
	Search(ctx context.Context, opts SearchOptions) (SearchResult, error)
	List(ctx context.Context) ([]Paper, error)

	// commands
	Save(ctx context.Context, paper Paper) error
	Update(ctx context.Context, doi string, paper Paper) error
	Delete(ctx context.Context, uuid string) error
}

type SearchOptions struct {
	Query    string
	SortBy   string
	Desc     bool
	Page     int
	PageSize int
}

type SearchResult struct {
	Items    []Paper
	Total    int
	Page     int
	PageSize int
	HasNext  bool
}
