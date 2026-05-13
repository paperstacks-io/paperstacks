package memory

import (
	"context"
	"strings"
	"sync"

	"github.com/paperstacks.io/paperstacks/internal/user/domain"
)

type Repository struct {
	mu                sync.RWMutex
	byExternalID      map[string]domain.User
	emailToExternalID map[string]string
}

func NewRepository() *Repository {
	return &Repository{
		byExternalID:      make(map[string]domain.User),
		emailToExternalID: make(map[string]string),
	}
}

func (r *Repository) GetByExternalID(_ context.Context, externalID string) (domain.User, error) {
	externalID = strings.TrimSpace(externalID)

	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.byExternalID[externalID]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}

	return user, nil
}

func (r *Repository) GetByEmail(_ context.Context, email string) (domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	r.mu.RLock()
	defer r.mu.RUnlock()

	externalID, ok := r.emailToExternalID[email]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}

	return r.byExternalID[externalID], nil
}

func (r *Repository) List(_ context.Context) ([]domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]domain.User, 0, len(r.byExternalID))
	for _, user := range r.byExternalID {
		out = append(out, user)
	}

	return out, nil
}

func (r *Repository) Save(_ context.Context, user domain.User) (domain.User, error) {
	user = user.Normalize()

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byExternalID[user.ExternalID]; ok {
		return domain.User{}, domain.ErrUserAlreadyExists
	}

	if _, ok := r.emailToExternalID[user.Email]; ok {
		return domain.User{}, domain.ErrUserAlreadyExists
	}

	r.byExternalID[user.ExternalID] = user
	r.emailToExternalID[user.Email] = user.ExternalID

	return user, nil
}

func (r *Repository) Update(_ context.Context, externalID string, user domain.User) error {
	externalID = strings.TrimSpace(externalID)
	user = user.Normalize()
	if user.ExternalID != externalID {
		return domain.ErrExternalIDMismatch
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.byExternalID[externalID]
	if !ok {
		return domain.ErrUserNotFound
	}

	if ownerExternalID, ok := r.emailToExternalID[user.Email]; ok && ownerExternalID != externalID {
		return domain.ErrUserAlreadyExists
	}

	delete(r.emailToExternalID, existing.Email)
	r.byExternalID[externalID] = user
	r.emailToExternalID[user.Email] = externalID

	return nil
}

func (r *Repository) Delete(_ context.Context, externalID string) error {
	externalID = strings.TrimSpace(externalID)

	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.byExternalID[externalID]
	if !ok {
		return domain.ErrUserNotFound
	}

	delete(r.byExternalID, externalID)
	delete(r.emailToExternalID, user.Email)

	return nil
}
