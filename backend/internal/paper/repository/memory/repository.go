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

	if title == "" && keyword == "" {
		return []domain.Paper{}, nil
	}

	for _, paper := range r.data {

		matchesTitle := title != "" && strings.Contains(strings.ToLower(paper.Title), title)
		matchesKeyword := keyword != "" && containsText(paper.Keywords, keyword)

		if matchesTitle || matchesKeyword {
			result = append(result, paper)
		}
	}

	return result, nil
}

func containsText(slice []string, text string) bool {
	for _, v := range slice {
		if strings.Contains(strings.ToLower(v), text) {
			return true
		}
	}
	return false
}
