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

func NewUserService(repo domain.Repository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) List(ctx context.Context) ([]domain.User, error) {
	return s.repo.List(ctx)
}

func (s *UserService) GetByExternalID(ctx context.Context, externalID string) (domain.User, error) {
	return s.repo.GetByExternalID(ctx, strings.TrimSpace(externalID))
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	return s.repo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
}

func (s *UserService) CreateIfNotExist(ctx context.Context, externalID, email string) (domain.User, error) {
	user := domain.NewUser(externalID, email)
	if err := user.Validate(); err != nil {
		return domain.User{}, err
	}

	return s.repo.SaveIfNotExist(ctx, user)
}

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

func (s *UserService) Delete(ctx context.Context, externalID string) error {
	return s.repo.Delete(ctx, strings.TrimSpace(externalID))
}
