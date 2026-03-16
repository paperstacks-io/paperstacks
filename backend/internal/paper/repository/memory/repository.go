package memory

import (
	"context"
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

func (r *Repository) Save(_ context.Context, paper domain.Paper) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, item := range r.data {
		if item.DOI == paper.DOI {
			return domain.ErrPaperAlreadyExists
		}
	}

	r.data = append(r.data, paper)
	return nil
}

func (r *Repository) Update(_ context.Context, doi string, paper domain.Paper) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, item := range r.data {
		if item.DOI == doi {
			r.data[i] = paper
			return nil
		}
	}

	return domain.ErrPaperNotFound
}

func (r *Repository) Delete(_ context.Context, doi string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, item := range r.data {
		if item.DOI == doi {
			r.data = append(r.data[:i], r.data[i+1:]...)
			return nil
		}
	}

	return domain.ErrPaperNotFound
}

func (r *Repository) Search(_ context.Context, title, keyword string) ([]domain.Paper, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Paper
	set := make(map[string]struct{})

	for _, item := range r.data {
		titleMatch := false
		keywordMatch := false

		if title != "" && strings.Contains(strings.ToLower(item.Title), title) {
			titleMatch = true
		}

		if keyword != "" {
			for _, k := range item.Keywords {
				if strings.Contains(strings.ToLower(k), keyword) {
					keywordMatch = true
					break
				}
			}
		}

		if titleMatch || keywordMatch {
			key := item.DOI

			if _, exist := set[key]; !exist {
				set[key] = struct{}{}
				result = append(result, item)
			}
		}
	}

	if len(result) == 0 {
		return []domain.Paper{}, nil
	}

	return result, nil
}
