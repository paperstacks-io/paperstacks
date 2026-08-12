package application

import (
	"context"
	"testing"

	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/repository/memory"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

const existingPaperUUID = "36583bb4-8cdc-554e-bcf5-f67b60d0b290"

type fakePaperGetter struct {
	papers map[string]paperDomain.Paper
}

func (f fakePaperGetter) GetByUUID(ctx context.Context, uuid string) (paperDomain.Paper, error) {
	paper, ok := f.papers[uuid]
	if !ok {
		return paperDomain.Paper{}, paperDomain.ErrPaperNotFound
	}

	return paper, nil
}

func newTestStackService() *StackService {
	return NewStackService(memory.NewRepository(), fakePaperGetter{
		papers: map[string]paperDomain.Paper{
			existingPaperUUID: {
				UUID:  existingPaperUUID,
				Title: "Existing Paper",
			},
		},
	})
}

type duplicateDefaultStackRepository struct {
	domain.Repository
}

func (duplicateDefaultStackRepository) List(context.Context, string) ([]domain.Stack, error) {
	return nil, nil
}

func (duplicateDefaultStackRepository) Create(context.Context, domain.Stack) error {
	return domain.ErrStackAlreadyExists
}

func TestServiceCreateNormalizesAndValidatesStack(t *testing.T) {
	t.Parallel()

	service := newTestStackService()
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

func TestEnsureDefaultCreatesOnePrivateEmptyStack(t *testing.T) {
	t.Parallel()

	service := newTestStackService()
	user := userDomain.User{ExternalID: "default-user", Email: "default@example.com"}

	if err := service.EnsureDefault(context.Background(), user); err != nil {
		t.Fatalf("EnsureDefault() error = %v", err)
	}
	if err := service.EnsureDefault(context.Background(), user); err != nil {
		t.Fatalf("second EnsureDefault() error = %v", err)
	}

	stacks, err := service.List(context.Background(), user.ExternalID)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(stacks) != 1 {
		t.Fatalf("List() returned %d stacks, want 1", len(stacks))
	}

	stack := stacks[0]
	if stack.Name != defaultStackName {
		t.Fatalf("default stack name = %q, want %q", stack.Name, defaultStackName)
	}
	if stack.Owner != user {
		t.Fatalf("default stack owner = %#v, want %#v", stack.Owner, user)
	}
	if stack.IsPublic {
		t.Fatal("default stack is public, want private")
	}
	if len(stack.Papers) != 0 {
		t.Fatalf("default stack papers = %#v, want empty", stack.Papers)
	}
}

func TestEnsureDefaultAcceptsConcurrentCreation(t *testing.T) {
	t.Parallel()

	service := NewStackService(duplicateDefaultStackRepository{Repository: memory.NewRepository()}, nil)

	if err := service.EnsureDefault(context.Background(), userDomain.User{ExternalID: "default-user"}); err != nil {
		t.Fatalf("EnsureDefault() error = %v, want nil", err)
	}
}

func TestGetByUUIDError(t *testing.T) {
	t.Parallel()

	service := newTestStackService()

	_, err := service.GetByUUID(context.Background(), "unkown")
	if err != domain.ErrStackNotFound {
		t.Fatalf("GetByUUID() did not return an error for unknown UUID")
	}
}

func TestServiceUpdateModifiesUpdateAt(t *testing.T) {
	t.Parallel()

	service := newTestStackService()
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

	service := newTestStackService()

	err := service.Delete(context.Background(), "unknown-stack")
	if err == nil {
		t.Fatalf("Delete() did not return an error for unknown stack")
	}
}

func TestServiceGetByUUIDTrimsUUID(t *testing.T) {
	t.Parallel()

	service := newTestStackService()

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

	service := newTestStackService()
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

func TestServiceCreateInvalidNameError(t *testing.T) {
	t.Parallel()

	service := newTestStackService()
	stack := domain.NewStack("", userDomain.User{})

	err := service.Create(context.Background(), *stack)

	if err != domain.ErrInvalidName {
		t.Fatalf("Create() error = %v, want %v", err, domain.ErrInvalidName)
	}
}

func TestServiceUpdateInvalidNameError(t *testing.T) {
	t.Parallel()

	service := newTestStackService()
	stack := domain.NewStack("", userDomain.User{})

	_, err := service.Update(context.Background(), *stack)
	if err != domain.ErrInvalidName {
		t.Fatalf("Update() error = %v, want %v", err, domain.ErrInvalidName)
	}
}

func TestServiceListReturnsInvalidUserError(t *testing.T) {
	t.Parallel()

	service := newTestStackService()

	_, err := service.List(context.Background(), "")
	if err == nil {
		t.Fatalf("List() expexted error but got nil")
	}
}

func TestServiceListPublicReturnsInvalidUserError(t *testing.T) {
	t.Parallel()

	service := newTestStackService()

	_, err := service.ListPublic(context.Background(), "")
	if err == nil {
		t.Fatalf("ListPublic() expected error but got nil")
	}
}

func TestServiceAddPaperAddsPaperToStack(t *testing.T) {
	t.Parallel()

	service := newTestStackService()

	initialStack, err := service.GetByUUID(context.Background(), "9e1a819a-24ab-47b6-be29-92b49325e4c2")
	if err != nil {
		t.Fatalf("GetByUUID() initial error = %v", err)
	}
	initialPaperCount := len(initialStack.Papers)

	err = service.AddPaper(context.Background(), "9e1a819a-24ab-47b6-be29-92b49325e4c2", existingPaperUUID)
	if err != nil {
		t.Fatalf("AddPaper() error = %v", err)
	}

	stack, err := service.GetByUUID(context.Background(), "9e1a819a-24ab-47b6-be29-92b49325e4c2")
	if err != nil {
		t.Fatalf("GetByUUID() error = %v", err)
	}

	if len(stack.Papers) != initialPaperCount+1 {
		t.Fatalf("Stack has %d papers, want %d", len(stack.Papers), initialPaperCount+1)
	}

	addedPaper := stack.Papers[len(stack.Papers)-1]
	if addedPaper.UUID != existingPaperUUID {
		t.Fatalf("Paper UUID = %s, want %s", addedPaper.UUID, existingPaperUUID)
	}

	if addedPaper.Title != "Existing Paper" {
		t.Fatalf("Paper title = %s, want %s", addedPaper.Title, "Existing Paper")
	}
}

func TestServiceAddPaperReturnsErrorForUnknownPaper(t *testing.T) {
	t.Parallel()

	service := newTestStackService()

	initialStack, err := service.GetByUUID(context.Background(), "9e1a819a-24ab-47b6-be29-92b49325e4c2")
	if err != nil {
		t.Fatalf("GetByUUID() initial error = %v", err)
	}
	initialPaperCount := len(initialStack.Papers)

	err = service.AddPaper(context.Background(), "9e1a819a-24ab-47b6-be29-92b49325e4c2", "unknown-paper")
	if err != paperDomain.ErrPaperNotFound {
		t.Fatalf("AddPaper() error = %v, want %v", err, paperDomain.ErrPaperNotFound)
	}

	stack, err := service.GetByUUID(context.Background(), "9e1a819a-24ab-47b6-be29-92b49325e4c2")
	if err != nil {
		t.Fatalf("GetByUUID() error = %v", err)
	}

	if len(stack.Papers) != initialPaperCount {
		t.Fatalf("Stack has %d papers, want %d", len(stack.Papers), initialPaperCount)
	}
}

func TestServiceRemovePaperRemovesPaperFromStack(t *testing.T) {
	t.Parallel()

	service := newTestStackService()

	initialStack, err := service.GetByUUID(context.Background(), "9e1a819a-24ab-47b6-be29-92b49325e4c2")
	if err != nil {
		t.Fatalf("GetByUUID() initial error = %v", err)
	}
	initialPaperCount := len(initialStack.Papers)

	err = service.AddPaper(context.Background(), "9e1a819a-24ab-47b6-be29-92b49325e4c2", existingPaperUUID)
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

	if len(stack.Papers) != initialPaperCount {
		t.Fatalf("Stack has %d papers, want %d", len(stack.Papers), initialPaperCount)
	}
}
