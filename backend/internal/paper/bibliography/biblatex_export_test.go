package bibliography

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

func TestExportBibLaTeX(t *testing.T) {
	t.Parallel()

	papers := []domain.Paper{{
		UUID:            "a4b065f1-1b88-4f50-a7fe-1177f3489fcf",
		DOI:             "10.1000/example",
		Title:           "A {brace} & 50%_# $ ~ ^ \\",
		TitleShort:      "Short title",
		Authors:         []domain.Author{{NameFirst: "Jane", NameMiddle: "Q.", NameLast: "Doe"}, {NameFirst: "Pat", NameLast: "O'Neil"}},
		PublicationDate: domain.Date{Year: 2024, Month: 8, Day: 13},
		Abstract:        "A & B",
		Keywords:        []string{" software testing ", "", "bibliography"},
		Type:            domain.PublicationTypeJournalArticle,
		Metadata: domain.Metadata{
			PublishedIn: "Journal of Examples",
			Publisher:   "Example Press",
			Volume:      "12",
			Issue:       "3",
			Pages:       "42-53",
			ISBN:        []string{"978-1-234-56789-0", "978-9-876-54321-0"},
			ISSN:        []string{"1234-5678", "8765-4321"},
			References:  []string{"mailto:editor@example.com", "https://example.com/paper", "https://second.example.com/paper"},
		},
	}}

	got, err := ExportBibLaTeX(papers)
	if err != nil {
		t.Fatalf("ExportBibLaTeX() error = %v", err)
	}

	want, err := os.ReadFile("testdata/rich-paper.bib")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("ExportBibLaTeX() =\n%s\nwant:\n%s", got, want)
	}
}

func TestExportBibLaTeXUsesUUIDWhenDOIIsUnavailable(t *testing.T) {
	t.Parallel()

	got, err := ExportBibLaTeX([]domain.Paper{{
		UUID:  "a4b065f1-1b88-4f50-a7fe-1177f3489fcf",
		Title: "Untitled DOI",
	}})
	if err != nil {
		t.Fatalf("ExportBibLaTeX() error = %v", err)
	}

	want := "@article{a4b065f1-1b88-4f50-a7fe-1177f3489fcf,\n  title = {Untitled DOI}\n}\n"
	if string(got) != want {
		t.Fatalf("ExportBibLaTeX() = %q, want %q", got, want)
	}
}

func TestExportBibLaTeXMapsTypesAndContainers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		type_     domain.PublicationType
		entryType string
		container string
	}{
		{name: "journal article", type_: domain.PublicationTypeJournalArticle, entryType: "article", container: "journaltitle"},
		{name: "conference article", type_: domain.PublicationTypeConferenceArticle, entryType: "inproceedings", container: "booktitle"},
		{name: "book", type_: domain.PublicationTypeBook, entryType: "book"},
		{name: "book chapter", type_: domain.PublicationTypeBookChapter, entryType: "incollection", container: "booktitle"},
		{name: "thesis", type_: domain.PublicationTypeThesis, entryType: "thesis", container: "institution"},
		{name: "report", type_: domain.PublicationTypeReport, entryType: "report", container: "institution"},
		{name: "dataset", type_: domain.PublicationTypeDataset, entryType: "dataset"},
		{name: "web page", type_: domain.PublicationTypeWebPage, entryType: "online"},
		{name: "unspecified", entryType: "article", container: "journaltitle"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ExportBibLaTeX([]domain.Paper{{
				UUID: "a4b065f1-1b88-4f50-a7fe-1177f3489fcf",
				Type: tt.type_,
				Metadata: domain.Metadata{
					PublishedIn: "Container",
				},
			}})
			if err != nil {
				t.Fatalf("ExportBibLaTeX() error = %v", err)
			}

			if !strings.HasPrefix(string(got), "@"+tt.entryType+"{") {
				t.Fatalf("entry type = %q, want %q", got, tt.entryType)
			}
			if tt.container == "" && strings.Contains(string(got), "Container") {
				t.Fatalf("unexpected container field in %q", got)
			}
			if tt.container != "" && !strings.Contains(string(got), "  "+tt.container+" = {Container}") {
				t.Fatalf("missing %q container field in %q", tt.container, got)
			}
		})
	}
}

func TestExportBibLaTeXRejectsInvalidOrDuplicateKeys(t *testing.T) {
	t.Parallel()

	_, err := ExportBibLaTeX([]domain.Paper{{DOI: "invalid key"}})
	if !errors.Is(err, ErrInvalidBibLaTeXEntryKey) {
		t.Fatalf("invalid key error = %v, want %v", err, ErrInvalidBibLaTeXEntryKey)
	}

	_, err = ExportBibLaTeX([]domain.Paper{{DOI: "10.1000/example"}, {DOI: "10.1000/example"}})
	if !errors.Is(err, ErrDuplicateBibLaTeXEntryKey) {
		t.Fatalf("duplicate key error = %v, want %v", err, ErrDuplicateBibLaTeXEntryKey)
	}
}
