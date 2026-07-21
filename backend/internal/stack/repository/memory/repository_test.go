package memory

import (
	"context"
	"reflect"
	"testing"

	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	paperMemory "github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

var user = userDomain.User{
	ExternalID: "dbe3febc-ab91-486c-b51f-38ab0f59a4d9",
}

func TestSeedDataUsesPaperSeedData(t *testing.T) {
	t.Parallel()

	papersByUUID := make(map[string]paperDomain.Paper)
	for _, paper := range paperMemory.SeedData() {
		papersByUUID[paper.UUID] = paper
	}

	for _, stack := range seedData() {
		for _, paper := range stack.Papers {
			want, ok := papersByUUID[paper.UUID]
			if !ok {
				t.Fatalf("seedData() stack %s contains paper %s missing from paper seed data", stack.UUID, paper.UUID)
			}
			if !reflect.DeepEqual(paper, want) {
				t.Fatalf("seedData() stack %s paper %s = %#v, want %#v", stack.UUID, paper.UUID, paper, want)
			}
		}
	}
}

func TestRepositoryCreateReturnsAlreadyExists(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	ctx := context.Background()

	stackOne := domain.NewStack("already-exists Stack", user)
	stackTwo := domain.NewStack("already-exists Stack", user)
	err := repo.Create(ctx, *stackOne)
	if err != nil {
		t.Fatalf("Create() first stack error = %v", err)
	}

	err = repo.Create(ctx, *stackTwo)
	if err != domain.ErrStackAlreadyExists {
		t.Fatalf("Create() second stack error = %v, want %v", err, domain.ErrStackAlreadyExists)
	}

	if err != domain.ErrStackAlreadyExists {
		t.Fatalf("Create() error = %v, want %v", err, domain.ErrStackAlreadyExists)
	}
}

func TestRepositoryUpdateReturnsNotFound(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	_, err := repo.Update(context.Background(), domain.Stack{
		Name:  "non-existent Stack",
		Owner: user,
	})

	if err != domain.ErrStackNotFound {
		t.Fatalf("Update() error = %v, want %v", err, domain.ErrStackNotFound)
	}
}

func TestRepositoryCreateAndDelete(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	err := repo.Create(context.Background(), domain.Stack{
		UUID:  "da572e9d-4d1d-4c17-9034-b3f0fbc6cdf1",
		Name:  "Delete Stack",
		Owner: user,
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

	stacks, err := repo.List(context.Background(), user.ExternalID)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(stacks) != len(seedData()) {
		t.Fatalf("List() returned %d stacks, want %d", len(stacks), len(seedData()))
	}

	for _, stack := range stacks {
		if stack.Owner.ExternalID != user.ExternalID {
			t.Fatalf("List() returned stack with owner %s, want %s", stack.Owner.ExternalID, user.ExternalID)
		}
	}
}

func TestRepositoryListPublicReturnsOnlyPublicUserStacks(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	publicStackOne := domain.Stack{
		UUID:     "da572e9d-4d1d-4c17-9034-b3f0fbc6cdf1",
		Name:     "Public Stack",
		Owner:    user,
		IsPublic: true,
	}
	publicStackTwo := domain.Stack{
		UUID:     "cc92837a-d280-42cb-a689-ea58a46cdb4b",
		Name:     "Other Public Stack",
		Owner:    user,
		IsPublic: true,
	}
	privateStack := domain.Stack{
		UUID:     "873be0a7-3568-40c5-b2a2-63b3b8fa41d1",
		Name:     "Private Stack",
		Owner:    user,
		IsPublic: false,
	}

	if err := repo.Create(context.Background(), publicStackOne); err != nil {
		t.Fatalf("Create() public stack error = %v", err)
	}
	if err := repo.Create(context.Background(), publicStackTwo); err != nil {
		t.Fatalf("Create() other user stack error = %v", err)
	}
	if err := repo.Create(context.Background(), privateStack); err != nil {
		t.Fatalf("Create() private stack error = %v", err)
	}

	stacks, err := repo.ListPublic(context.Background(), user.ExternalID)
	if err != nil {
		t.Fatalf("ListPublic() error = %v", err)
	}

	expected := countPublicSeedStacks() + 2
	if len(stacks) != expected {
		t.Fatalf("ListPublic() returned %d stacks, want %d", len(stacks), expected)
	}
	if stacks[len(stacks)-2].UUID != publicStackOne.UUID {
		t.Fatalf("ListPublic() UUID = %s, want %s", stacks[len(stacks)-2].UUID, publicStackOne.UUID)
	}
	if stacks[len(stacks)-1].UUID != publicStackTwo.UUID {
		t.Fatalf("ListPublic() UUID = %s, want %s", stacks[len(stacks)-1].UUID, publicStackTwo.UUID)
	}
}

func TestRepositoryAddPaperDoesNotAddDuplicatePaper(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	initialStack, err := repo.GetByUUID(context.Background(), "9e1a819a-24ab-47b6-be29-92b49325e4c2")
	if err != nil {
		t.Fatalf("GetByUUID() initial error = %v", err)
	}
	initialPaperCount := len(initialStack.Papers)

	err = repo.AddPaper(context.Background(), "9e1a819a-24ab-47b6-be29-92b49325e4c2", paperDomain.Paper{
		UUID: "36583bb4-8cdc-554e-bcf5-f67b60d0b290",
	})
	if err != nil {
		t.Fatalf("AddPaper() first add error = %v", err)
	}

	err = repo.AddPaper(context.Background(), "9e1a819a-24ab-47b6-be29-92b49325e4c2", paperDomain.Paper{
		UUID: "36583bb4-8cdc-554e-bcf5-f67b60d0b290",
	})
	if err != nil {
		t.Fatalf("AddPaper() second add error = %v", err)
	}

	stack, err := repo.GetByUUID(context.Background(), "9e1a819a-24ab-47b6-be29-92b49325e4c2")
	if err != nil {
		t.Fatalf("GetByUUID() error = %v", err)
	}

	if len(stack.Papers) != initialPaperCount+1 {
		t.Fatalf("Stack has %d papers, want %d", len(stack.Papers), initialPaperCount+1)
	}
}

func TestRepositoryAddPaperReturnsStackNotFound(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	err := repo.AddPaper(context.Background(), "unknown-stack", paperDomain.Paper{
		UUID: "36583bb4-8cdc-554e-bcf5-f67b60d0b290",
	})
	if err != domain.ErrStackNotFound {
		t.Fatalf("AddPaper() error = %v, want %v", err, domain.ErrStackNotFound)
	}
}

func TestRepositoryRemovePaperReturnsNotFound(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	err := repo.RemovePaper(context.Background(), "unknown-stack", "paper-1")
	if err != domain.ErrStackNotFound {
		t.Fatalf("RemovePaper() error = %v, want %v", err, domain.ErrStackNotFound)
	}
}

func TestRepositorySearchByName(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	publicStack := domain.Stack{
		UUID:     "2070926b-0afc-4471-8e2f-37a29fed20ea",
		Name:     "Unique Public Stack",
		Owner:    user,
		IsPublic: true,
	}

	if err := repo.Create(context.Background(), publicStack); err != nil {
		t.Fatalf("Create() public stack error = %v", err)
	}

	result, err := repo.SearchPublic(context.Background(), domain.SearchOptions{
		Query: "  Unique Public Stack  ",
	})
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}

	if result.Total != 1 {
		t.Fatalf("Search() total = %d, want %d", result.Total, 1)
	}
	if len(result.Items) != 1 {
		t.Fatalf("Search() returned %d stacks, want %d", len(result.Items), 1)
	}
	if result.Items[0].UUID != publicStack.UUID {
		t.Fatalf("Search() UUID = %s, want %s", result.Items[0].UUID, publicStack.UUID)
	}
}

func TestRepositorySearchEmptyQueryReturnsAllStacks(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	stacks := []domain.Stack{
		{
			UUID:     "34cef86a-c369-4f61-9ac1-6c4ca09f50f3",
			Name:     "Public Stack One",
			Owner:    user,
			IsPublic: true,
		},
		{
			UUID:     "2051a4d9-23f6-4bfa-8d44-367c28760198",
			Name:     "Public Stack Two",
			Owner:    user,
			IsPublic: true,
		},
		{
			UUID:     "bd9ee496-7381-4c69-a7ba-65bc38010af4",
			Name:     "Public Stack Three",
			Owner:    user,
			IsPublic: false,
		},
	}

	for _, stack := range stacks {
		if err := repo.Create(context.Background(), stack); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	result, err := repo.SearchPublic(context.Background(), domain.SearchOptions{})

	expected := countPublicSeedStacks() + 2
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if result.Total != expected {
		t.Fatalf("Search() total = %d, want %d", result.Total, expected)
	}
	if len(result.Items) != expected {
		t.Fatalf("Search() items = %d, want %d", len(result.Items), expected)
	}
}

func TestRepositoryStatsByOwner(t *testing.T) {
	t.Parallel()

	otherUser := userDomain.User{ExternalID: "other-user"}
	repo := &Repository{
		data: []domain.Stack{
			{
				UUID:     "34cef86a-c369-4f61-9ac1-6c4ca09f50f3",
				Name:     "Public Stack One",
				Owner:    user,
				IsPublic: true,
				Papers: []paperDomain.Paper{
					{UUID: "c8b0970b-114a-44cb-97fb-1a83297cc46d"},
					{UUID: "8eca1e47-a689-457b-b2ce-98bd5c7ee565"},
				},
			},
			{
				UUID:     "2051a4d9-23f6-4bfa-8d44-367c28760198",
				Name:     "Private Stack",
				Owner:    user,
				IsPublic: false,
				Papers: []paperDomain.Paper{
					{UUID: "c7633fde-474a-423d-9863-e78449f14a3c"},
				},
			},
			{
				UUID:     "bd9ee496-7381-4c69-a7ba-65bc38010af4",
				Name:     "Other User Stack",
				Owner:    otherUser,
				IsPublic: true,
				Papers: []paperDomain.Paper{
					{UUID: "a376c7e7-d614-4cd9-b348-eed9207c549b"},
				},
			},
		},
	}

	stats, err := repo.StatsByOwner(context.Background(), user.ExternalID)
	if err != nil {
		t.Fatalf("StatsByOwner() error = %v", err)
	}

	if stats.TotalStacks != 2 {
		t.Fatalf("StatsByOwner() TotalStacks = %d, want %d", stats.TotalStacks, 2)
	}
	if stats.PublicStacks != 1 {
		t.Fatalf("StatsByOwner() PublicStacks = %d, want %d", stats.PublicStacks, 1)
	}
	if stats.TotalPapers != 3 {
		t.Fatalf("StatsByOwner() TotalPapers = %d, want %d", stats.TotalPapers, 3)
	}
}

func countPublicSeedStacks() int {
	count := 0
	for _, stack := range seedData() {
		if stack.IsPublic {
			count++
		}
	}
	return count
}

func TestRepositorySearchPaginatesResults(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	stacks := []domain.Stack{
		{
			UUID:     "a13057e3-f6d8-4b4e-9230-fc281af13e33",
			Name:     "pagination-unique Stack One",
			Owner:    user,
			IsPublic: true,
		},
		{
			UUID:     "ce26475e-4ba2-469f-9346-9fda30161e92",
			Name:     "pagination-unique Stack Two",
			Owner:    user,
			IsPublic: true,
		},
		{
			UUID:     "122697ab-95c1-4502-aac0-d755631b8767",
			Name:     "pagination-unique Stack Three",
			Owner:    user,
			IsPublic: true,
		},
	}

	for _, stack := range stacks {
		if err := repo.Create(context.Background(), stack); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	result, err := repo.SearchPublic(context.Background(), domain.SearchOptions{
		Query: "pagination-unique",
	})
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}

	if result.Total != 3 {
		t.Fatalf("Search() total = %d, want %d", result.Total, 3)
	}
	if result.Page != 1 {
		t.Fatalf("Search() page = %d, want %d", result.Page, 1)
	}
	if result.PageSize != 3 {
		t.Fatalf("Search() pageSize = %d, want %d", result.PageSize, 3)
	}
	if result.HasNext {
		t.Fatalf("Search() hasNext = true, want false")
	}
}

func TestRepositorySearchSortByUpdatedAtDescending(t *testing.T) {
	t.Parallel()

	repo := NewRepository()
	ctx := context.Background()

	stackOne := domain.NewStack("sort-created Stack One", user)
	stackOne.IsPublic = true

	stackTwo := domain.NewStack("sort-created Stack Two", user)
	stackTwo.IsPublic = true

	stackThree := domain.NewStack("sort-created Stack Three", user)
	stackThree.IsPublic = true

	stacks := []domain.Stack{
		*stackOne,
		*stackTwo,
		*stackThree,
	}

	for _, stack := range stacks {
		if err := repo.Create(ctx, stack); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	result, err := repo.SearchPublic(ctx, domain.SearchOptions{
		Query:  "sort-created",
		SortBy: "updated_at",
		Desc:   true,
	})
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}

	if result.Total != 3 {
		t.Fatalf("Search() returned %d stacks, want %d", len(result.Items), 3)
	}

	if result.Items[0].Name != "sort-created Stack Three" {
		t.Fatalf("Search() first stack = %q, want %q", result.Items[0].Name, "sort-created Stack Three")
	}
	if result.Items[1].Name != "sort-created Stack Two" {
		t.Fatalf("Search() second stack = %q, want %q", result.Items[1].Name, "sort-created Stack Two")
	}
	if result.Items[2].Name != "sort-created Stack One" {
		t.Fatalf("Search() third stack = %q, want %q", result.Items[2].Name, "sort-created Stack One")
	}
}
