// Package application provides stack application services.
package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paperstacks.io/paperstacks/internal/paper/bibliography"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

const (
	defaultSearchPage     = 1
	defaultSearchPageSize = 10
	maxSearchPageSize     = 100
	defaultStackName      = "Default"
)

type PaperGetter interface {
	GetByUUID(ctx context.Context, uuid string) (paperDomain.Paper, error)
	GetByDOI(ctx context.Context, doi string) (paperDomain.Paper, error)
	Create(ctx context.Context, paper paperDomain.Paper) (paperDomain.Paper, error)
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

type ImportResult struct {
	AlreadyInStack       []paperDomain.Paper
	ExistingPaperAdded   []paperDomain.Paper
	CreatedPaperAndAdded []paperDomain.Paper
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

// CreateByName validates and stores a new stack for owner.
// It returns an error if the stack is invalid or could not be stored.
func (s *StackService) CreateByName(ctx context.Context, name string, owner userDomain.User) (domain.Stack, error) {
	stack := domain.NewStack(name, owner)
	if err := stack.Validate(); err != nil {
		return domain.Stack{}, err
	}

	return *stack, s.repo.Create(ctx, *stack)
}

// EnsureDefault creates the user's default stack when it does not already exist.
func (s *StackService) EnsureDefault(ctx context.Context, user userDomain.User) error {
	stacks, err := s.List(ctx, user.ExternalID)
	if err != nil {
		return err
	}

	for _, stack := range stacks {
		if strings.EqualFold(stack.Name, defaultStackName) {
			return nil
		}
	}

	stack := domain.NewStack(defaultStackName, user)
	err = s.Create(ctx, *stack)
	if errors.Is(err, domain.ErrStackAlreadyExists) {
		return nil
	}

	return err
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

// Import adds candidates to a stack, reusing papers with matching DOIs and creating missing papers.
func (s *StackService) Import(ctx context.Context, stackUUID string, candidates []bibliography.PaperEntry) (ImportResult, error) {
	stack, err := s.repo.GetByUUID(ctx, stackUUID)
	if err != nil {
		return ImportResult{}, err
	}

	result := ImportResult{}

	for _, candidate := range candidates {
		inStackPaper, foundInStack := stack.ContainsPaperWithDOI(candidate.Paper.DOI)
		if foundInStack {
			result.AlreadyInStack = append(result.AlreadyInStack, inStackPaper)
			continue
		}

		paper, err := s.paperGetter.GetByDOI(ctx, candidate.Paper.DOI)
		if err == nil {
			result.ExistingPaperAdded = append(result.ExistingPaperAdded, paper)
		}
		if errors.Is(err, paperDomain.ErrPaperNotFound) {
			paper, err = s.paperGetter.Create(ctx, candidate.Paper)
			result.CreatedPaperAndAdded = append(result.CreatedPaperAndAdded, paper)
		}
		if err != nil {
			return ImportResult{}, err
		}

		if err := s.repo.AddPaper(ctx, stackUUID, paper); err != nil {
			return ImportResult{}, err
		}
	}

	return result, nil
}

// RemovePaper removes a paper from the specified stack.
//
// If the paper is not assigned to the stack, no changes are made.
// It return an error if the stack does not exist
func (s *StackService) RemovePaper(ctx context.Context, stackUUID string, paperUUID string) error {
	return s.repo.RemovePaper(ctx, strings.TrimSpace(stackUUID), strings.TrimSpace(paperUUID))
}

// Search returns public stacks matching the provided search options.
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

	err := opts.Validate()
	if err != nil {
		return domain.SearchResult{}, err
	}

	return s.repo.SearchPublic(ctx, opts)
}

// SearchByOwner returns stacks of a user that match the provided search options.
// It normalizes the query and applies default pagination values before
//
// It returns an error if the stacks could not be searched.
func (s *StackService) SearchByOwner(ctx context.Context, userExternalID string, opts domain.SearchOptions) (domain.SearchResult, error) {
	opts.Query = strings.ToLower(strings.TrimSpace(opts.Query))
	opts.SortBy = strings.ToLower(strings.TrimSpace(opts.SortBy))

	opts.Page = max(defaultSearchPage, opts.Page)

	if opts.PageSize <= 1 {
		opts.PageSize = defaultSearchPageSize
	}
	opts.PageSize = min(maxSearchPageSize, opts.PageSize)

	err := opts.Validate()
	if err != nil {
		return domain.SearchResult{}, err
	}

	return s.repo.SearchByOwner(ctx, userExternalID, opts)
}

func (s *StackService) GetStatsByOwner(ctx context.Context, userExternalID string) (domain.Stats, error) {
	return s.repo.StatsByOwner(ctx, userExternalID)
}
