package memory

import (
	"context"
	"sync"

	"github.com/paperstacks.io/paperstacks/internal/document/domain"
)

var _ domain.Repository = (*Repository)(nil)

type Repository struct {
	mu   sync.RWMutex
	data map[string]domain.Document
}

func NewRepository() *Repository {
	repo := &Repository{
		data: make(map[string]domain.Document),
	}
	return repo
}

func (r *Repository) Save(ctx context.Context, doc domain.Document) (domain.Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[doc.UUID] = doc
	return doc, nil
}
