package application

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

type PaginationItem struct {
	Page       int
	IsActive   bool
	IsEllipsis bool
	TargetPage int
}

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

func (s *PaperService) BuildPagination(
	total, pageSize, currentPage int,
) []PaginationItem {
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	if totalPages <= 0 {
		return nil
	}

	const visiblePagesCount = 7

	var items []PaginationItem

	firstPage := 1
	lastPage := totalPages
	windowSize := visiblePagesCount
	lastWindowStart := totalPages - windowSize + 1
	nextWindowTarget := windowSize + 1

	// Show all pages if they fit within the visible window
	if totalPages <= windowSize {
		for page := firstPage; page <= lastPage; page++ {
			items = appendPage(items, page, currentPage, totalPages)
		}
		return items
	}

	/*
		Beginning:
		Show the first window the last page
		Example: 1 2 3 4 5 6 7 ... last
	*/
	if currentPage < windowSize {
		for page := firstPage; page <= windowSize; page++ {
			items = appendPage(items, page, currentPage, totalPages)
		}
		items = appendEllipsis(items, nextWindowTarget, totalPages)
		items = appendPage(items, lastPage, currentPage, totalPages)
		return items
	}

	/*
		End:
		Show first page, ellipsis, and the last window
		Example: 1 ... last
	*/
	if currentPage > totalPages-(windowSize-1) {
		items = appendPage(items, firstPage, currentPage, totalPages)
		items = appendEllipsis(items, totalPages-windowSize, totalPages)

		for page := lastWindowStart; page <= lastPage; page++ {
			items = appendPage(items, page, currentPage, totalPages)
		}
		return items
	}

	/*
		Middle:
		Show first page, ellipsis, current page, ellipsis, and last page
		Example: 1 ... [window] ... last
	*/
	halfWindowSize := windowSize / 2
	windowStart := currentPage - halfWindowSize
	windowEnd := currentPage + halfWindowSize

	items = appendPage(items, firstPage, currentPage, totalPages)
	items = appendEllipsis(items, currentPage-windowSize, totalPages)

	for page := windowStart; page <= windowEnd; page++ {
		items = appendPage(items, page, currentPage, totalPages)
	}

	items = appendEllipsis(items, currentPage+windowSize, totalPages)
	items = appendPage(items, lastPage, currentPage, totalPages)

	return items
}

func appendPage(items []PaginationItem, page, currentPage, totalPages int) []PaginationItem {
	if page < 1 || page > totalPages {
		return items
	}

	return append(items, PaginationItem{
		Page:     page,
		IsActive: page == currentPage,
	})
}

func appendEllipsis(items []PaginationItem, targetPage, totalPages int) []PaginationItem {
	if targetPage < 1 {
		targetPage = 1
	}
	if targetPage > totalPages {
		targetPage = totalPages
	}

	return append(items, PaginationItem{
		IsEllipsis: true,
		TargetPage: targetPage,
	})
}
