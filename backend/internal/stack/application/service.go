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

// Create validates and stores a new stack.
// It initializes missing timestamps and generates a UUID if necessary.
//
// It returns an error if the stack is invalid or could not be stored.
func (s *StackService) Create(ctx context.Context, stack domain.Stack) error {
	if err := stack.Validate(); err != nil {
		return err
	}

	if stack.CreatedAt.IsZero() || stack.UpdatedAt.IsZero() {
		now := time.Now()
		stack.CreatedAt = now
		stack.UpdatedAt = now
	}

	err := uuid.Validate(stack.UUID)
	if err != nil {
		stack.UUID = uuid.NewString()
	}

	return s.repo.Create(ctx, stack)
}

// Update validates and updates an existing stack.
// It refreshes the UpdatedAt timestamp before storing the changes.
//
// It returns an error if the stack is invalid or could not be stored.
func (s *StackService) Update(ctx context.Context, modified domain.Stack) (domain.Stack, error) {
	if err := modified.Validate(); err != nil {
		return domain.Stack{}, err
	}

	modified.UpdatedAt = time.Now()

	return s.repo.Update(ctx, modified)
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
	if err := user.Validate(); err != nil {
		return nil, err
	}

	return s.repo.List(ctx, user)
}
