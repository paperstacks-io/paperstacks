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

	stacks, err := repo.List(context.Background(), "0")
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

func TestRepositoryListPublicReturnsOnlyPublicUserStacks(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	publicStack := domain.Stack{
		UUID:     "da572e9d-4d1d-4c17-9034-b3f0fbc6cdf1",
		Name:     "Public Stack",
		Owner:    userDomain.User{ExternalID: "owner-1"},
		IsPublic: true,
	}
	privateStack := domain.Stack{
		UUID:     "873be0a7-3568-40c5-b2a2-63b3b8fa41d1",
		Name:     "Private Stack",
		Owner:    userDomain.User{ExternalID: "owner-1"},
		IsPublic: false,
	}
	otherUserStack := domain.Stack{
		UUID:     "cc92837a-d280-42cb-a689-ea58a46cdb4b",
		Name:     "Other User Stack",
		Owner:    userDomain.User{ExternalID: "owner-2"},
		IsPublic: true,
	}

	if err := repo.Create(context.Background(), publicStack); err != nil {
		t.Fatalf("Create() public stack error = %v", err)
	}
	if err := repo.Create(context.Background(), privateStack); err != nil {
		t.Fatalf("Create() private stack error = %v", err)
	}
	if err := repo.Create(context.Background(), otherUserStack); err != nil {
		t.Fatalf("Create() other user stack error = %v", err)
	}

	stacks, err := repo.ListPublic(context.Background(), "owner-1")
	if err != nil {
		t.Fatalf("ListPublic() error = %v", err)
	}

	if len(stacks) != 1 {
		t.Fatalf("ListPublic() returned %d stacks, want %d", len(stacks), 1)
	}
	if stacks[0].UUID != publicStack.UUID {
		t.Fatalf("ListPublic() UUID = %s, want %s", stacks[0].UUID, publicStack.UUID)
	}
}
