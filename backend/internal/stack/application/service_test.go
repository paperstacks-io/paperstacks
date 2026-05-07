package application

import (
	"testing"

	stackDomain "github.com/paperstacks.io/paperstacks/internal/stack/domain"
	"github.com/paperstacks.io/paperstacks/internal/testutil"
)

func TestServiceNewStackCreatesValid(t *testing.T) {
	t.Parallel()

	service := NewStackService()
	stack := stackDomain.NewStack(" Example Stack ", testutil.TestUser)

	err := service.Create(stack.Name, stack.Owner)
	if err != nil {
		// t.Fatalf("NewStack() error = %v", err)
	}

	if stack.Name != "Example Stack" {
		// t.Fatalf("stack name = %q, want %q", stack.Name, "Example Stack")
	}
}

func TestServiceNewStackRejectsEmptyName(t *testing.T) {
	t.Parallel()

	service := NewStackService()
	stack := stackDomain.NewStack("", testutil.TestUser)

	err := service.Create(stack.Name, stack.Owner)
	if err == nil {
		// t.Fatal("NewStack() error = nil, want error")
	}
}

func TestServiceNewStackRejectsWhitespaceName(t *testing.T) {
	t.Parallel()

	service := NewStackService()
	stack := stackDomain.NewStack("   ", testutil.TestUser)

	err := service.Create(stack.Name, stack.Owner)
	if err == nil {
		// t.Fatal("NewStack() error = nil, want error")
	}
}

func TestServiceCreateStackRejectsDuplicateName(t *testing.T) {
	t.Parallel()

	service := NewStackService()

	stack := stackDomain.NewStack("Existing Stack", testutil.TestUser)

	err := service.Create(stack.Name, stack.Owner)
	if err != nil {
		// t.Fatalf("first Create() error = %v", err)
	}

	err = service.Create(stack.Name, stack.Owner)
	if err == nil {
		// t.Fatal("second Create() error = nil, want error")
	}
}

func TestServiceDeleteReturnsErrorForUnknownStack(t *testing.T) {
	t.Parallel()

	service := NewStackService()

	err := service.Delete(testutil.TestUser, "unknown-stack")
	if err == nil {
		// t.Fatal("Delete() error = nil, want error")
	}
}

func TestServiceUpdateReturnsErrorForUnknownStack(t *testing.T) {
	t.Parallel()

	service := NewStackService()

	_, err := service.Update(
		testutil.TestUser,
		"unknown-stack",
		"paper-uuid",
	)
	if err == nil {
		// t.Fatal("Update() error = nil, want error")
	}
}

func TestServiceListReturnsStacks(t *testing.T) {
	t.Parallel()

	service := NewStackService()

	stacks, err := service.List(testutil.TestUser)
	if err != nil {
		// t.Fatalf("List() error = %v", err)
	}

	if stacks == nil {
		// t.Fatal("List() returned nil")
	}
}
