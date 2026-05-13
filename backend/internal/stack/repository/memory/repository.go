package memory

import (
	"context"
	"sync"

	userDomain "github.com/paperstacks.io/paperstacks/internal/auth/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

type Repository struct {
	mu   sync.RWMutex
	data []domain.Stack
}

func NewRepository() *Repository {
	return &Repository{data: seedData()}
}

func (r *Repository) Create(ctx context.Context, stack domain.Stack) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, item := range r.data {
		if item.Name == stack.Name && item.Owner.ExternalID == stack.Owner.ExternalID {
			return domain.ErrStackAlreadyExists
		}
	}

	r.data = append(r.data, stack)
	return nil
}

func (r *Repository) Update(ctx context.Context, modified domain.Stack) (domain.Stack, error) {
	return domain.Stack{}, nil
}

func (r *Repository) Delete(ctx context.Context, uuid string) error {
	return nil
}

func (r *Repository) List(ctx context.Context, user userDomain.User) ([]domain.Stack, error) {
	return nil, nil
}
