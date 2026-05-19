package memory

import (
	"context"
	"sync"

	"github.com/paperstacks.io/paperstacks/internal/user/domain"
)

var _ domain.Repository = (*Repository)(nil)

type Repository struct {
	mu   sync.RWMutex
	data []domain.User
}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) GetByExternalID(_ context.Context, externalID string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, item := range r.data {
		if item.ExternalID == externalID {
			return item, nil
		}
	}

	return domain.User{}, domain.ErrUserNotFound
}

func (r *Repository) GetByEmail(_ context.Context, email string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, item := range r.data {
		if item.Email == email {
			return item, nil
		}
	}

	return domain.User{}, domain.ErrUserNotFound
}

func (r *Repository) List(_ context.Context) ([]domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]domain.User, 0, len(r.data))
	out = append(out, r.data...)

	return out, nil
}

func (r *Repository) SaveIfNotExist(_ context.Context, user domain.User) (domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, item := range r.data {
		if item.ExternalID == user.ExternalID {
			return item, nil
		}
	}

	r.data = append(r.data, user)
	return user, nil
}

func (r *Repository) Update(_ context.Context, externalID string, user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, item := range r.data {
		if item.ExternalID == externalID {
			r.data[i] = user
			return nil
		}
	}

	return domain.ErrUserNotFound
}

func (r *Repository) Delete(_ context.Context, externalID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, item := range r.data {
		if item.ExternalID == externalID {
			r.data = append(r.data[:i], r.data[i+1:]...)
			return nil
		}
	}

	return domain.ErrUserNotFound
}
