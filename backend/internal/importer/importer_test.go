package importer

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	paperapp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	paper "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	papermemory "github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
)

func TestRunCreatesEmptyLogFile(t *testing.T) {
	logDir := filepath.Join(t.TempDir(), "logs")
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	papers := paperapp.NewPaperService(papermemory.NewRepository(), nil)
	imp := New(Config{Dir: t.TempDir(), LogDir: logDir}, log, papers)

	if err := imp.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	entries, err := os.ReadDir(logDir)
	if err != nil {
		t.Fatalf("read log dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d log files, want 1", len(entries))
	}

	name := entries[0].Name()
	if !regexp.MustCompile(`^\d{8}-\d{6}-importer-log\.txt$`).MatchString(name) {
		t.Errorf("log file name %q does not match YYYYMMDD-HHMMSS-importer-log.txt", name)
	}

	info, err := entries[0].Info()
	if err != nil {
		t.Fatalf("stat log file: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("log file size = %d, want 0", info.Size())
	}
}

func TestRunLogsOneLinePerFailedRecord(t *testing.T) {
	dumpDir := t.TempDir()
	dump := `{"DOI":"10.1/valid","title":["Valid"],"type":"journal-article"}
{"title":["No DOI"]}
{"DOI":"10.1/no-title"}
{"DOI":"10.1/bad-date","title":["Bad Date"],"issued":{"date-parts":[[2020,13]]}}
`
	dumpPath := filepath.Join(dumpDir, "0.jsonl")
	if err := os.WriteFile(dumpPath, []byte(dump), 0o644); err != nil {
		t.Fatalf("write dump: %v", err)
	}

	logDir := t.TempDir()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	papers := paperapp.NewPaperService(papermemory.NewRepository(), nil)
	imp := New(Config{Dir: dumpDir, LogDir: logDir}, log, papers)

	if err := imp.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	entries, err := os.ReadDir(logDir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("read log dir: %d entries, err %v", len(entries), err)
	}
	content, err := os.ReadFile(filepath.Join(logDir, entries[0].Name()))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSuffix(string(content), "\n"), "\n")
	want := []string{
		"file=" + dumpPath + " line=2 doi= error=line 2: record has no DOI",
		"file=" + dumpPath + " line=3 doi=10.1/no-title error=missing title",
		"file=" + dumpPath + " line=4 doi=10.1/bad-date error=invalid publication date (year=2020 month=13 day=0)",
	}
	if len(lines) != len(want) {
		t.Fatalf("got %d log lines, want %d:\n%s", len(lines), len(want), content)
	}
	timestamp := regexp.MustCompile(`^time=\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}(Z|[+-]\d{2}:\d{2}) `)
	for i, line := range lines {
		if !timestamp.MatchString(line) {
			t.Errorf("line %d = %q, want time= prefix", i+1, line)
		}
		if got := timestamp.ReplaceAllString(line, ""); got != want[i] {
			t.Errorf("line %d = %q, want %q", i+1, got, want[i])
		}
	}
}

func TestInvalidFields(t *testing.T) {
	valid := paper.Paper{
		DOI:             "10.1/valid",
		Title:           "Valid",
		Type:            paper.PublicationTypeJournalArticle,
		PublicationDate: paper.Date{Year: 2020, Month: 2, Day: 29},
		Metadata:        paper.Metadata{DataSourceTimestamp: "2025-03-01T12:00:00Z"},
	}

	tests := []struct {
		name   string
		modify func(*paper.Paper)
		want   string
	}{
		{"valid", func(*paper.Paper) {}, ""},
		{"missing DOI", func(p *paper.Paper) { p.DOI = " " }, "missing DOI"},
		{"missing title", func(p *paper.Paper) { p.Title = "" }, "missing title"},
		{"invalid publication date", func(p *paper.Paper) { p.PublicationDate.Day = 30 }, "invalid publication date (year=2020 month=2 day=30)"},
		{"invalid type", func(p *paper.Paper) { p.Type = "posted-content" }, `invalid type "posted-content"`},
		{"invalid data source timestamp", func(p *paper.Paper) { p.Metadata.DataSourceTimestamp = "2025-03-01" }, `invalid data source timestamp "2025-03-01"`},
		{"multiple problems", func(p *paper.Paper) { p.DOI = ""; p.Title = "" }, "missing DOI; missing title"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := valid
			tt.modify(&p)
			if got := invalidFields(p); got != tt.want {
				t.Errorf("invalidFields() = %q, want %q", got, tt.want)
			}
		})
	}
}
