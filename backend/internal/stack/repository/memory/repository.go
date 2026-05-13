package memory

import "sync"

type Repository struct {
	mu sync.RWMutex
}

func NewRepository() *Repository {
	return &Repository{}
}
