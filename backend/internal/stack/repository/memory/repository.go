package memory

import (
	"context"
	"sync"

	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
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
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, item := range r.data {
		if item.UUID == modified.UUID && item.Owner.ExternalID == modified.Owner.ExternalID {
			r.data[i] = modified
			return modified, nil
		}
	}

	return domain.Stack{}, domain.ErrStackNotFound
}

func (r *Repository) Delete(ctx context.Context, uuid string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, item := range r.data {
		if item.UUID == uuid {
			r.data = append(r.data[:i], r.data[i+1:]...)
			return nil
		}
	}

	return domain.ErrStackNotFound
}

func (r *Repository) GetByUUID(ctx context.Context, uuid string) (domain.Stack, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, item := range r.data {
		if item.UUID == uuid {
			return item, nil
		}
	}

	return domain.Stack{}, domain.ErrStackNotFound
}

func (r *Repository) List(ctx context.Context, userExternalID string) ([]domain.Stack, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stacks := make([]domain.Stack, 0)

	for _, item := range r.data {
		if item.Owner.ExternalID == userExternalID {
			stacks = append(stacks, item)
		}
	}

	return stacks, nil
}

func (r *Repository) ListPublic(ctx context.Context, userExternalID string) ([]domain.Stack, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stacks := make([]domain.Stack, 0)

	for _, item := range r.data {
		if item.Owner.ExternalID == userExternalID && item.IsPublic {
			stacks = append(stacks, item)
		}
	}

	return stacks, nil
}

func (r *Repository) AddPaper(ctx context.Context, stackUUID string, paper paperDomain.Paper) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, s := range r.data {
		if s.UUID == stackUUID {
			for _, p := range s.Papers {
				if p.UUID == paper.UUID {
					return nil
				}
			}
			r.data[i].Papers = append(r.data[i].Papers, paper)
			return nil
		}
	}
	return domain.ErrStackNotFound
}

func (r *Repository) RemovePaper(ctx context.Context, stackUUID string, paperUUID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, item := range r.data {
		if item.UUID == stackUUID {
			for j, p := range item.Papers {
				if p.UUID == paperUUID {
					r.data[i].Papers = append(r.data[i].Papers[:j], r.data[i].Papers[j+1:]...)
					return nil
				}
			}
			return nil
		}
	}
	return domain.ErrStackNotFound
}

func (r *Repository) ListAllPublic(ctx context.Context) ([]domain.Stack, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stacks := make([]domain.Stack, 0)

	for _, item := range r.data {
		if item.IsPublic {
			stacks = append(stacks, item)
		}
	}

	return stacks, nil
}
