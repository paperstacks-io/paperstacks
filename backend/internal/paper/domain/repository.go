package domain

import (
	"context"
	"errors"
)

var (
	ErrPaperNotFound      = errors.New("paper not found")
	ErrPaperAlreadyExists = errors.New("paper already exists")
)

type Repository interface {
	// queries
	GetByDOI(ctx context.Context, doi string) (Paper, error)
	List(ctx context.Context) ([]Paper, error)
	// commands
	Save(ctx context.Context, paper Paper) error
	Update(ctx context.Context, doi string, paper Paper) error
	Delete(ctx context.Context, doi string) error
}
