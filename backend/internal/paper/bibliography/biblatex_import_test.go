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
			source, err := os.ReadFile(filepath.Join("testdata", tt.file))
			if err != nil {
				t.Fatal(err)
			}

			result, err := ImportBibLaTeX(source)
			if err != nil {
				t.Fatal(err)
			}

			if len(result.Imported) != 1 {
				t.Fatalf("expected one imported candidate, got %d", len(result.Imported))
			}
			if len(result.Failed) != 0 {
				t.Fatalf("expected no failed candidates, got %#v", result.Failed)
			}

			imported := result.Imported[0]
			if len(imported.Errors) != 0 {
				t.Fatalf("expected no errors, got %#v", imported.Errors)
			}
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

	source, err := os.ReadFile(filepath.Join("testdata", "import-partial.bib"))
	if err != nil {
		t.Fatal(err)
	}

	result, err := ImportBibLaTeX(source)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Imported) != 0 {
		t.Fatalf("expected 0 imported candidate, got %d", len(result.Imported))
	}
	if len(result.Failed) != 1 {
		t.Fatalf("expected 1 failed imported candidate, got %d", len(result.Failed))
	}

	imported := result.Failed[0]
	if !imported.Paper.PublicationDate.IsZero() {
		t.Errorf("publication date = %#v, want zero date", imported.Paper.PublicationDate)
	}
	if len(imported.Paper.Metadata.References) != 0 {
		t.Errorf("references = %#v, want omitted invalid URL", imported.Paper.Metadata.References)
	}
	if !hasDiagnostic(imported.Errors, "date", "partial") {
		t.Errorf("missing date error: %#v", imported.Errors)
	}
	for _, field := range []string{"url", "editor"} {
		if !hasDiagnostic(imported.Warnings, field, "partial") {
			t.Errorf("missing %s warning: %#v", field, imported.Warnings)
		}
	}
}

func TestImportBibLaTeXReportsEntryErrors(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile(filepath.Join("testdata", "import-error.bib"))
	if err != nil {
		t.Fatal(err)
	}

	result, err := ImportBibLaTeX(source)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Imported) != 1 {
		t.Fatalf("expected 1 imported candidates, got %d", len(result.Imported))
	}
	if result.Imported[0].SourceKey != "ok" {
		t.Fatalf("expected %s as key for imported candidates, got %s", "ok", result.Imported[0].SourceKey)
	}

	if len(result.Failed) != 3 {
		t.Fatalf("expected 3 failed candidates, got %#v", result.Failed)
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
	if len(result.Imported) != 0 || len(result.Failed) != 0 {
		t.Errorf("result = %#v, want empty result", result)
	}
}

func TestParseBibLaTeXDate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		want    domain.Date
		wantErr bool
	}{
		{name: "year", value: "2024", want: domain.Date{Year: 2024}},
		{name: "year and month", value: "2024-02", want: domain.Date{Year: 2024, Month: 2}},
		{name: "full leap date", value: "2024-02-29", want: domain.Date{Year: 2024, Month: 2, Day: 29}},
		{name: "date range", value: "2024/2025", wantErr: true},
		{name: "zero year", value: "0000", wantErr: true},
		{name: "invalid month", value: "2024-13", wantErr: true},
		{name: "invalid calendar day", value: "2023-02-29", wantErr: true},
		{name: "too precise", value: "2024-01-02-03", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBibLaTeXDate(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseBibLaTeXDate(%q) succeeded, want error", tt.value)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseBibLaTeXDate(%q) error = %v", tt.value, err)
			}
			if got != tt.want {
				t.Errorf("parseBibLaTeXDate(%q) = %#v, want %#v", tt.value, got, tt.want)
			}
		})
	}
}

func hasDiagnostic(diagnostics []Diagnostic, field, entryKey string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Field == field && diagnostic.EntryKey == entryKey {
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
