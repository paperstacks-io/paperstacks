package domain

import (
	"context"
)

// TODO: Define specific errors for stack operations, such as ErrStackNotFound, ErrStackAlreadyExists, etc.

type Visibility string

const (
	VisibilityPrivate Visibility = "private"
	VisibilityPublic  Visibility = "public"
)

type Repository interface {
	// TODO: Add more queries and commands as needed
	// queries
	ListPublic(ctx context.Context) ([]Stack, error)
	GetPublicByOwner(ctx context.Context, owner string) (Stack, error)

	// commands
	Save(ctx context.Context, stack Stack) error
	Delete(ctx context.Context, uuid string) error
	SetVisibility(ctxt context.Context, uuid string, visibility bool) error
}

// TODO: Add SearchOptions and SearchResult types for stack searching
