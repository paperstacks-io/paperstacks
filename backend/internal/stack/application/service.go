package application

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

type StackService struct {
	repo domain.Repository
}

func NewStackService(repo domain.Repository) *StackService {
	return &StackService{
		repo: repo,
	}
}

func (s *StackService) Create(ctx context.Context, stack domain.Stack) error {
	if stack.Owner.ExternalID == "" {
		return userDomain.ErrInvalidOwner
	}

	err := uuid.Validate(stack.UUID)
	if err != nil {
		stack.UUID = uuid.NewString()
	}

	return s.repo.Create(ctx, stack)
}

func (s *StackService) Update(ctx context.Context, modified domain.Stack) (domain.Stack, error) {
	if modified.Owner.ExternalID == "" {
		return domain.Stack{}, userDomain.ErrInvalidOwner
	}

	modified.UpdatedAt = time.Now()

	return s.Update(ctx, modified)
}

// Delete removes the specified stack.
// It returns an error if the stack does not exist
func (s *StackService) Delete(ctx context.Context, uuid string) error {
	return s.repo.Delete(ctx, strings.TrimSpace(uuid))
}

// List returns all stacks of a given user.
// This includes public and private stacks.
//
// It returns an error if the stacks could not be loaded.
func (s *StackService) List(ctx context.Context, user userDomain.User) ([]domain.Stack, error) {
	if user.ExternalID == "" {
		return []domain.Stack{}, userDomain.ErrInvalidOwner
	}

	return s.repo.List(ctx, user)
}
