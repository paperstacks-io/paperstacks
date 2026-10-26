package memory

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

func TestPDFAttachmentPreservesPaperEdits(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository()
	paper := domain.Paper{UUID: "test-paper", Title: "Original", Authors: []domain.Author{{NameLast: "Author"}}, Metadata: domain.Metadata{Publisher: "Publisher"}}
	if _, err := repo.Save(ctx, paper); err != nil {
		t.Fatal(err)
	}
	if err := repo.AttachPDF(ctx, paper.UUID, "pdf/test.pdf"); err != nil {
		t.Fatal(err)
	}
	want := paper
	want.PDFStorageKey = "pdf/test.pdf"
	got, err := repo.GetByUUID(ctx, paper.UUID)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("after attach = %#v, %v; want %#v", got, err, want)
	}

	paper.Title = "Edited"
	if err := repo.Update(ctx, paper.UUID, paper); err != nil {
		t.Fatal(err)
	}
	want.Title = paper.Title
	got, err = repo.GetByUUID(ctx, paper.UUID)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("after stale edit = %#v, %v; want %#v", got, err, want)
	}
	if err := repo.DetachPDF(ctx, paper.UUID); err != nil {
		t.Fatal(err)
	}
	// A stale edit must not restore a removed attachment either.
	if err := repo.Update(ctx, paper.UUID, got); err != nil {
		t.Fatal(err)
	}
	want.PDFStorageKey = ""
	got, err = repo.GetByUUID(ctx, paper.UUID)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("after remove and stale edit = %#v, %v; want %#v", got, err, want)
	}
}

func TestPDFAttachmentMissingPaper(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()
	if err := repo.AttachPDF(ctx, "missing", "pdf/test.pdf"); !errors.Is(err, domain.ErrPaperNotFound) {
		t.Fatalf("AttachPDF error = %v, want ErrPaperNotFound", err)
	}
	if err := repo.DetachPDF(ctx, "missing"); !errors.Is(err, domain.ErrPaperNotFound) {
		t.Fatalf("DetachPDF error = %v, want ErrPaperNotFound", err)
	}
}

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
	if result.Items[0].DOI != "10.1002/stvr.70024" {
		t.Fatalf("Search() first DOI = %q, want %q", result.Items[0].DOI, "10.1002/stvr.70024")
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
