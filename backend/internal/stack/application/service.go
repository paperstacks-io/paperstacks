// Package application provides stack application services.
package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

const (
	defaultSearchPage     = 1
	defaultSearchPageSize = 10
	maxSearchPageSize     = 100
)

type PaperGetter interface {
	GetByUUID(ctx context.Context, uuid string) (paperDomain.Paper, error)
}

type StackService struct {
	repo        domain.Repository
	paperGetter PaperGetter
}

func NewStackService(repo domain.Repository, paperGetter PaperGetter) *StackService {
	return &StackService{
		repo:        repo,
		paperGetter: paperGetter,
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

// GetByUUID returns the stack with the specified UUID.
//
// It returns an error if the stack does not exist.
func (s *StackService) GetByUUID(ctx context.Context, uuid string) (domain.Stack, error) {
	return s.repo.GetByUUID(ctx, strings.TrimSpace(uuid))
}

// List returns all stacks of a given user.
// This includes public and private stacks.
//
// It returns an error if the stacks could not be loaded.
func (s *StackService) List(ctx context.Context, userExternalID string) ([]domain.Stack, error) {
	if userExternalID == "" {
		return nil, errors.New("invalid user externalID")
	}

	return s.repo.List(ctx, userExternalID)
}

// ListPublic returns public stacks of a given user.
//
// It returns an error if the stacks could not be loaded.
func (s *StackService) ListPublic(ctx context.Context, userExternalID string) ([]domain.Stack, error) {
	if userExternalID == "" {
		return nil, errors.New("invalid user externalID")
	}

	return s.repo.ListPublic(ctx, userExternalID)
}

func (s StackService) CountPublic(ctx context.Context) (int, error) {
	return s.repo.CountPublic(ctx)
}

// AddPaper adds a paper to the specified stack.
//
// If the paper is already assigned to the stack, no changes are made.
// It return an error if the stack does not exist
func (s *StackService) AddPaper(ctx context.Context, stackUUID string, paperUUID string) error {
	paper, err := s.paperGetter.GetByUUID(ctx, paperUUID)
	if err != nil {
		return err
	}

	return s.repo.AddPaper(ctx, stackUUID, paper)
}

// RemovePaper removes a paper from the specified stack.
//
// If the paper is not assigned to the stack, no changes are made.
// It return an error if the stack does not exist
func (s *StackService) RemovePaper(ctx context.Context, stackUUID string, paperUUID string) error {
	return s.repo.RemovePaper(ctx, strings.TrimSpace(stackUUID), strings.TrimSpace(paperUUID))
}

// Search returns stacks matching the provided search options.
// It normalizes the query and applies default pagination values before
//
// It returns an error if the stacks could not be searched.
func (s *StackService) Search(ctx context.Context, opts domain.SearchOptions) (domain.SearchResult, error) {
	opts.Query = strings.ToLower(strings.TrimSpace(opts.Query))
	opts.SortBy = strings.ToLower(strings.TrimSpace(opts.SortBy))

	opts.Page = max(defaultSearchPage, opts.Page)

	if opts.PageSize <= 1 {
		opts.PageSize = defaultSearchPageSize
	}
	opts.PageSize = min(maxSearchPageSize, opts.PageSize)

	if opts.SortBy != "" && opts.SortBy != "name" && opts.SortBy != "updated_at" {
		return domain.SearchResult{}, domain.ErrInvalidSearch
	}

	return s.repo.Search(ctx, opts)
}
