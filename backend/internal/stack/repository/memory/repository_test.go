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
		UUID: "da572e9d-4d1d-4c17-9034-b3f0fbc6cdf1",
		Name: "Delete Stack",
		Owner: userDomain.User{
			ExternalID: "0",
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err = repo.Delete(context.Background(), "da572e9d-4d1d-4c17-9034-b3f0fbc6cdf1")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestRepositoryGetByUUIDReturnsStack(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	stack, err := repo.GetByUUID(context.Background(), "9e1a819a-24ab-47b6-be29-92b49325e4c2")
	if err != nil {
		t.Fatalf("GetByUUID() error = %v", err)
	}

	if stack.UUID != "9e1a819a-24ab-47b6-be29-92b49325e4c2" {
		t.Fatalf("GetByUUID() UUID = %s, want %s", stack.UUID, "9e1a819a-24ab-47b6-be29-92b49325e4c2")
	}
}

func TestRepositoryGetByUUIDReturnsNotFound(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	_, err := repo.GetByUUID(context.Background(), "unknown-stack")
	if err != domain.ErrStackNotFound {
		t.Fatalf("GetByUUID() error = %v, want %v", err, domain.ErrStackNotFound)
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
