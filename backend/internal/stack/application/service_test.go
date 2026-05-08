package application

import (
	"testing"

	userDomain "github.com/paperstacks.io/paperstacks/internal/auth/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

var testUser = userDomain.User{
	Email: "test@example.com",
}

// Create tests
func TestServiceCreateNormalizesAndValidatesStack(t *testing.T) {
	t.Parallel()
	stack := domain.NewStack(" Test Stack ", testUser)

	if stack.Name != "Test Stack" {
		t.Fatalf("expected name to be trimmed, got %q", stack.Name)
	}

	if stack.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}

	if stack.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be set")
	}

	if stack.UpdatedAt != stack.CreatedAt {
		t.Fatal("exptcted UpdatedAt to equal CreatedAt on creation")
	}

	if stack.UUID == "" {
		t.Fatal("expected UUID to be set")
	}

	if stack.Owner.Email != testUser.Email {
		t.Fatalf("expected owner email to be %q, got %q", testUser.Email, stack.Owner.Email)
	}

	if len(stack.Papers) != 0 {
		t.Fatalf("expected Papers to be empty, got %d", len(stack.Papers))
	}

	if stack.IsPublic {
		t.Fatal("exptected IsPublic to be false by default")
	}

	t.Skip("service not fully implemented")
}

func TestServiceDeleteReturnsErrorForUnknownStack(t *testing.T) {
	t.Parallel()
	t.Skip("service not implemented yet")
}

func TestServiceCreateRejectsDuplicateStackNameForSameUser(t *testing.T) {
	t.Parallel()
	t.Skip("service not implemented yet")
}
