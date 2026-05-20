// Package application provides paper application services.
package application

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

const (
	defaultSearchPage     = 1
	defaultSearchPageSize = 10
	maxSearchPageSize     = 100
)

type PaperService struct {
	repo domain.Repository
}

func NewPaperService(repo domain.Repository) *PaperService {
	return &PaperService{
		repo: repo,
	}
}

func (s *PaperService) List(ctx context.Context) ([]domain.Paper, error) {
	return s.repo.List(ctx)
}

func (s *PaperService) GetByUUID(ctx context.Context, uuid string) (domain.Paper, error) {
	return s.repo.GetByUUID(ctx, strings.TrimSpace(uuid))
}

func (s *PaperService) GetByDOI(ctx context.Context, doi string) (domain.Paper, error) {
	return s.repo.GetByDOI(ctx, strings.TrimSpace(doi))
}

func (s *PaperService) Create(ctx context.Context, paper domain.Paper) (domain.Paper, error) {
	paper.UUID = uuid.NewString()
	paper = paper.Normalize()
	if err := paper.Validate(); err != nil {
		return domain.Paper{}, err
	}

	return s.repo.Save(ctx, paper)
}

func (s *PaperService) Update(ctx context.Context, uuid string, paper domain.Paper) error {
	uuid = strings.TrimSpace(uuid)
	paper = paper.Normalize()

	if paper.UUID != uuid {
		return domain.ErrUUIDMismatch
	}

	if err := paper.Validate(); err != nil {
		return err
	}

	return s.repo.Update(ctx, uuid, paper)
}

func (s *PaperService) Delete(ctx context.Context, uuid string) error {
	return s.repo.Delete(ctx, strings.TrimSpace(uuid))
}

func (s *PaperService) Search(ctx context.Context, opts domain.SearchOptions) (domain.SearchResult, error) {
	opts.Query = strings.ToLower(strings.TrimSpace(opts.Query))
	opts.SortBy = strings.ToLower(strings.TrimSpace(opts.SortBy))

	opts.Page = max(defaultSearchPage, opts.Page)

	if opts.PageSize < 1 {
		opts.PageSize = defaultSearchPageSize
	}
	opts.PageSize = min(maxSearchPageSize, opts.PageSize)

	if opts.SortBy != "" && opts.SortBy != "title" && opts.SortBy != "year" {
		return domain.SearchResult{}, domain.ErrInvalidSearch
	}

	return s.repo.Search(ctx, opts)
}
