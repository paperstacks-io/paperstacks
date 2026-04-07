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
	uuid := "6752de78-0264-4ac5-8bd3-4eed7d3f5484"
	res, err := repo.GetByUUID(ctx, uuid)
	if err != nil {
		t.Fatalf("GetByUUID() err = %v, want nil", err)
	}

	if res.UUID != uuid {
		t.Fatalf("GetByUUID() = %q, want %q", res.UUID, uuid)
	}

	uuid = "202a0a80-dec8-48ea-b140-5fd26fee8dbd"
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

	err := repo.Save(context.Background(), domain.Paper{
		DOI:   "10.1109/isese.2005.1541817",
		Title: "duplicate",
	})
	if err != domain.ErrPaperAlreadyExists {
		t.Fatalf("Save() error = %v, want %v", err, domain.ErrPaperAlreadyExists)
	}
}

func TestRepositorySearchByTitle(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	result, err := repo.Search(context.Background(), domain.SearchOptions{Query: "exploratory testing"})
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("Search() returned %d papers, want 1", len(result.Items))
	}
	if result.Items[0].DOI != "10.1109/isese.2005.1541817" {
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
	if len(result.Items) != 3 {
		t.Fatalf("Search() returned %d papers, want 3", len(result.Items))
	}
}

func TestRepositorySearchEmptyQueryReturnsAllPapers(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	result, err := repo.Search(context.Background(), domain.SearchOptions{})
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if result.Total != 4 {
		t.Fatalf("Search() total = %d, want 4", result.Total)
	}
	if len(result.Items) != 4 {
		t.Fatalf("Search() items = %d, want 4", len(result.Items))
	}
}

func TestRepositorySearchSortByYearDescending(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	result, err := repo.Search(context.Background(), domain.SearchOptions{SortBy: "year", Desc: true})
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if len(result.Items) != 4 {
		t.Fatalf("Search() returned %d papers, want 4", len(result.Items))
	}
	if result.Items[0].DOI != "10.1007/s10664-024-10522-z" {
		t.Fatalf("Search() first DOI = %q, want %q", result.Items[0].DOI, "10.1007/s10664-024-10522-z")
	}
}

func TestRepositorySearchPaginatesResults(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	result, err := repo.Search(context.Background(), domain.SearchOptions{SortBy: "title", Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}

	if result.Total != 4 {
		t.Fatalf("Search() total = %d, want 4", result.Total)
	}
	if len(result.Items) != 2 {
		t.Fatalf("Search() items = %d, want 2", len(result.Items))
	}
	if result.Page != 2 {
		t.Fatalf("Search() page = %d, want 2", result.Page)
	}
	if result.PageSize != 2 {
		t.Fatalf("Search() pageSize = %d, want 2", result.PageSize)
	}
	if result.HasNext {
		t.Fatalf("Search() hasNext = true, want false")
	}

	if result.Items[0].DOI != "10.1109/isese.2005.1541817" {
		t.Fatalf("Search() first item DOI = %q, want %q", result.Items[0].DOI, "10.1109/isese.2005.1541817")
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
