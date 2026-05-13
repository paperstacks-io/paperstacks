package domain

import (
	"context"
	"strings"
)

type UserService struct {
	repo Repository
}

func NewUserService(repo Repository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) List(ctx context.Context) ([]User, error) {
	return s.repo.List(ctx)
}

func (s *UserService) GetByExternalID(ctx context.Context, externalID string) (User, error) {
	return s.repo.GetByExternalID(ctx, strings.TrimSpace(externalID))
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (User, error) {
	return s.repo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
}

func (s *UserService) Create(ctx context.Context, user User) (User, error) {
	user = user.Normalize()
	if err := user.Validate(); err != nil {
		return User{}, err
	}

	return s.repo.Save(ctx, user)
}

func (s *UserService) Update(ctx context.Context, externalID string, user User) error {
	externalID = strings.TrimSpace(externalID)
	user = user.Normalize()

	if user.ExternalID != externalID {
		return ErrExternalIDMismatch
	}

	if err := user.Validate(); err != nil {
		return err
	}

	return s.repo.Update(ctx, externalID, user)
}

func (s *UserService) Delete(ctx context.Context, externalID string) error {
	return s.repo.Delete(ctx, strings.TrimSpace(externalID))
}
