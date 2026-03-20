package memory

import (
	"context"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

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

	papers, err := repo.Search(context.Background(), "exploratory testing", "", nil, nil)
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if len(papers) != 1 {
		t.Fatalf("Search() returned %d papers, want 1", len(papers))
	}
	if papers[0].DOI != "10.1109/isese.2005.1541817" {
		t.Fatalf("Search() DOI = %q, want %q", papers[0].DOI, "10.1109/isese.2005.1541817")
	}
}

func TestRepositorySearchByKeyword(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	papers, err := repo.Search(context.Background(), "", "gui testing", nil, nil)
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if len(papers) != 3 {
		t.Fatalf("Search() returned %d papers, want 3", len(papers))
	}
}

func TestRepositorySearchByTitleAndKeywordSamePaper(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	papers, err := repo.Search(context.Background(), "augmented testing", "bayesian data analysis", nil, nil)
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if len(papers) != 1 {
		t.Fatalf("Search() returned %d papers, want 1", len(papers))
	}
	if papers[0].DOI != "10.1007/s10664-024-10522-z" {
		t.Fatalf("Search() DOI = %q, want %q", papers[0].DOI, "10.1007/s10664-024-10522-z")
	}
}

func TestRepositorySearchByTitleAndKeywordDifferentPapers(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	papers, err := repo.Search(context.Background(), "a multiple case study", "bayesian data analysis", nil, nil)
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
	if len(papers) != 2 {
		t.Fatalf("Search() returned %d papers, want 2", len(papers))
	}
}

func TestRepositorySearchNoResultReturnsNotFound(t *testing.T) {
	t.Parallel()

	repo := NewRepository()

	result, err := repo.Search(context.Background(), "nonexistent title", "", nil, nil)
	if len(result) != 0 {
		t.Fatalf("Search() result = %v, want empty slice", len(result))
	}
	if err != nil {
		t.Fatalf("Search() error = %v, want nil", err)
	}
}
