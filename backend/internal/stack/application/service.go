package application

import (
	userDomain "github.com/paperstacks.io/paperstacks/internal/auth/domain"
	stackDomain "github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

type StackService struct {
}

func NewStackService() *StackService {
	return &StackService{}
}

// Add creates and stores a new stack // It returns an error if the stack could not be created
// (e.g. the stack name already exists in the user's stack list).
func (s *StackService) Create(name string, user userDomain.User) error {
	return nil
}

// Update the specified stack.
// It returns the updated stack and an error if the operation fails.
//
// An error is returned if:
//   - the stack does not exist
//   - the paper does not exist
//   - the paper is not part of the stack
//   - the stack reached the maximum number of papers
func (s *StackService) Update(user userDomain.User, stackUUID string, paperUUID string) (stackDomain.Stack, error) {
	return stackDomain.Stack{}, nil
}

// Delete removes the specified stack from User.
// It returns an error if the stack does not exist
func (s *StackService) Delete(user userDomain.User, stackUUID string) error {
	return nil
}

// List returns all stacks of the specified user.
// This includes public and private stacks.
//
// It returns an error if the stacks could not be loaded.
func (s *StackService) List(user userDomain.User) ([]stackDomain.Stack, error) {
	return []stackDomain.Stack{}, nil
}
