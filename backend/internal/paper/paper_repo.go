package paper

import (
	"errors"
	"sync"

	"github.com/paperstacks.io/paperstacks/internal/domain"
)

type Repository interface {
	Create(paper domain.Paper) error
	ReadAll() (map[string]domain.Paper, error)
	Read(id string) (domain.Paper, error)
	Update(id string, paper domain.Paper) error
	Delete(id string) error
}

var (
	ErrPaperNotFound      = errors.New("not found")
	ErrPaperAlreadyExists = errors.New("paper already exists")
)

type MemoryRepo struct {
	data map[string]domain.Paper
	// mu protects concurrent access to data. RWMutex has a usable zero value.
	mu sync.RWMutex
}

func NewMemoryRepo() Repository {
	data := map[string]domain.Paper{
		"paper-1": {
			ID:    "paper-1",
			Title: "Example Paper One",
		},
		"paper-2": {
			ID:    "paper-2",
			Title: "Example Paper Two",
		},
		"paper-3": {
			ID:    "paper-3",
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

	if _, exists := r.data[paper.ID]; exists {
		return ErrPaperAlreadyExists
	}

	r.data[paper.ID] = paper
	return nil
}

func (r *MemoryRepo) ReadAll() (map[string]domain.Paper, error) {
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

	paper, exists := r.data[id]
	if !exists {
		return domain.Paper{}, ErrPaperNotFound
	}

	return paper, nil
}

func (r *MemoryRepo) Update(id string, paper domain.Paper) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[id]; !exists {
		return ErrPaperNotFound
	}

	paper.ID = id
	r.data[id] = paper

	return nil
}

func (r *MemoryRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[id]; !exists {
		return ErrPaperNotFound
	}

	delete(r.data, id)
	return nil
}
