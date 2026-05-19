package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/user/application"
	"github.com/paperstacks.io/paperstacks/internal/user/domain"
	"github.com/paperstacks.io/paperstacks/internal/user/repository/memory"
)

func TestUserServiceCreateIfNotExistNormalizesUser(t *testing.T) {
	t.Parallel()

	service := application.NewUserService(memory.NewRepository())

	user, err := service.CreateIfNotExist(context.Background(), " external-1 ", " ONE@EXAMPLE.COM ")
	if err != nil {
		t.Fatalf("CreateIfNotExist() error = %v, want nil", err)
	}
	if user.ExternalID != "external-1" {
		t.Fatalf("CreateIfNotExist() externalID = %q, want %q", user.ExternalID, "external-1")
	}
	if user.Email != "one@example.com" {
		t.Fatalf("CreateIfNotExist() email = %q, want %q", user.Email, "one@example.com")
	}
	if user.CreatedAt.IsZero() {
		t.Fatalf("CreateIfNotExist() CreatedAt is zero, want set")
	}
	if user.UpdatedAt.IsZero() {
		t.Fatalf("CreateIfNotExist() UpdatedAt is zero, want set")
	}
}

func TestUserServiceCreateIfNotExistReturnsInvalidUser(t *testing.T) {
	t.Parallel()

	service := application.NewUserService(memory.NewRepository())

	_, err := service.CreateIfNotExist(context.Background(), "external-1", "invalid-email")
	if err != domain.ErrInvalidUser {
		t.Fatalf("CreateIfNotExist() error = %v, want %v", err, domain.ErrInvalidUser)
	}
}

func TestUserServiceCreateIfNotExistReturnsExistingUser(t *testing.T) {
	t.Parallel()

	service := application.NewUserService(memory.NewRepository())

	existing, err := service.CreateIfNotExist(context.Background(), "external-1", "one@example.com")
	if err != nil {
		t.Fatalf("CreateIfNotExist() error = %v, want nil", err)
	}

	res, err := service.CreateIfNotExist(context.Background(), "external-1", "changed@example.com")
	if err != nil {
		t.Fatalf("CreateIfNotExist() error = %v, want nil", err)
	}
	if res != existing {
		t.Fatalf("CreateIfNotExist() = %#v, want existing %#v", res, existing)
	}
}

func TestUserServiceGetByExternalIDTrimsInput(t *testing.T) {
	t.Parallel()

	service := application.NewUserService(memory.NewRepository())
	created, err := service.CreateIfNotExist(context.Background(), "external-1", "one@example.com")
	if err != nil {
		t.Fatalf("CreateIfNotExist() error = %v, want nil", err)
	}

	user, err := service.GetByExternalID(context.Background(), " external-1 ")
	if err != nil {
		t.Fatalf("GetByExternalID() error = %v, want nil", err)
	}
	if user != created {
		t.Fatalf("GetByExternalID() = %#v, want %#v", user, created)
	}
}

func TestUserServiceGetByEmailNormalizesInput(t *testing.T) {
	t.Parallel()

	service := application.NewUserService(memory.NewRepository())
	created, err := service.CreateIfNotExist(context.Background(), "external-1", "one@example.com")
	if err != nil {
		t.Fatalf("CreateIfNotExist() error = %v, want nil", err)
	}

	user, err := service.GetByEmail(context.Background(), " ONE@EXAMPLE.COM ")
	if err != nil {
		t.Fatalf("GetByEmail() error = %v, want nil", err)
	}
	if user != created {
		t.Fatalf("GetByEmail() = %#v, want %#v", user, created)
	}
}

func TestUserServiceListReturnsUsers(t *testing.T) {
	t.Parallel()

	service := application.NewUserService(memory.NewRepository())
	if _, err := service.CreateIfNotExist(context.Background(), "external-1", "one@example.com"); err != nil {
		t.Fatalf("CreateIfNotExist() error = %v, want nil", err)
	}
	if _, err := service.CreateIfNotExist(context.Background(), "external-2", "two@example.com"); err != nil {
		t.Fatalf("CreateIfNotExist() error = %v, want nil", err)
	}

	users, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(users) != 2 {
		t.Fatalf("List() returned %d users, want 2", len(users))
	}
}

func TestUserServiceUpdate(t *testing.T) {
	t.Parallel()

	service := application.NewUserService(memory.NewRepository())
	created, err := service.CreateIfNotExist(context.Background(), "external-1", "one@example.com")
	if err != nil {
		t.Fatalf("CreateIfNotExist() error = %v, want nil", err)
	}

	updated := created
	updated.Email = "updated@example.com"
	updated.UpdatedAt = time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)

	err = service.Update(context.Background(), " external-1 ", updated)
	if err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}

	stored, err := service.GetByExternalID(context.Background(), "external-1")
	if err != nil {
		t.Fatalf("GetByExternalID() error = %v, want nil", err)
	}
	if stored.Email != "updated@example.com" {
		t.Fatalf("stored email = %q, want %q", stored.Email, "updated@example.com")
	}
	if !stored.UpdatedAt.After(updated.UpdatedAt) {
		t.Fatalf("stored UpdatedAt = %v, want after %v", stored.UpdatedAt, updated.UpdatedAt)
	}
}

func TestUserServiceUpdateReturnsExternalIDMismatch(t *testing.T) {
	t.Parallel()

	service := application.NewUserService(memory.NewRepository())
	user := domain.NewUser("external-2", "two@example.com")

	err := service.Update(context.Background(), "external-1", user)
	if err != domain.ErrExternalIDMismatch {
		t.Fatalf("Update() error = %v, want %v", err, domain.ErrExternalIDMismatch)
	}
}

func TestUserServiceUpdateReturnsInvalidUser(t *testing.T) {
	t.Parallel()

	service := application.NewUserService(memory.NewRepository())
	user := domain.User{ExternalID: "external-1", Email: "invalid-email"}

	err := service.Update(context.Background(), "external-1", user)
	if err != domain.ErrInvalidUser {
		t.Fatalf("Update() error = %v, want %v", err, domain.ErrInvalidUser)
	}
}

func TestUserServiceDeleteTrimsInput(t *testing.T) {
	t.Parallel()

	service := application.NewUserService(memory.NewRepository())
	if _, err := service.CreateIfNotExist(context.Background(), "external-1", "one@example.com"); err != nil {
		t.Fatalf("CreateIfNotExist() error = %v, want nil", err)
	}

	err := service.Delete(context.Background(), " external-1 ")
	if err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}

	_, err = service.GetByExternalID(context.Background(), "external-1")
	if err != domain.ErrUserNotFound {
		t.Fatalf("GetByExternalID() error = %v, want %v", err, domain.ErrUserNotFound)
	}
}
