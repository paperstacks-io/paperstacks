// Package application provides user application services.
package application

import (
	"context"
	"strings"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/user/domain"
)

type UserService struct {
	repo domain.Repository
}

// NewUserService creates a user service backed by repo.
func NewUserService(repo domain.Repository) *UserService {
	return &UserService{repo: repo}
}

// List returns all users.
func (s *UserService) List(ctx context.Context) ([]domain.User, error) {
	return s.repo.List(ctx)
}

// GetByExternalID returns a user by external authentication ID.
func (s *UserService) GetByExternalID(ctx context.Context, externalID string) (domain.User, error) {
	return s.repo.GetByExternalID(ctx, strings.TrimSpace(externalID))
}

// GetByEmail returns a user by email address.
func (s *UserService) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	return s.repo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
}

// CreateIfNotExist creates a user or returns the existing user.
func (s *UserService) CreateIfNotExist(ctx context.Context, externalID, email string) (domain.User, error) {
	user := domain.NewUser(externalID, email)
	if err := user.Validate(); err != nil {
		return domain.User{}, err
	}

	return s.repo.SaveIfNotExist(ctx, user)
}

// Update replaces an existing user after validating identity and fields.
func (s *UserService) Update(ctx context.Context, externalID string, user domain.User) error {
	externalID = strings.TrimSpace(externalID)

	if user.ExternalID != externalID {
		return domain.ErrExternalIDMismatch
	}

	if err := user.Validate(); err != nil {
		return err
	}

	user.UpdatedAt = time.Now()

	return s.repo.Update(ctx, externalID, user)
}

// Delete removes a user by external authentication ID.
func (s *UserService) Delete(ctx context.Context, externalID string) error {
	return s.repo.Delete(ctx, strings.TrimSpace(externalID))
}
