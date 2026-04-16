package memory

import (
	"cmp"
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

type Repository struct {
	mu   sync.RWMutex
	data []domain.Paper
}

func NewRepository() *Repository {
	return &Repository{data: seedData()}
}

func (r *Repository) GetByUUID(_ context.Context, uuid string) (domain.Paper, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, item := range r.data {
		if item.UUID == uuid {
			return item, nil
		}
	}

	return domain.Paper{}, domain.ErrPaperNotFound
}

func (r *Repository) GetByDOI(_ context.Context, doi string) (domain.Paper, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, item := range r.data {
		if item.DOI == doi {
			return item, nil
		}
	}

	return domain.Paper{}, domain.ErrPaperNotFound
}

func (r *Repository) List(_ context.Context) ([]domain.Paper, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]domain.Paper, 0, len(r.data))
	for _, item := range r.data {
		out = append(out, item)
	}

	return out, nil
}

func (r *Repository) Save(_ context.Context, paper domain.Paper) (domain.Paper, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, item := range r.data {
		if item.UUID == paper.UUID {
			return domain.Paper{}, domain.ErrPaperAlreadyExists
		}
	}

	r.data = append(r.data, paper)
	return paper, nil
}

func (r *Repository) Update(_ context.Context, uuid string, paper domain.Paper) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, item := range r.data {
		if item.UUID == uuid {
			r.data[i] = paper
			return nil
		}
	}

	return domain.ErrPaperNotFound
}

func (r *Repository) Delete(_ context.Context, uuid string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, item := range r.data {
		if item.UUID == uuid {
			r.data = append(r.data[:i], r.data[i+1:]...)
			return nil
		}
	}

	return domain.ErrPaperNotFound
}

func (r *Repository) Search(_ context.Context, opts domain.SearchOptions) (domain.SearchResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Paper, 0, len(r.data))

	for _, paper := range r.data {
		if matchesQuery(paper, opts.Query) {
			result = append(result, paper)
		}
	}

	if opts.SortBy != "" {
		sortPapersByOrder(result, opts.SortBy, opts.Desc)
	}

	page := max(1, opts.Page)

	pageSize := len(result)
	if opts.PageSize > 0 {
		pageSize = opts.PageSize
	}
	pageSize = max(1, pageSize)

	total := len(result)
	start := min((page-1)*pageSize, total)
	end := min(start+pageSize, total)

	items := make([]domain.Paper, 0, end-start)
	items = append(items, result[start:end]...)

	println("total:", total)
	println("page:", page)

	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	pages := make([]int, 0, totalPages)
	for i := 1; i <= totalPages; i++ {
		pages = append(pages, i)
	}

	pagination := buildPagination(page, totalPages)

	return domain.SearchResult{
		Items:         items,
		Total:         total,
		Page:          page,
		NextPage:      page + 1,
		PrevPage:      page - 1,
		PageSize:      pageSize,
		SearchOptions: opts,
		HasNext:       end < total,
		Pages:         pages,
		Pagination:    pagination,
	}, nil
}

func sortPapersByOrder(papers []domain.Paper, sortBy string, desc bool) {
	sort.Slice(papers, func(i, j int) bool {
		if papers[i].DOI == papers[j].DOI {
			return compare(papers[i].UUID, papers[j].UUID, false)
		}

		switch sortBy {
		case "title":
			if papers[i].Title == papers[j].Title {
				return compare(papers[i].DOI, papers[j].DOI, false)
			}

			return compare(papers[i].Title, papers[j].Title, desc)
		case "year":
			if papers[i].PublicationYear == papers[j].PublicationYear {
				return compare(papers[i].DOI, papers[j].DOI, false)
			}

			return compare(papers[i].PublicationYear, papers[j].PublicationYear, desc)

		default:
			return compare(papers[i].DOI, papers[j].DOI, false)
		}
	})
}

func compare[T cmp.Ordered](a, b T, desc bool) bool {
	if desc {
		return a > b
	}

	return a < b
}

func containsText(slice []string, text string) bool {
	for _, v := range slice {
		if strings.Contains(strings.ToLower(v), text) {
			return true
		}
	}
	return false
}

func matchesQuery(paper domain.Paper, query string) bool {
	if query == "" {
		return true
	}

	if strings.Contains(strings.ToLower(paper.Title), query) {
		return true
	}

	return containsText(paper.Keywords, query)
}

func buildPagination(currentPage, totalPages int) []domain.PaginationItem {
	if totalPages <= 0 {
		return nil
	}

	const visiblePagesCount = 7

	var items []domain.PaginationItem

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

func appendPage(items []domain.PaginationItem, page, currentPage, totalPages int) []domain.PaginationItem {
	if page < 1 || page > totalPages {
		return items
	}

	return append(items, domain.PaginationItem{
		Page:     page,
		IsActive: page == currentPage,
	})
}

func appendEllipsis(items []domain.PaginationItem, targetPage, totalPages int) []domain.PaginationItem {
	if targetPage < 1 {
		targetPage = 1
	}
	if targetPage > totalPages {
		targetPage = totalPages
	}

	return append(items, domain.PaginationItem{
		IsEllipsis: true,
		TargetPage: targetPage,
	})
}
