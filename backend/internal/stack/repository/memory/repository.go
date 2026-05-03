package memory

import (
	"sync"

	"github.com/paperstacks.io/paperstacks/internal/stacks/domain"
)

type Repository struct {
	mu   sync.RWMutex
	data []domain.Stack
}

func NewRepository() *Repository {
	return &Repository{data: seedData()}
}
