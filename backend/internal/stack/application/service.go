package application

import "github.com/paperstacks.io/paperstacks/internal/stack/domain"

type StackService struct {
}

func NewStackService() *StackService {
	return &StackService{}
}

// Add creates and stores a new stack in the in-memory repository.
// It returns an error if the stack could not be created
// (e.g. the stack name already exists in the user's stack list).
func (s *StackService) Add() error {
	return nil
}

// AddPaper adds a paper to the specified stack.
// It returns the updated stack and an error if the operation fails.
//
// An error is returned if:
//   - the stack does not exist
//   - the paper does not exist
//   - the paper is already part of the stack
//   - the stack reached the maximum number of papers
func (s *StackService) AddPaper(stackUUID string, paperUUID string) (domain.Stack, error) {
	return domain.Stack{}, nil
}

// RemovePaper removes a paper from the specified stack.
// It returns the updated stack and an error if the operation fails.
//
// An error is returned if:
//   - the stack does not exist
//   - the paper does not exist
//   - the paper is not part of the stack
func (s *StackService) RemovePaper(stackUUID string, paperUUID string) (domain.Stack, error) {
	return domain.Stack{}, nil
}

// Delete removes the specified stack from the in-memory repository.
// It returns an error if the stack does not exist
// or could not be deleted.
func (s *StackService) Delete(stackUUID string) error {
	return nil
}
