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

	return domain.SearchResult{
		Items:         items,
		Total:         total,
		Page:          page,
		NextPage:      page + 1,
		PageSize:      pageSize,
		SearchOptions: opts,
		HasNext:       end < total,
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
