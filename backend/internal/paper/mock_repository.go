package paper

import (
	"errors"
	"slices"
	"sync"

	"github.com/paperstacks.io/paperstacks/internal/domain"
)

var (
	ErrPaperNotFound      = errors.New("not found")
	ErrPaperAlreadyExists = errors.New("paper already exists")
)

type MemoryRepo struct {
	data []domain.Paper
	// mu protects concurrent access to data. RWMutex has a usable zero value.
	mu sync.RWMutex
}

func NewMemoryRepo() Repository {
	data := []domain.Paper{
		{
			DOI:   "1",
			Title: "Example Paper One",
		},
		{
			DOI:   "2",
			Title: "Example Paper Two",
		},
		{
			DOI:   "3",
			Title: "Example Paper Three",
		},
	}

	return &MemoryRepo{
		data: data,
	}
}

func (r *MemoryRepo) Create(paper domain.Paper) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if slices.ContainsFunc(r.data, func(p domain.Paper) bool {
		return p.DOI == paper.DOI
	}) {
		return ErrPaperAlreadyExists
	}

	r.data = append(r.data, paper)
	return nil
}

func (r *MemoryRepo) ReadAll() ([]domain.Paper, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.data) == 0 {
		return nil, ErrPaperNotFound
	}

	return r.data, nil
}

func (r *MemoryRepo) Read(id string) (domain.Paper, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	idx := slices.IndexFunc(r.data, func(p domain.Paper) bool {
		return p.DOI == id
	})

	if idx == -1 {
		return domain.Paper{}, ErrPaperNotFound
	}

	return r.data[idx], nil
}

func (r *MemoryRepo) Update(id string, paper domain.Paper) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	idx := slices.IndexFunc(r.data, func(p domain.Paper) bool {
		return p.DOI == id
	})

	if idx == -1 {
		return ErrPaperNotFound
	}

	paper.DOI = id
	r.data[idx] = paper

	return nil
}

func (r *MemoryRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	idx := slices.IndexFunc(r.data, func(p domain.Paper) bool {
		return p.DOI == id
	})

	if idx == -1 {
		return ErrPaperNotFound
	}

	r.data = append(r.data[:idx], r.data[idx+1:]...)

	return nil
}
