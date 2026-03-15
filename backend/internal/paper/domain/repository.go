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
)

type Repository interface {
	// queries
	GetByDOI(ctx context.Context, doi string) (Paper, error)
	Search(ctx context.Context, title, keyword string) ([]Paper, error)
	List(ctx context.Context) ([]Paper, error)

	// commands
	Save(ctx context.Context, paper Paper) error
	Update(ctx context.Context, doi string, paper Paper) error
	Delete(ctx context.Context, doi string) error
}
