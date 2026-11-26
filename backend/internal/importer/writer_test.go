package importer

import (
	"context"
	"slices"
	"testing"

	paperapp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	paper "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	papermemory "github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
)

func crossrefPaper(title string) paper.Paper {
	return paper.Paper{
		DOI:             "10.1/upsert",
		Title:           title,
		Type:            paper.PublicationTypeJournalArticle,
		PublicationDate: paper.Date{Year: 2026, Month: 3},
		Authors:         []paper.Author{{NameFirst: "Ada", NameLast: "Lovelace"}},
		Metadata:        paper.Metadata{Publisher: "Crossref Publisher", Volume: "7"},
	}
}

func TestServiceWriterCreatesNewPaper(t *testing.T) {
	ctx := context.Background()
	papers := paperapp.NewPaperService(papermemory.NewRepository(), nil)

	if err := NewServiceWriter(papers).Upsert(ctx, crossrefPaper("First")); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, err := papers.GetByDOI(ctx, "10.1/upsert")
	if err != nil {
		t.Fatalf("GetByDOI: %v", err)
	}
	if got.UUID == "" || got.Title != "First" || got.Metadata.Publisher != "Crossref Publisher" {
		t.Errorf("created paper = %+v", got)
	}
}

func TestServiceWriterUpdatesCrossrefFieldsOnly(t *testing.T) {
	ctx := context.Background()
	papers := paperapp.NewPaperService(papermemory.NewRepository(), nil)

	existing := crossrefPaper("Old title")
	existing.TitleShort = "Old"
	existing.Keywords = []string{"kept"}
	existing.PublicationStatus = "published"
	existing.Metadata.License = "CC-BY"
	created, err := papers.Create(ctx, existing)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	changed := crossrefPaper("New title")
	changed.Authors = []paper.Author{{NameFirst: "Grace", NameLast: "Hopper"}}
	changed.Metadata.Volume = "8"
	if err := NewServiceWriter(papers).Upsert(ctx, changed); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, err := papers.GetByDOI(ctx, "10.1/upsert")
	if err != nil {
		t.Fatalf("GetByDOI: %v", err)
	}
	if got.UUID != created.UUID {
		t.Errorf("UUID = %q, want unchanged %q", got.UUID, created.UUID)
	}
	if got.Title != "New title" || got.Metadata.Volume != "8" ||
		len(got.Authors) != 1 || got.Authors[0].NameLast != "Hopper" {
		t.Errorf("Crossref fields not updated: %+v", got)
	}
	if got.TitleShort != "Old" || !slices.Equal(got.Keywords, []string{"kept"}) ||
		got.PublicationStatus != "published" || got.Metadata.License != "CC-BY" {
		t.Errorf("non-Crossref fields not kept: %+v", got)
	}
}
