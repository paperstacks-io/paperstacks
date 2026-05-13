package memory

import (
	"context"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/user/domain"
)

func TestRepositorySaveAndGetByExternalID(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	user := domain.User{ExternalID: "auth0|123", Email: "user@example.com"}

	_, err := repo.Save(context.Background(), user)
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	got, err := repo.GetByExternalID(context.Background(), user.ExternalID)
	if err != nil {
		t.Fatalf("GetByExternalID() error = %v, want nil", err)
	}
	if got != user {
		t.Fatalf("GetByExternalID() = %v, want %v", got, user)
	}
}

func TestRepositoryGetByEmail(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	user := domain.User{ExternalID: "auth0|123", Email: "user@example.com"}

	_, err := repo.Save(context.Background(), user)
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	got, err := repo.GetByEmail(context.Background(), user.Email)
	if err != nil {
		t.Fatalf("GetByEmail() error = %v, want nil", err)
	}
	if got != user {
		t.Fatalf("GetByEmail() = %v, want %v", got, user)
	}
}

func TestRepositorySaveReturnsAlreadyExistsForDuplicateExternalID(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	_, err := repo.Save(context.Background(), domain.User{ExternalID: "auth0|123", Email: "one@example.com"})
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	_, err = repo.Save(context.Background(), domain.User{ExternalID: "auth0|123", Email: "two@example.com"})
	if err != domain.ErrUserAlreadyExists {
		t.Fatalf("Save() error = %v, want %v", err, domain.ErrUserAlreadyExists)
	}
}

func TestRepositorySaveReturnsAlreadyExistsForDuplicateEmail(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	_, err := repo.Save(context.Background(), domain.User{ExternalID: "auth0|123", Email: "user@example.com"})
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	_, err = repo.Save(context.Background(), domain.User{ExternalID: "auth0|456", Email: "user@example.com"})
	if err != domain.ErrUserAlreadyExists {
		t.Fatalf("Save() error = %v, want %v", err, domain.ErrUserAlreadyExists)
	}
}

func TestRepositoryUpdateReindexesEmail(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	ctx := context.Background()
	user := domain.User{ExternalID: "auth0|123", Email: "old@example.com"}
	updated := domain.User{ExternalID: "auth0|123", Email: "new@example.com"}

	_, err := repo.Save(ctx, user)
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	err = repo.Update(ctx, user.ExternalID, updated)
	if err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}

	_, err = repo.GetByEmail(ctx, user.Email)
	if err != domain.ErrUserNotFound {
		t.Fatalf("GetByEmail(old) error = %v, want %v", err, domain.ErrUserNotFound)
	}

	got, err := repo.GetByEmail(ctx, updated.Email)
	if err != nil {
		t.Fatalf("GetByEmail(new) error = %v, want nil", err)
	}
	if got != updated {
		t.Fatalf("GetByEmail(new) = %v, want %v", got, updated)
	}
}

func TestRepositoryUpdateRejectsDuplicateEmail(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	ctx := context.Background()

	_, err := repo.Save(ctx, domain.User{ExternalID: "auth0|123", Email: "one@example.com"})
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}
	_, err = repo.Save(ctx, domain.User{ExternalID: "auth0|456", Email: "two@example.com"})
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	err = repo.Update(ctx, "auth0|123", domain.User{ExternalID: "auth0|123", Email: "two@example.com"})
	if err != domain.ErrUserAlreadyExists {
		t.Fatalf("Update() error = %v, want %v", err, domain.ErrUserAlreadyExists)
	}
}

func TestRepositoryDeleteRemovesUserAndEmailIndex(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	ctx := context.Background()
	user := domain.User{ExternalID: "auth0|123", Email: "user@example.com"}

	_, err := repo.Save(ctx, user)
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	err = repo.Delete(ctx, user.ExternalID)
	if err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}

	_, err = repo.GetByExternalID(ctx, user.ExternalID)
	if err != domain.ErrUserNotFound {
		t.Fatalf("GetByExternalID() error = %v, want %v", err, domain.ErrUserNotFound)
	}

	_, err = repo.GetByEmail(ctx, user.Email)
	if err != domain.ErrUserNotFound {
		t.Fatalf("GetByEmail() error = %v, want %v", err, domain.ErrUserNotFound)
	}
}

func TestRepositoryListReturnsUsers(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	ctx := context.Background()

	_, err := repo.Save(ctx, domain.User{ExternalID: "auth0|123", Email: "one@example.com"})
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}
	_, err = repo.Save(ctx, domain.User{ExternalID: "auth0|456", Email: "two@example.com"})
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	got, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 2 {
		t.Fatalf("List() returned %d users, want 2", len(got))
	}
}
