package application

import (
	"context"

	userDomain "github.com/paperstacks.io/paperstacks/internal/auth/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

type StackService struct {
}

func NewStackService() *StackService {
	return &StackService{}
}

// Add creates and stores a new stack
// It returns an error if the stack could not be created
// (e.g. the stack name already exists in the user's stack list).
func (s *StackService) Create(ctx context.Context, stack domain.Stack) error {
	panic("Not implemented")
}

// Update the specified stack.
// It returns the updated stack and an error if the operation fails.
//
// An error is returned if, e.g.:
//   - the stack does not exist
//   - the paper does not exist
//   - the paper is not part of the stack
//   - the stack reached the maximum number of papers
//   - the updated stack name already exists in the user's stack list
func (s *StackService) Update(ctx context.Context, modified domain.Stack) (domain.Stack, error) {
	panic("Not implemented")
}

// Delete removes the specified stack.
// It returns an error if the stack does not exist
func (s *StackService) Delete(ctx context.Context, uuid string) error {
	panic("Not implemented")
}

// List returns all stacks of the specified user.
// This includes public and private stacks.
//
// It returns an error if the stacks could not be loaded.
func (s *StackService) List(ctx context.Context, user userDomain.User) ([]domain.Stack, error) {
	panic("Not implemented")
}
