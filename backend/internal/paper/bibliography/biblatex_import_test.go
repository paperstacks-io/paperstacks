package bibliography

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

func TestImportBibLaTeXMapsSeedPapers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		file string
		uuid string
	}{
		{
			name: "journal article",
			file: "bauer.bib",
			uuid: "67132cd6-3213-4b49-ac5e-0d3ffb030a85",
		},
		{
			name: "conference article",
			file: "bosu.bib",
			uuid: "14815b6a-6e2d-5b73-8aa3-1a2d7b517106",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := importBibLaTeXFixture(t, tt.file)
			if len(result.Entries) != 1 {
				t.Fatalf("expected one candidate, got %d", len(result.Entries))
			}
			if len(result.Diagnostics) != 0 {
				t.Fatalf("expected no diagnostics, got %#v", result.Diagnostics)
			}

			imported := result.Entries[0]
			want := getSeedPaper(tt.uuid)
			if imported.SourceKey != want.DOI {
				t.Errorf("source key = %q, want %q", imported.SourceKey, want.DOI)
			}
			if imported.Paper.UUID != "" {
				t.Errorf("UUID = %q, want no UUID allocation", imported.Paper.UUID)
			}
			if !papersEqual(imported.Paper, want) {
				t.Errorf("paper = %#v, want seed paper %#v", imported.Paper, want)
			}
		})
	}
}

func TestImportBibLaTeXReportsRepresentationalProblems(t *testing.T) {
	t.Parallel()

	result := importBibLaTeXFixture(t, "import-partial.bib")
	if len(result.Entries) != 1 {
		t.Fatalf("expected one candidate, got %d", len(result.Entries))
	}

	imported := result.Entries[0]
	if !imported.Paper.PublicationDate.IsZero() {
		t.Errorf("publication date = %#v, want zero date", imported.Paper.PublicationDate)
	}
	if len(imported.Paper.Metadata.References) != 0 {
		t.Errorf("references = %#v, want omitted invalid URL", imported.Paper.Metadata.References)
	}
	for _, code := range []string{"unrepresentable-date", "unrepresentable-url", "unsupported-field"} {
		if !hasDiagnostic(imported.Warnings, code, "partial") {
			t.Errorf("missing %s diagnostic: %#v", code, imported.Warnings)
		}
	}
}

func TestImportBibLaTeXReturnsDOILessCandidate(t *testing.T) {
	t.Parallel()

	result := importBibLaTeXFixture(t, "import-doi-less.bib")
	if len(result.Entries) != 1 {
		t.Fatalf("expected one candidate, got %d", len(result.Entries))
	}

	imported := result.Entries[0]
	if imported.Paper.DOI != "" {
		t.Errorf("DOI = %q, want empty", imported.Paper.DOI)
	}
	if got, want := imported.Paper.PublicationDate, (domain.Date{Year: 2025, Month: 3}); got != want {
		t.Errorf("publication date = %#v, want %#v", got, want)
	}
	if !hasDiagnostic(imported.Warnings, "missing-doi", "doi-less") {
		t.Errorf("missing DOI diagnostic: %#v", imported.Warnings)
	}
}

func TestImportBibLaTeXRejectsMalformedDocument(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile(filepath.Join("testdata", "import-malformed.bib"))
	if err != nil {
		t.Fatal(err)
	}

	result, err := ImportBibLaTeX(source)
	if err == nil {
		t.Fatal("expected malformed source error")
	}
	if !errors.Is(err, ErrInvalidBibLaTeX) {
		t.Errorf("error = %v, want ErrInvalidBibLaTeX", err)
	}
	if len(result.Entries) != 0 || len(result.Diagnostics) != 0 {
		t.Errorf("result = %#v, want empty result", result)
	}
}

func TestImportBibLaTeXPreservesEntryOrderAndReportsUnsupportedType(t *testing.T) {
	t.Parallel()

	result, err := ImportBibLaTeX([]byte(`
@article{first,
  title = "First",
  doi = "10.1000/first",
}
@patent{second,
  title = {Second},
  doi = {10.1000/second},
}
`))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(result.Entries), 2; got != want {
		t.Fatalf("entries = %d, want %d", got, want)
	}
	if got, want := result.Entries[0].SourceKey, "first"; got != want {
		t.Errorf("first source key = %q, want %q", got, want)
	}
	if got, want := result.Entries[1].SourceKey, "second"; got != want {
		t.Errorf("second source key = %q, want %q", got, want)
	}
	if !hasDiagnostic(result.Entries[1].Warnings, "unsupported-entry-type", "second") {
		t.Errorf("missing unsupported type diagnostic: %#v", result.Entries[1].Warnings)
	}
	if !hasDiagnostic(result.Entries[1].Warnings, "invalid-type", "second") {
		t.Errorf("missing invalid type diagnostic: %#v", result.Entries[1].Warnings)
	}
}

func importBibLaTeXFixture(t *testing.T, name string) ImportResult {
	t.Helper()

	source, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ImportBibLaTeX(source)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func hasDiagnostic(diagnostics []Diagnostic, code, entryKey string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code && diagnostic.EntryKey == entryKey {
			return true
		}
	}
	return false
}

func papersEqual(got, want domain.Paper) bool {
	return got.DOI == want.DOI &&
		got.Title == want.Title &&
		got.TitleShort == want.TitleShort &&
		got.PublicationDate == want.PublicationDate &&
		got.Abstract == want.Abstract &&
		got.Type == want.Type &&
		authorsEqual(got.Authors, want.Authors) &&
		stringsEqual(got.Keywords, want.Keywords) &&
		metadataEqual(got.Metadata, want.Metadata)
}

func authorsEqual(got, want []domain.Author) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}

func metadataEqual(got, want domain.Metadata) bool {
	return got.Publisher == want.Publisher &&
		got.JournalTitle == want.JournalTitle &&
		got.JournalAbbrev == want.JournalAbbrev &&
		got.BookTitle == want.BookTitle &&
		got.SeriesTitle == want.SeriesTitle &&
		got.EventTitle == want.EventTitle &&
		got.EventLocation == want.EventLocation &&
		got.Institution == want.Institution &&
		got.Pages == want.Pages &&
		got.Volume == want.Volume &&
		got.Issue == want.Issue &&
		stringsEqual(got.ISBN, want.ISBN) &&
		stringsEqual(got.ISSN, want.ISSN) &&
		stringsEqual(got.References, want.References)
}

func stringsEqual(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}
