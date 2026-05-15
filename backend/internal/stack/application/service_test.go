package application

import (
	"context"
	"slices"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/repository/memory"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

func TestServiceCreateNormalizesAndValidatesStack(t *testing.T) {
	t.Parallel()

	service := NewStackService(memory.NewRepository())
	stack := domain.NewStack(" Normalized Stack ", userDomain.User{ExternalID: "0"})
	err := service.Create(context.Background(), *stack)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	stacks, _ := service.List(context.Background(), userDomain.User{ExternalID: "0"})
	idx := slices.IndexFunc(stacks, func(s domain.Stack) bool {
		return s.Name == "Normalized Stack"
	})

	if idx == -1 {
		t.Fatalf("Create() did not store the normalized stack name")
	}

	stack = &stacks[idx]
	if stack.UUID == "" {
		t.Fatalf("Create() did not generate a UUID for the stack")
	}

	if stack.CreatedAt.IsZero() || stack.UpdatedAt.IsZero() {
		t.Fatalf("Create() did not set CreatedAt timestamp")
	}

	if stack.CreatedAt != stack.UpdatedAt {
		t.Fatalf("Create() should not set UpdatedAt timestamp")
	}

	if stack.Owner.ExternalID != "0" {
		t.Fatalf("Create() did not associate the stack with the correct user")
	}
}

func TestServiceUpdateModifiesUpdateAt(t *testing.T) {
	t.Parallel()

	service := NewStackService(memory.NewRepository())
	stack := domain.NewStack("Update Test Stack", userDomain.User{ExternalID: "0"})
	err := service.Create(context.Background(), *stack)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	stacks, _ := service.List(context.Background(), userDomain.User{ExternalID: "0"})
	idx := slices.IndexFunc(stacks, func(s domain.Stack) bool {
		return s.Name == "Update Test Stack"
	})

	if idx == -1 {
		t.Fatalf("Create() did not store the stack for update test")
	}

	stack = &stacks[idx]
	originalUpdatedAt := stack.UpdatedAt

	modified := *stack
	modified.Name = "Updated Stack Name"
	updatedStack, err := service.Update(context.Background(), modified)

	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if updatedStack.UpdatedAt.Equal(originalUpdatedAt) {
		t.Fatalf("Update() did not modify UpdatedAt timestamp")
	}
}

func TestServiceDeleteReturnsErrorForUnknownStack(t *testing.T) {
	t.Parallel()

	service := NewStackService(memory.NewRepository())

	err := service.Delete(context.Background(), "unknown-stack")
	if err == nil {
		t.Fatalf("Delete() did not return an error for unknown stack")
	}
}

func TestServiceCreateRejectsDuplicateStackNameForSameUser(t *testing.T) {
	t.Parallel()

	service := NewStackService(memory.NewRepository())
	user := userDomain.User{ExternalID: "0"}
	stack := domain.NewStack("Duplicate Stack", user)

	err := service.Create(context.Background(), *stack)
	if err != nil {
		t.Fatalf("Create() first stack error = %v", err)
	}

	duplicate := domain.NewStack(" Duplicate Stack ", user)

	err = service.Create(context.Background(), *duplicate)
	if err == nil {
		t.Fatalf("Create() expected error for duplicate stack name, got nil")
	}
}

func TestServiceListReturnsInvalidUserError(t *testing.T) {
	t.Parallel()

	service := NewStackService(memory.NewRepository())

	_, err := service.List(context.Background(), userDomain.User{ExternalID: ""})
	if err != userDomain.ErrInvalidUser {
		t.Fatalf("List() error = %v, want %v", err, userDomain.ErrInvalidUser)
	}
}

func TestServiceCreateReturnsInvalidStackError(t *testing.T) {
	t.Parallel()

	service := NewStackService(memory.NewRepository())

	stack := domain.NewStack("Invalid Owner Stack", userDomain.User{ExternalID: ""})

	err := service.Create(context.Background(), *stack)
	if err != domain.ErrInvalidStack {
		t.Fatalf("Create() error = %v, want %v", err, domain.ErrInvalidStack)
	}

	_, err = service.Update(context.Background(), *stack)
	if err != domain.ErrInvalidStack {
		t.Fatalf("Update() error = %v, want %v", err, domain.ErrInvalidStack)
	}
}
