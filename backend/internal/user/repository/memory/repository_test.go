package memory

import (
	"context"
	"testing"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/user/domain"
)

func TestRepositorySaveIfNotExistCreatesUser(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	user := newUser("external-1", "one@example.com")

	res, err := repo.SaveIfNotExist(context.Background(), user)
	if err != nil {
		t.Fatalf("SaveIfNotExist() error = %v, want nil", err)
	}
	if res != user {
		t.Fatalf("SaveIfNotExist() = %#v, want %#v", res, user)
	}
}

func TestRepositorySaveIfNotExistReturnsExistingUser(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	existing := newUser("external-1", "one@example.com")
	duplicate := newUser("external-1", "changed@example.com")

	if _, err := repo.SaveIfNotExist(context.Background(), existing); err != nil {
		t.Fatalf("SaveIfNotExist() error = %v, want nil", err)
	}

	res, err := repo.SaveIfNotExist(context.Background(), duplicate)
	if err != nil {
		t.Fatalf("SaveIfNotExist() error = %v, want nil", err)
	}
	if res != existing {
		t.Fatalf("SaveIfNotExist() = %#v, want existing %#v", res, existing)
	}
}

func TestRepositoryGetByExternalID(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	user := newUser("external-1", "one@example.com")
	if _, err := repo.SaveIfNotExist(context.Background(), user); err != nil {
		t.Fatalf("SaveIfNotExist() error = %v, want nil", err)
	}

	res, err := repo.GetByExternalID(context.Background(), "external-1")
	if err != nil {
		t.Fatalf("GetByExternalID() error = %v, want nil", err)
	}
	if res != user {
		t.Fatalf("GetByExternalID() = %#v, want %#v", res, user)
	}
}

func TestRepositoryGetByExternalIDReturnsNotFound(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	_, err := repo.GetByExternalID(context.Background(), "missing")
	if err != domain.ErrUserNotFound {
		t.Fatalf("GetByExternalID() error = %v, want %v", err, domain.ErrUserNotFound)
	}
}

func TestRepositoryGetByEmail(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	user := newUser("external-1", "one@example.com")
	if _, err := repo.SaveIfNotExist(context.Background(), user); err != nil {
		t.Fatalf("SaveIfNotExist() error = %v, want nil", err)
	}

	res, err := repo.GetByEmail(context.Background(), "one@example.com")
	if err != nil {
		t.Fatalf("GetByEmail() error = %v, want nil", err)
	}
	if res != user {
		t.Fatalf("GetByEmail() = %#v, want %#v", res, user)
	}
}

func TestRepositoryGetByEmailReturnsNotFound(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	_, err := repo.GetByEmail(context.Background(), "missing@example.com")
	if err != domain.ErrUserNotFound {
		t.Fatalf("GetByEmail() error = %v, want %v", err, domain.ErrUserNotFound)
	}
}

func TestRepositoryListReturnsAllUsers(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	users := []domain.User{
		newUser("external-1", "one@example.com"),
		newUser("external-2", "two@example.com"),
	}
	for _, user := range users {
		if _, err := repo.SaveIfNotExist(context.Background(), user); err != nil {
			t.Fatalf("SaveIfNotExist() error = %v, want nil", err)
		}
	}

	res, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(res) != len(users) {
		t.Fatalf("List() returned %d users, want %d", len(res), len(users))
	}
	for i, user := range users {
		if res[i] != user {
			t.Fatalf("List()[%d] = %#v, want %#v", i, res[i], user)
		}
	}
}

func TestRepositoryListReturnsCopy(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	user := newUser("external-1", "one@example.com")
	if _, err := repo.SaveIfNotExist(context.Background(), user); err != nil {
		t.Fatalf("SaveIfNotExist() error = %v, want nil", err)
	}

	res, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	res[0].Email = "modified@example.com"

	stored, err := repo.GetByExternalID(context.Background(), "external-1")
	if err != nil {
		t.Fatalf("GetByExternalID() error = %v, want nil", err)
	}
	if stored.Email != user.Email {
		t.Fatalf("stored email = %q, want %q", stored.Email, user.Email)
	}
}

func TestRepositoryUpdate(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	user := newUser("external-1", "one@example.com")
	updated := newUser("external-1", "updated@example.com")
	if _, err := repo.SaveIfNotExist(context.Background(), user); err != nil {
		t.Fatalf("SaveIfNotExist() error = %v, want nil", err)
	}

	err := repo.Update(context.Background(), "external-1", updated)
	if err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}

	stored, err := repo.GetByExternalID(context.Background(), "external-1")
	if err != nil {
		t.Fatalf("GetByExternalID() error = %v, want nil", err)
	}
	if stored != updated {
		t.Fatalf("stored user = %#v, want %#v", stored, updated)
	}
}

func TestRepositoryUpdateReturnsNotFound(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	user := newUser("external-1", "one@example.com")

	err := repo.Update(context.Background(), "missing", user)
	if err != domain.ErrUserNotFound {
		t.Fatalf("Update() error = %v, want %v", err, domain.ErrUserNotFound)
	}
}

func TestRepositoryDelete(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	user := newUser("external-1", "one@example.com")
	if _, err := repo.SaveIfNotExist(context.Background(), user); err != nil {
		t.Fatalf("SaveIfNotExist() error = %v, want nil", err)
	}

	err := repo.Delete(context.Background(), "external-1")
	if err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}

	_, err = repo.GetByExternalID(context.Background(), "external-1")
	if err != domain.ErrUserNotFound {
		t.Fatalf("GetByExternalID() error = %v, want %v", err, domain.ErrUserNotFound)
	}
}

func TestRepositoryDeleteReturnsNotFound(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	err := repo.Delete(context.Background(), "missing")
	if err != domain.ErrUserNotFound {
		t.Fatalf("Delete() error = %v, want %v", err, domain.ErrUserNotFound)
	}
}

func newUser(externalID, email string) domain.User {
	now := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	return domain.User{
		ExternalID: externalID,
		Email:      email,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
