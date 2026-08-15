package bibliography

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"

	paperMemory "github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
)

func getSeedPaper(uuid string) domain.Paper {
	paper, ok := seedPapersByUUID[uuid]
	if !ok {
		panic("missing paper seed data for " + uuid)
	}

	return paper
}

var seedPapersByUUID = indexSeedPapers(paperMemory.SeedData())

func indexSeedPapers(papers []domain.Paper) map[string]domain.Paper {
	papersByUUID := make(map[string]domain.Paper, len(papers))
	for _, paper := range papers {
		papersByUUID[paper.UUID] = paper
	}

	return papersByUUID
}

func TestExportBibLaTeX(t *testing.T) {
	t.Parallel()

	tests := []struct {
		UUID         string
		testDataFile string
	}{
		{UUID: "67132cd6-3213-4b49-ac5e-0d3ffb030a85", testDataFile: "testdata/bauer.bib"},
		{UUID: "14815b6a-6e2d-5b73-8aa3-1a2d7b517106", testDataFile: "testdata/bosu.bib"},
	}

	for _, tt := range tests {
		t.Run(tt.testDataFile, func(t *testing.T) {
			t.Parallel()

			paper := getSeedPaper(tt.UUID)
			papers := []domain.Paper{paper}

			got, err := ExportBibLaTeX(papers)
			if err != nil {
				t.Fatalf("ExportBibLaTeX() error = %v", err)
			}

			want, err := os.ReadFile(tt.testDataFile)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			gotLines := strings.Split(string(got), "\n")
			wantLines := strings.Split(string(want), "\n")
			for line := range max(len(gotLines), len(wantLines)) {
				var gotLine, wantLine string
				if line < len(gotLines) {
					gotLine = gotLines[line]
				}
				if line < len(wantLines) {
					wantLine = wantLines[line]
				}
				if gotLine != wantLine {
					t.Fatalf("ExportBibLaTeX() differs at line %d:\n got: %q\nwant: %q", line+1, gotLine, wantLine)
				}
			}
		})
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
		metadata  domain.Metadata
		container string
	}{
		{name: "journal article", type_: domain.PublicationTypeJournalArticle, entryType: "article", metadata: domain.Metadata{JournalTitle: "Container"}, container: "journaltitle"},
		{name: "conference article", type_: domain.PublicationTypeConferenceArticle, entryType: "inproceedings", metadata: domain.Metadata{BookTitle: "Container"}, container: "booktitle"},
		{name: "book", type_: domain.PublicationTypeBook, entryType: "book"},
		{name: "book chapter", type_: domain.PublicationTypeBookChapter, entryType: "incollection", metadata: domain.Metadata{BookTitle: "Container"}, container: "booktitle"},
		{name: "thesis", type_: domain.PublicationTypeThesis, entryType: "thesis", metadata: domain.Metadata{Institution: "Container"}, container: "institution"},
		{name: "report", type_: domain.PublicationTypeReport, entryType: "report", metadata: domain.Metadata{Institution: "Container"}, container: "institution"},
		{name: "dataset", type_: domain.PublicationTypeDataset, entryType: "dataset"},
		{name: "web page", type_: domain.PublicationTypeWebPage, entryType: "online"},
		{name: "unspecified", entryType: "article", metadata: domain.Metadata{JournalTitle: "Container"}, container: "journaltitle"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ExportBibLaTeX([]domain.Paper{{
				UUID:     "a4b065f1-1b88-4f50-a7fe-1177f3489fcf",
				Type:     tt.type_,
				Metadata: tt.metadata,
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

func TestExportBibLaTeXMapsDistinctContainerFields(t *testing.T) {
	t.Parallel()

	got, err := ExportBibLaTeX([]domain.Paper{{
		UUID: "a4b065f1-1b88-4f50-a7fe-1177f3489fcf",
		Metadata: domain.Metadata{
			JournalTitle:  "Journal Title",
			JournalAbbrev: "J. Title",
			BookTitle:     "Book Title",
			SeriesTitle:   "Series Title",
			EventTitle:    "Conference Title",
			EventLocation: "Gothenburg",
			Institution:   "Example University",
		},
	}})
	if err != nil {
		t.Fatalf("ExportBibLaTeX() error = %v", err)
	}

	for _, want := range []string{
		"  journaltitle = {Journal Title}",
		"  shortjournal = {J. Title}",
		"  booktitle = {Book Title}",
		"  series = {Series Title}",
		"  eventtitle = {Conference Title}",
		"  location = {Gothenburg}",
		"  institution = {Example University}",
	} {
		if !strings.Contains(string(got), want) {
			t.Fatalf("ExportBibLaTeX() missing %q in:\n%s", want, got)
		}
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
