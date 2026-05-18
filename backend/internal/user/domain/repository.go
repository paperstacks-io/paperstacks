package domain

import (
	"context"
	"errors"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidUser        = errors.New("invalid user")
	ErrExternalIDMismatch = errors.New("user external ID does not match")
)

type Repository interface {
	// queries
	GetByExternalID(ctx context.Context, externalID string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	List(ctx context.Context) ([]User, error)

	// commands
	SaveIfNotExist(ctx context.Context, user User) (User, error)
	Update(ctx context.Context, externalID string, user User) error
	Delete(ctx context.Context, externalID string) error
}
