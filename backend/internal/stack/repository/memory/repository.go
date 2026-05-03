package memory

import (
	"context"
	"sync"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

type Repository struct {
	mu   sync.RWMutex
	data []domain.Stack
}

func NewRepository() *Repository {
	return &Repository{data: seedData()}
}

func (r *Repository) List(ctx context.Context) ([]domain.Stack, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := []domain.Stack{}

	for _, item := range r.data {
		if item.Visibility == domain.VisibilityPublic {
			out = append(out, item)
		}
	}

	return out, nil
}

func (r *Repository) ListByOwner(ctx context.Context, owner string) ([]domain.Stack, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := []domain.Stack{}

	for _, item := range r.data {
		if item.Owner == owner {
			out = append(out, item)
		}
	}

	return out, nil
}

func (r *Repository) ListPrivateByOwner(_ context.Context, owner string) ([]domain.Stack, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := []domain.Stack{}

	for _, item := range r.data {
		if item.Owner == owner && item.Visibility == domain.VisibilityPrivate {
			out = append(out, item)
		}
	}

	return out, nil
}

func (r *Repository) ListPublicByOwner(ctx context.Context, owner string) ([]domain.Stack, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := []domain.Stack{}

	for _, item := range r.data {
		if item.Owner == owner && item.Visibility == domain.VisibilityPublic {
			out = append(out, item)
		}
	}

	return out, nil
}

func (r *Repository) Save(ctx context.Context, stack domain.Stack) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, item := range r.data {
		if item.Name == stack.Name && item.Owner == stack.Owner {
			return domain.ErrStackAlreadyExists
		}
	}

	r.data = append(r.data, stack)
	return nil
}

func (r *Repository) Delete(ctx context.Context, stack domain.Stack) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, item := range r.data {
		if item.UUID == stack.UUID {
			r.data = append(r.data[:i], r.data[i+1:]...)
			return nil
		}
	}

	return domain.ErrStackNotFound
}

func (r *Repository) SetVisibility(ctx context.Context, stack domain.Stack, visibility domain.Visibility) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, item := range r.data {
		if item.UUID == stack.UUID {
			r.data[i].Visibility = visibility
			r.data[i].UpdatedAt = time.Now()
			return nil
		}
	}

	return domain.ErrStackNotFound
}
