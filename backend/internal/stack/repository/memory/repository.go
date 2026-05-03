package memory

import (
	"sync"

	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

type Repository struct {
	mu   sync.RWMutex
	data []domain.Stack
}

func NewRepository() *Repository {
	return &Repository{data: seedData()}
}
