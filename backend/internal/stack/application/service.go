package application

import (
	"context"

	userDomain "github.com/paperstacks.io/paperstacks/internal/auth/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

type StackService struct{}

func NewStackService() *StackService {
	return &StackService{}
}

func (s *StackService) Create(ctx context.Context, stack domain.Stack) error {
	panic("Not implemented")
}

func (s *StackService) Update(ctx context.Context, modified domain.Stack) (domain.Stack, error) {
	panic("Not implemented")
}

// Delete removes the specified stack.
// It returns an error if the stack does not exist
func (s *StackService) Delete(ctx context.Context, uuid string) error {
	panic("Not implemented")
}

// List returns all stacks of a given user.
// This includes public and private stacks.
//
// It returns an error if the stacks could not be loaded.
func (s *StackService) List(ctx context.Context, user userDomain.User) ([]domain.Stack, error) {
	panic("Not implemented")
}
