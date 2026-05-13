package domain_test

import (
	"context"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/user/domain"
	"github.com/paperstacks.io/paperstacks/internal/user/repository/memory"
)

func TestServiceCreateNormalizesAndValidatesUser(t *testing.T) {
	t.Parallel()

	service := domain.NewUserService(memory.NewRepository())

	_, err := service.Create(context.Background(), domain.User{
		ExternalID: " auth0|123 ",
		Email:      " User@Example.COM ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	got, err := service.GetByEmail(context.Background(), " USER@example.com ")
	if err != nil {
		t.Fatalf("GetByEmail() error = %v, want nil", err)
	}

	if got.ExternalID != "auth0|123" {
		t.Fatalf("stored externalID = %q, want %q", got.ExternalID, "auth0|123")
	}
	if got.Email != "user@example.com" {
		t.Fatalf("stored email = %q, want %q", got.Email, "user@example.com")
	}
}

func TestServiceCreateRejectsInvalidUser(t *testing.T) {
	t.Parallel()

	service := domain.NewUserService(memory.NewRepository())

	_, err := service.Create(context.Background(), domain.User{
		ExternalID: "auth0|123",
		Email:      "not an email",
	})
	if err != domain.ErrInvalidUser {
		t.Fatalf("Create() error = %v, want %v", err, domain.ErrInvalidUser)
	}
}

func TestServiceUpdateRejectsMismatchedExternalID(t *testing.T) {
	t.Parallel()

	service := domain.NewUserService(memory.NewRepository())

	err := service.Update(context.Background(), "auth0|123", domain.User{
		ExternalID: "auth0|456",
		Email:      "user@example.com",
	})
	if err != domain.ErrExternalIDMismatch {
		t.Fatalf("Update() error = %v, want %v", err, domain.ErrExternalIDMismatch)
	}
}

func TestServiceGetByExternalIDTrimsInput(t *testing.T) {
	t.Parallel()

	service := domain.NewUserService(memory.NewRepository())

	_, err := service.Create(context.Background(), domain.User{
		ExternalID: "auth0|123",
		Email:      "user@example.com",
	})
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}

	got, err := service.GetByExternalID(context.Background(), " auth0|123 ")
	if err != nil {
		t.Fatalf("GetByExternalID() error = %v, want nil", err)
	}
	if got.ExternalID != "auth0|123" {
		t.Fatalf("GetByExternalID() externalID = %q, want %q", got.ExternalID, "auth0|123")
	}
}
