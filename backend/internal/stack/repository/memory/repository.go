package memory

import (
	"cmp"
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

type Repository struct {
	mu               sync.RWMutex
	data             []domain.Stack
	countPublicCache int
	countPublicDirty bool
}

func NewRepository() *Repository {
	return &Repository{data: seedData(), countPublicDirty: true}
}

func (r *Repository) Create(ctx context.Context, stack domain.Stack) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, item := range r.data {
		if strings.ToLower(item.Name) == strings.ToLower(stack.Name) && item.Owner.ExternalID == stack.Owner.ExternalID {
			return domain.ErrStackAlreadyExists
		}
	}

	r.data = append(r.data, stack)
	r.countPublicDirty = true
	return nil
}

func (r *Repository) Update(ctx context.Context, modified domain.Stack) (domain.Stack, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, item := range r.data {
		if item.UUID == modified.UUID && item.Owner.ExternalID == modified.Owner.ExternalID {
			r.data[i] = modified
			return modified, nil
		}
	}

	return domain.Stack{}, domain.ErrStackNotFound
}

func (r *Repository) Delete(ctx context.Context, uuid string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.countPublicDirty = true

	for i, item := range r.data {
		if item.UUID == uuid {
			r.data = append(r.data[:i], r.data[i+1:]...)
			return nil
		}
	}

	return domain.ErrStackNotFound
}

func (r *Repository) GetByUUID(ctx context.Context, uuid string) (domain.Stack, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, item := range r.data {
		if item.UUID == uuid {
			return item, nil
		}
	}

	return domain.Stack{}, domain.ErrStackNotFound
}

func (r *Repository) List(ctx context.Context, userExternalID string) ([]domain.Stack, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stacks := make([]domain.Stack, 0)

	for _, item := range r.data {
		if item.Owner.ExternalID == userExternalID {
			stacks = append(stacks, item)
		}
	}

	return stacks, nil
}

func (r *Repository) ListPublic(ctx context.Context, userExternalID string) ([]domain.Stack, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stacks := make([]domain.Stack, 0)

	for _, item := range r.data {
		if item.Owner.ExternalID == userExternalID && item.IsPublic {
			stacks = append(stacks, item)
		}
	}

	return stacks, nil
}

func (r *Repository) CountPublic(ctx context.Context) (int, error) {
	r.mu.RLock()
	if !r.countPublicDirty {
		count := r.countPublicCache
		r.mu.RUnlock()
		return count, nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.countPublicDirty {
		return r.countPublicCache, nil
	}

	counter := 0
	for _, item := range r.data {
		if item.IsPublic {
			counter++
		}
	}

	r.countPublicCache = counter
	r.countPublicDirty = false
	return counter, nil
}

func (r *Repository) AddPaper(ctx context.Context, stackUUID string, paper paperDomain.Paper) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, s := range r.data {
		if s.UUID == stackUUID {
			for _, p := range s.Papers {
				if p.UUID == paper.UUID {
					return nil
				}
			}
			r.data[i].Papers = append(r.data[i].Papers, paper)
			return nil
		}
	}
	return domain.ErrStackNotFound
}

func (r *Repository) RemovePaper(ctx context.Context, stackUUID string, paperUUID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, item := range r.data {
		if item.UUID == stackUUID {
			for j, p := range item.Papers {
				if p.UUID == paperUUID {
					r.data[i].Papers = append(r.data[i].Papers[:j], r.data[i].Papers[j+1:]...)
					return nil
				}
			}
			return nil
		}
	}
	return domain.ErrStackNotFound
}

func (r *Repository) SearchPublic(_ context.Context, opts domain.SearchOptions) (domain.SearchResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Stack, 0, len(r.data))

	for _, stack := range r.data {
		if stack.IsPublic && matchesQuery(stack, opts.Query) {
			result = append(result, stack)
		}
	}

	if opts.SortBy != "" {
		sortStacksByOrder(result, opts.SortBy, opts.Desc)
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

	items := make([]domain.Stack, 0, end-start)
	items = append(items, result[start:end]...)

	return domain.SearchResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasNext:  end < total,
	}, nil
}

func (r *Repository) SearchByOwner(ctx context.Context, userExternalID string, opts domain.SearchOptions) (domain.SearchResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Stack, 0, len(r.data))

	for _, stack := range r.data {
		if stack.Owner.ExternalID == userExternalID && matchesQuery(stack, opts.Query) {
			result = append(result, stack)
		}
	}

	if opts.SortBy != "" {
		sortStacksByOrder(result, opts.SortBy, opts.Desc)
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

	items := make([]domain.Stack, 0, end-start)
	items = append(items, result[start:end]...)
	return domain.SearchResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasNext:  end < total,
	}, nil
}

func (r *Repository) StatsByOwner(ctx context.Context, userExternalID string) (domain.Stats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := domain.Stats{}
	for _, stack := range r.data {
		if stack.Owner.ExternalID != userExternalID {
			continue
		}

		stats.TotalStacks++
		stats.TotalPapers += len(stack.Papers)
		if stack.IsPublic {
			stats.PublicStacks++
		}
	}

	return stats, nil
}

func sortStacksByOrder(stacks []domain.Stack, sortBy string, desc bool) {
	sort.Slice(stacks, func(i, j int) bool {
		switch sortBy {
		case "name":
			if strings.EqualFold(strings.ToLower(stacks[i].Name), strings.ToLower(stacks[j].Name)) {
				return compare(stacks[i].UUID, stacks[j].UUID, false)
			}

			return compare(strings.ToLower(stacks[i].Name), strings.ToLower(stacks[j].Name), desc)
		case "updated_at":
			if stacks[i].UpdatedAt.Equal(stacks[j].UpdatedAt) {
				return compare(stacks[i].UUID, stacks[j].UUID, false)
			}

			return compareTime(stacks[i].UpdatedAt, stacks[j].UpdatedAt, desc)

		default:
			return compare(stacks[i].UUID, stacks[j].UUID, false)
		}
	})
}

func compare[T cmp.Ordered](a, b T, desc bool) bool {
	if desc {
		return a > b
	}
	return a < b
}

func compareTime(a, b time.Time, desc bool) bool {
	if desc {
		return a.After(b)
	}

	return a.Before(b)
}

func matchesQuery(stack domain.Stack, query string) bool {
	if query == "" {
		return true
	}

	query = strings.TrimSpace(strings.ToLower(query))

	return strings.Contains(strings.ToLower(stack.Name), query)
}
