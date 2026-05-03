package domain

import (
	"context"
	"errors"
)

var (
	ErrStackAlreadyExists = errors.New("stack already exists")
	ErrStackNotFound      = errors.New("stack not found")
)

type Visibility string

const (
	VisibilityPrivate Visibility = "private"
	VisibilityPublic  Visibility = "public"
)

type Repository interface {
	// TODO: Add more queries and commands as needed
	// queries
	List(ctx context.Context) ([]Stack, error)
	ListByOwner(ctx context.Context, owner string) ([]Stack, error)
	ListPrivateByOwner(ctx context.Context, owner string) ([]Stack, error)
	ListPublicByOwner(ctx context.Context, owner string) ([]Stack, error)

	// commands
	Save(ctx context.Context, stack Stack) error
	Delete(ctx context.Context, stack Stack) error
	SetVisibility(ctx context.Context, stack Stack, visibility Visibility) error
}

// TODO: Add SearchOptions and SearchResult types for stack searching
