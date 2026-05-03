package memory

import (
	"context"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

func TestRepositoryListAllPublicStacks(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	stacks, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stacks) == 0 {
		t.Fatalf("expected at least %d public stack, got %d", 1, len(stacks))
	}

	for _, s := range stacks {
		if s.Visibility != domain.VisibilityPublic {
			t.Fatalf("expected public stack, got %v", s.Visibility)
		}
	}
}

func TestRepositoryListByOwner(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	stacks, err := repo.ListByOwner(context.Background(), "Andy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stacks) == 0 {
		t.Fatalf("expected at least one stack, got %d", len(stacks))
	}

	for _, s := range stacks {
		if s.Owner != "Andy" {
			t.Fatalf("expected owner Andy, got %s", s.Owner)
		}
	}
}

func TestRepositoryListPrivateStacksByOwner(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	stacks, err := repo.ListPrivateByOwner(context.Background(), "Andy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stacks) != 1 {
		t.Fatalf("expected 1 private stack, got %d", len(stacks))
	}

	for _, s := range stacks {
		if s.Owner != "Andy" {
			t.Fatalf("expected owner Andy, got %s", s.Owner)
		}
		if s.Visibility != domain.VisibilityPrivate {
			t.Fatalf("expected private stack, got %v", s.Visibility)
		}
	}
}

func TestRepositoryListPublicStacksByOwner(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	stacks, err := repo.ListPublicByOwner(context.Background(), "Andy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stacks) != 3 {
		t.Fatalf("expected %d public stacks, got %d", 3, len(stacks))
	}

	for _, s := range stacks {
		if s.Owner != "Andy" {
			t.Fatalf("expected owner Andy, got %s", s.Owner)
		}
		if s.Visibility != domain.VisibilityPublic {
			t.Fatalf("expected public stack, got %v", s.Visibility)
		}
	}
}

func TestRepositorySaveStackSuccess(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	stack := domain.Stack{
		UUID:       "test-UUID",
		Owner:      "Will",
		Name:       "Test Stack",
		Visibility: domain.VisibilityPublic,
	}

	err := repo.Save(context.Background(), stack)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stacks, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, s := range stacks {
		if s.UUID == "test-UUID" {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected stack to be saved")
	}
}

func TestRepositorySaveStackAlreadyExists(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	stack := domain.Stack{
		Owner:      "Will",
		Name:       "Agile & Software Quality Stack",
		Visibility: domain.VisibilityPublic,
	}

	err := repo.Save(context.Background(), stack)
	if err != domain.ErrStackAlreadyExists {
		t.Fatalf("expected error: %v", err)
	}
}

func TestRepositoryDeleteStackSuccess(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	stack := domain.Stack{UUID: "5"}

	err := repo.Delete(context.Background(), stack)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stacks, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, s := range stacks {
		if s.UUID == "5" {
			t.Fatalf("expected stack to be deleted")
		}
	}
}

func TestRepositoryDeleteStackNotFound(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	stack := domain.Stack{UUID: "10"}

	err := repo.Delete(context.Background(), stack)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrStackNotFound {
		t.Fatalf("expected ErrStackNotFound, got %v", err)
	}
}

func TestRepositorySetVisibilitySuccess(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	stack := domain.Stack{UUID: "6"}

	err := repo.SetVisibility(context.Background(), stack, domain.VisibilityPrivate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stacks, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, s := range stacks {
		if s.UUID == "6" {
			if s.Visibility != domain.VisibilityPrivate {
				t.Fatalf("expected visibility private, got %v", s.Visibility)
			}
		}
	}
}
