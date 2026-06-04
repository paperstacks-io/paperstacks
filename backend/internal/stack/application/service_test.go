package application

import (
	"context"
	"testing"

	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/repository/memory"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

func TestServiceCreateNormalizesAndValidatesStack(t *testing.T) {
	t.Parallel()

	service := NewStackService(memory.NewRepository())
	stack := domain.NewStack(" Normalized Stack ", userDomain.User{ExternalID: "0", Email: "testUser@example.com"})
	err := service.Create(context.Background(), *stack)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	created, err := service.GetByUUID(context.Background(), stack.UUID)
	if err != nil {
		t.Fatalf("Create() stack not retrieve via GetByUUID()")
	}

	if created.Name != "Normalized Stack" {
		t.Fatalf("Create() did not store the normalized stack name")
	}

	if created.CreatedAt.IsZero() || stack.UpdatedAt.IsZero() {
		t.Fatalf("Create() did not set CreatedAt timestamp")
	}

	if created.CreatedAt != stack.UpdatedAt {
		t.Fatalf("Create() should not set UpdatedAt timestamp")
	}

	if created.Owner.ExternalID != "0" {
		t.Fatalf("Create() did not associate the stack with the correct user")
	}
}

func TestGetByUUIDError(t *testing.T) {
	t.Parallel()

	service := NewStackService(memory.NewRepository())

	_, err := service.GetByUUID(context.Background(), "unkown")
	if err != domain.ErrStackNotFound {
		t.Fatalf("GetByUUID() did not return an error for unknown UUID")
	}
}

func TestServiceUpdateModifiesUpdateAt(t *testing.T) {
	t.Parallel()

	service := NewStackService(memory.NewRepository())
	stack := domain.NewStack("Update Test Stack", userDomain.User{ExternalID: "0", Email: "testUser@example.com"})
	originalUpdatedAt := stack.UpdatedAt
	err := service.Create(context.Background(), *stack)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	modified := *stack
	modified.Name = "Updated Stack Name"
	updated, err := service.Update(context.Background(), modified)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if updated.UpdatedAt.Equal(originalUpdatedAt) {
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

func TestServiceGetByUUIDTrimsUUID(t *testing.T) {
	t.Parallel()

	service := NewStackService(memory.NewRepository())

	stack, err := service.GetByUUID(context.Background(), " 9e1a819a-24ab-47b6-be29-92b49325e4c2 ")
	if err != nil {
		t.Fatalf("GetByUUID() error = %v", err)
	}

	if stack.UUID != "9e1a819a-24ab-47b6-be29-92b49325e4c2" {
		t.Fatalf("GetByUUID() UUID = %s, want %s", stack.UUID, "9e1a819a-24ab-47b6-be29-92b49325e4c2")
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

func TestServiceCreateInvalidStackError(t *testing.T) {
	t.Parallel()

	service := NewStackService(memory.NewRepository())
	stack := domain.NewStack("", userDomain.User{})

	err := service.Create(context.Background(), *stack)

	if err != domain.ErrInvalidStack {
		t.Fatalf("Create() error = %v, want %v", err, domain.ErrInvalidStack)
	}
}

func TestServiceUpdateInvalidStackError(t *testing.T) {
	t.Parallel()

	service := NewStackService(memory.NewRepository())
	stack := domain.NewStack("", userDomain.User{})

	_, err := service.Update(context.Background(), *stack)
	if err != domain.ErrInvalidStack {
		t.Fatalf("Update() error = %v, want %v", err, domain.ErrInvalidStack)
	}
}

func TestServiceListReturnsInvalidUserError(t *testing.T) {
	t.Parallel()

	service := NewStackService(memory.NewRepository())

	_, err := service.List(context.Background(), "")
	if err == nil {
		t.Fatalf("List() expexted error but got nil")
	}
}

func TestServiceListPublicReturnsInvalidUserError(t *testing.T) {
	t.Parallel()

	service := NewStackService(memory.NewRepository())

	_, err := service.ListPublic(context.Background(), "")
	if err == nil {
		t.Fatalf("ListPublic() expected error but got nil")
	}
}

func TestServiceAddPaperAddsPaperToStack(t *testing.T) {
	t.Parallel()

	service := NewStackService(memory.NewRepository())

	paper := paperDomain.Paper{
		UUID: "36583bb4-8cdc-554e-bcf5-f67b60d0b290",
	}

	err := service.AddPaper(context.Background(), " 9e1a819a-24ab-47b6-be29-92b49325e4c2 ", paper)
	if err != nil {
		t.Fatalf("AddPaper() error = %v", err)
	}

	stack, err := service.GetByUUID(context.Background(), "9e1a819a-24ab-47b6-be29-92b49325e4c2")
	if err != nil {
		t.Fatalf("GetByUUID() error = %v", err)
	}

	if len(stack.Papers) != 1 {
		t.Fatalf("Stack has %d papers, want %d", len(stack.Papers), 1)
	}

	if stack.Papers[0].UUID != paper.UUID {
		t.Fatalf("Paper UUID = %s, want %s", stack.Papers[0].UUID, paper.UUID)
	}
}

func TestServiceRemovePaperRemovesPaperFromStack(t *testing.T) {
	t.Parallel()

	service := NewStackService(memory.NewRepository())

	err := service.AddPaper(context.Background(), "9e1a819a-24ab-47b6-be29-92b49325e4c2", paperDomain.Paper{
		UUID: "36583bb4-8cdc-554e-bcf5-f67b60d0b290",
	})
	if err != nil {
		t.Fatalf("AddPaper() error = %v", err)
	}

	err = service.RemovePaper(
		context.Background(),
		" 9e1a819a-24ab-47b6-be29-92b49325e4c2 ",
		" 36583bb4-8cdc-554e-bcf5-f67b60d0b290 ",
	)
	if err != nil {
		t.Fatalf("RemovePaper() error = %v", err)
	}

	stack, err := service.GetByUUID(context.Background(), "9e1a819a-24ab-47b6-be29-92b49325e4c2")
	if err != nil {
		t.Fatalf("GetByUUID() error = %v", err)
	}

	if len(stack.Papers) != 0 {
		t.Fatalf("Stack has %d papers, want %d", len(stack.Papers), 0)
	}
}
