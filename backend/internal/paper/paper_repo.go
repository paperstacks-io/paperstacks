package paper

import (
	"errors"
	"sync"

	"github.com/paperstacks.io/paperstacks/internal/domain"
)

type Repository interface {
	Create(paper domain.Paper) (domain.Paper, error)
	Read(id string) (domain.Paper, error)
	Update(id string, paper domain.Paper) (domain.Paper, error)
	Delete(id string) error
}

var (
	errPaperExists   = errors.New("paper already exists")
	errPaperNotFound = errors.New("paper not found")
)

type MemoryRepo struct {
	data map[string]domain.Paper
	// mu protects concurrent access to data. RWMutex has a usable zero value.
	mu sync.RWMutex
}

func NewMemoryRepo() Repository {
	return &MemoryRepo{
		data: make(map[string]domain.Paper),
	}
}

func (r *MemoryRepo) Create(paper domain.Paper) (domain.Paper, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[paper.ID]; exists {
		return domain.Paper{}, errPaperExists
	}

	r.data[paper.ID] = paper
	return paper, nil
}

func (r *MemoryRepo) Read(id string) (domain.Paper, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	paper, exists := r.data[id]
	if !exists {
		return domain.Paper{}, errPaperNotFound
	}

	return paper, nil
}

func (r *MemoryRepo) Update(id string, paper domain.Paper) (domain.Paper, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[id]; !exists {
		return domain.Paper{}, errPaperNotFound
	}

	paper.ID = id
	r.data[id] = paper

	return paper, nil
}

func (r *MemoryRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[id]; !exists {
		return errPaperNotFound
	}

	delete(r.data, id)
	return nil
}
