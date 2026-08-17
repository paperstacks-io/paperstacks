package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/paperstacks.io/paperstacks/internal/paper/bibliography"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
)

func TestBibliographyServiceExportBibLaTeXPreservesRequestedOrder(t *testing.T) {
	t.Parallel()

	service := NewBibliographyService(NewPaperService(memory.NewRepository()))
	firstUUID := "0f324174-926b-585d-b121-3a1e3f7fee0b"
	secondUUID := "36583bb4-8cdc-554e-bcf5-f67b60d0b290"

	got, err := service.ExportBibLaTeX(context.Background(), []string{" " + firstUUID + " ", secondUUID})
	if err != nil {
		t.Fatalf("ExportBibLaTeX() error = %v", err)
	}

	first := strings.Index(string(got), "@article{10.1109/52.28121,")
	second := strings.Index(string(got), "@article{10.48550/ARXIV.1709.08439,")
	if first == -1 || second == -1 || first >= second {
		t.Fatalf("ExportBibLaTeX() did not preserve request order:\n%s", got)
	}
}

func TestBibliographyServiceExportBibLaTeXReturnsMissingPaper(t *testing.T) {
	t.Parallel()

	service := NewBibliographyService(NewPaperService(memory.NewRepository()))

	_, err := service.ExportBibLaTeX(context.Background(), []string{"a4b065f1-1b88-4f50-a7fe-1177f3489fcf"})
	if !errors.Is(err, domain.ErrPaperNotFound) {
		t.Fatalf("ExportBibLaTeX() error = %v, want %v", err, domain.ErrPaperNotFound)
	}
}

func TestBibliographyServiceImportBibLaTeXPersistsValidPaper(t *testing.T) {
	t.Parallel()

	papers := NewPaperService(memory.NewRepository())
	service := NewBibliographyService(papers)

	result, err := service.ImportBibLaTeX(t.Context(), []byte(`@article{created,
  title = {Created Paper},
  date = {2024},
  doi = {10.9999/application-created},
  note = {not represented}
}`))
	if err != nil {
		t.Fatalf("ImportBibLaTeX() error = %v", err)
	}
	if len(result.Created) != 1 {
		t.Fatalf("created entries = %d, want 1", len(result.Created))
	}

	created := result.Created[0]
	if _, err := uuid.Parse(created.Paper.UUID); err != nil {
		t.Fatalf("created UUID = %q, want valid UUID: %v", created.Paper.UUID, err)
	}
	if len(created.Warnings) == 0 {
		t.Fatal("created warnings are empty, want unsupported-field warning")
	}
	if _, err := papers.GetByUUID(t.Context(), created.Paper.UUID); err != nil {
		t.Fatalf("created Paper was not persisted: %v", err)
	}
}

func TestBibliographyServiceImportBibLaTeXPersistsValidSiblings(t *testing.T) {
	t.Parallel()

	service := NewBibliographyService(NewPaperService(memory.NewRepository()))
	result, err := service.ImportBibLaTeX(t.Context(), []byte(`@article{valid,
  title = {Valid Paper},
  date = {2024},
  doi = {10.9999/application-valid}
}

@article{missing-doi,
  title = {Missing DOI},
  date = {2024}
}`))
	if err != nil {
		t.Fatalf("ImportBibLaTeX() error = %v", err)
	}
	if len(result.Created) != 1 || result.Created[0].SourceKey != "valid" {
		t.Fatalf("created entries = %#v, want only valid", result.Created)
	}
	if len(result.Rejected) != 1 || result.Rejected[0].SourceKey != "missing-doi" {
		t.Fatalf("rejected entries = %#v, want only missing-doi", result.Rejected)
	}
	if len(result.Rejected[0].Errors) == 0 || result.Rejected[0].Errors[0].Field != "doi" {
		t.Fatalf("rejection errors = %#v, want DOI diagnostic", result.Rejected[0].Errors)
	}
}

func TestBibliographyServiceImportBibLaTeXReusesExistingDOI(t *testing.T) {
	t.Parallel()

	papers := NewPaperService(memory.NewRepository())
	existing, err := papers.Create(t.Context(), domain.Paper{
		DOI:   "10.9999/application-existing",
		Title: "Existing Paper",
		Type:  domain.PublicationTypeJournalArticle,
	})
	if err != nil {
		t.Fatalf("seed existing Paper: %v", err)
	}
	service := NewBibliographyService(papers)

	result, err := service.ImportBibLaTeX(t.Context(), []byte(`@article{existing,
  title = {Imported Duplicate},
  date = {2024},
  doi = {10.9999/application-existing}
}`))
	if err != nil {
		t.Fatalf("ImportBibLaTeX() error = %v", err)
	}
	if len(result.Created) != 0 {
		t.Fatalf("created entries = %d, want 0", len(result.Created))
	}
	if len(result.Existing) != 1 || result.Existing[0].Paper.UUID != existing.UUID {
		t.Fatalf("existing entries = %#v, want stored Paper %q", result.Existing, existing.UUID)
	}
}

func TestBibliographyServiceImportBibLaTeXRejectsMalformedDocumentWithoutWrites(t *testing.T) {
	t.Parallel()

	repo := memory.NewRepository()
	before, err := repo.List(t.Context())
	if err != nil {
		t.Fatalf("list Papers before import: %v", err)
	}
	service := NewBibliographyService(NewPaperService(repo))

	result, err := service.ImportBibLaTeX(t.Context(), []byte("@article{broken,"))
	if !errors.Is(err, bibliography.ErrInvalidBibLaTeX) {
		t.Fatalf("ImportBibLaTeX() error = %v, want %v", err, bibliography.ErrInvalidBibLaTeX)
	}
	if len(result.Created) != 0 || len(result.Existing) != 0 || len(result.Rejected) != 0 {
		t.Fatalf("ImportBibLaTeX() result = %#v, want empty result", result)
	}

	after, err := repo.List(t.Context())
	if err != nil {
		t.Fatalf("list Papers after import: %v", err)
	}
	if len(after) != len(before) {
		t.Fatalf("Paper count after malformed import = %d, want %d", len(after), len(before))
	}
}

var errSecondSave = errors.New("second save failed")

type failSecondSaveRepository struct {
	domain.Repository
	saves int
}

func (r *failSecondSaveRepository) Save(ctx context.Context, paper domain.Paper) (domain.Paper, error) {
	r.saves++
	if r.saves == 2 {
		return domain.Paper{}, errSecondSave
	}
	return r.Repository.Save(ctx, paper)
}

func TestBibliographyServiceImportBibLaTeXReturnsPartialResultOnRepositoryFailure(t *testing.T) {
	repo := &failSecondSaveRepository{Repository: memory.NewRepository()}
	service := NewBibliographyService(NewPaperService(repo))

	result, err := service.ImportBibLaTeX(t.Context(), []byte(`@article{first,
  title = {First Paper},
  date = {2024},
  doi = {10.9999/application-first}
}

@article{second,
  title = {Second Paper},
  date = {2024},
  doi = {10.9999/application-second}
}`))
	if !errors.Is(err, errSecondSave) {
		t.Fatalf("ImportBibLaTeX() error = %v, want %v", err, errSecondSave)
	}
	if len(result.Created) != 1 || result.Created[0].SourceKey != "first" {
		t.Fatalf("created entries = %#v, want persisted first entry", result.Created)
	}
}
