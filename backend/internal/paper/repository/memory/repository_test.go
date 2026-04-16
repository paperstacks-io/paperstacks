package memory

import (
	"context"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

func TestRepositoryGetByUUID(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	ctx := context.Background()
	uuid := "ac4345f6-da8c-54c2-9365-73f9bdee8cad"
	res, err := repo.GetByUUID(ctx, uuid)
	if err != nil {
		t.Fatalf("GetByUUID() err = %v, want nil", err)
	}

	if res.UUID != uuid {
		t.Fatalf("GetByUUID() = %q, want %q", res.UUID, uuid)
	}

	uuid = "458a3f14-7242-5d63-bcd0-a37c79e4856a"
	res, err = repo.GetByUUID(ctx, uuid)
	if err != nil {
		t.Fatalf("GetByUUID() err = %v, want nil", err)
	}

	if res.UUID != uuid {
		t.Fatalf("GetByUUID() = %q, want %q", res.UUID, uuid)
	}
}

func TestRepositorySaveReturnsAlreadyExists(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	_, err := repo.Save(context.Background(), domain.Paper{
		UUID:  "36583bb4-8cdc-554e-bcf5-f67b60d0b290",
		Title: "duplicate",
	})
	if err != domain.ErrPaperAlreadyExists {
		t.Fatalf("Save() error = %v, want %v", err, domain.ErrPaperAlreadyExists)
	}
}

func TestRepositorySearchByTitle(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	result, err := repo.Search(context.Background(), domain.SearchOptions{Query: "exploratory testing: a multiple"})
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("Search() returned %d papers, want 1", len(result.Items))
	}
	if result.Items[0].DOI != "10.1109/ISESE.2005.1541817" {
		t.Fatalf("Search() DOI = %q, want %q", result.Items[0].DOI, "10.1109/isese.2005.1541817")
	}
}

func TestRepositorySearchByKeyword(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	result, err := repo.Search(context.Background(), domain.SearchOptions{Query: "gui testing"})
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if len(result.Items) != 21 {
		t.Fatalf("Search() returned %d papers, want 21", len(result.Items))
	}
}

func TestRepositorySearchEmptyQueryReturnsAllPapers(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	result, err := repo.Search(context.Background(), domain.SearchOptions{})

	expected := 326
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

func TestRepositorySearchSortByYearDescending(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	result, err := repo.Search(context.Background(), domain.SearchOptions{SortBy: "year", Desc: true})
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if len(result.Items) != 326 {
		t.Fatalf("Search() returned %d papers, want 326", len(result.Items))
	}
	if result.Items[0].DOI != "10.1016/j.infsof.2025.107723" {
		t.Fatalf("Search() first DOI = %q, want %q", result.Items[0].DOI, "10.1016/j.infsof.2025.107723")
	}
}

func TestRepositorySearchPaginatesResults(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	result, err := repo.Search(context.Background(), domain.SearchOptions{SortBy: "title", Page: 2, PageSize: 10})
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}

	if result.Total != 326 {
		t.Fatalf("Search() total = %d, want 326", result.Total)
	}
	if len(result.Items) != 10 {
		t.Fatalf("Search() items = %d, want 10", len(result.Items))
	}
	if result.Page != 2 {
		t.Fatalf("Search() page = %d, want 2", result.Page)
	}
	if result.PageSize != 10 {
		t.Fatalf("Search() pageSize = %d, want 10", result.PageSize)
	}
	if !result.HasNext {
		t.Fatalf("Search() hasNext = false, want true")
	}
}

func TestRepositorySearchNoResultReturnsEmpty(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	result, err := repo.Search(context.Background(), domain.SearchOptions{Query: "nonexistent title"})
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if len(result.Items) != 0 {
		t.Fatalf("Search() result = %v, want empty slice", len(result.Items))
	}
}
