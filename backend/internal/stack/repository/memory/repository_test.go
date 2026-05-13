package memory

import (
	"context"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

func TestRepositoryCreateReturnsAlreadyExists(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	err := repo.Create(context.Background(), domain.Stack{
		Name: "Example Stack",
		Owner: userDomain.User{
			ExternalID: "0",
		},
	})

	if err != domain.ErrStackAlreadyExists {
		t.Fatalf("Create() error = %v, want %v", err, domain.ErrStackAlreadyExists)
	}
}

func TestRepositoryUpdateReturnsNotFound(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	_, err := repo.Update(context.Background(), domain.Stack{
		Name: "Nonexistent Stack",
		Owner: userDomain.User{
			ExternalID: "0",
		},
	})

	if err != domain.ErrStackNotFound {
		t.Fatalf("Update() error = %v, want %v", err, domain.ErrStackNotFound)
	}
}

func TestRepositoryCreateAndDelete(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	err := repo.Create(context.Background(), domain.Stack{
		UUID: "Test-UUID-01",
		Name: "Delete Stack",
		Owner: userDomain.User{
			ExternalID: "0",
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err = repo.Delete(context.Background(), "Test-UUID-01")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestRepositoryListReturnsAllUserStacks(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	stacks, err := repo.List(context.Background(), userDomain.User{ExternalID: "0"})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(stacks) != 1 {
		t.Fatalf("List() returned %d stacks, want %d", len(stacks), 1)
	}

	for _, stack := range stacks {
		if stack.Owner.ExternalID != "0" {
			t.Fatalf("List() returned stack with owner %s, want %s", stack.Owner.ExternalID, "0")
		}
	}
}
