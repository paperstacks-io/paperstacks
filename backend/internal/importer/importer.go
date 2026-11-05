// Package importer drives the Crossref dump import: walking a directory
// of dump files, parsing every record, and reporting progress.
package importer

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	crossref "github.com/paperstacks.io/paperstacks/internal/importer/crossref/domain"
	"github.com/paperstacks.io/paperstacks/internal/importer/crossref/reader"
	paperapp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	paper "github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

// Config holds the parameters needed to run a Crossref dump import.
type Config struct {
	// Dir is the directory containing *.jsonl[.gz] dump files.
	Dir string
	// LogDir is the directory the per-run import log file is created in.
	LogDir string
}

// Importer walks a directory of Crossref dump files, parses every
// record, and reports import statistics.
type Importer struct {
	cfg          Config
	log          *slog.Logger
	paperService *paperapp.PaperService
}

// New constructs an Importer for the given configuration.
func New(cfg Config, log *slog.Logger, papers *paperapp.PaperService) *Importer {
	return &Importer{cfg: cfg, log: log, paperService: papers}
}

// Run walks cfg.Dir, parses every record in every dump file, and logs each
// completed file. A single bad file (e.g. corrupt gzip) does not abort the
// run; only ctx cancellation does.
func (im *Importer) Run(ctx context.Context) error {
	im.log.Info("crawler import starting", slog.String("dir", im.cfg.Dir))
	start := time.Now()

	logFile, err := createLogFile(im.cfg.LogDir, start, "importer")
	if err != nil {
		return err
	}
	defer logFile.Close()

	im.log.Info("import log file created", slog.String("file", logFile.Name()))

	stats := &runStats{im: im, ctx: ctx, start: start, logFile: logFile}
	walkErr := reader.WalkDumpDir(im.cfg.Dir, stats.visitFile)

	im.log.Info("import finished",
		slog.Int("files", stats.files),
		slog.Int("records", stats.records),
		slog.Int("errors", stats.errors),
		slog.Duration("elapsed", time.Since(start)),
	)

	return walkErr
}

// createLogFile creates dir if needed and an empty log file in it named
// after start and kind, e.g. 20250315-143005-importer-log.txt.
func createLogFile(dir string, start time.Time, kind string) (*os.File, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}
	name := start.Format("20060102-150405") + "-" + kind + "-log.txt"
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return nil, fmt.Errorf("create log file: %w", err)
	}
	return f, nil
}

// runStats tracks the running totals for a single Importer.Run call and
// drives the per-file processing that reader.WalkDumpDir invokes. It's
// its own type, separate from Importer, because these counters are
// scoped to one run rather than to the Importer's lifetime.
type runStats struct {
	im      *Importer
	ctx     context.Context
	start   time.Time
	logFile io.Writer

	files   int
	records int
	errors  int
}

// visitFile is called by reader.WalkDumpDir once per dump file, in
// WalkDumpDir's file-name order. A single bad file (e.g. corrupt gzip)
// doesn't abort the walk — only ctx cancellation does.
func (s *runStats) visitFile(path string) error {
	if err := s.ctx.Err(); err != nil {
		return err
	}

	fileRecords, fileErrors, err := s.importFile(path)
	s.files++

	if err != nil {
		s.im.log.Error("file failed", slog.String("file", path), slog.String("error", err.Error()))
		// A single bad file shouldn't abort the whole run — only stop
		// the walk if we were interrupted.
		return s.ctx.Err()
	}

	s.im.log.Info("file complete",
		slog.String("file", path),
		slog.Int("records", fileRecords),
		slog.Int("errors", fileErrors),
		slog.Int("total_records", s.records),
		slog.Duration("elapsed", time.Since(s.start)),
	)
	return nil
}

// importFile drains one dump file's records, tallying successes and
// per-line errors into s.records/s.errors. A non-nil error means either the
// file couldn't be opened/decompressed, or the run was interrupted.
func (s *runStats) importFile(path string) (records, errs int, err error) {
	err = reader.WalkFile(path, func(rec reader.Record) error {
		if rec.Err != nil {
			errs++
			s.errors++
			s.im.log.Warn("parse error",
				slog.String("file", path),
				slog.Int("line", rec.LineNumber),
				slog.String("error", rec.Err.Error()),
			)
			s.writeLogEntry(path, rec.LineNumber, "", rec.Err.Error())
			return s.ctx.Err()
		}

		p := toPaper(rec.Paper)
		if problems := invalidFields(p); problems != "" {
			errs++
			s.errors++
			s.im.log.Warn("invalid paper",
				slog.String("file", path),
				slog.Int("line", rec.LineNumber),
				slog.String("doi", rec.Paper.DOI),
				slog.String("error", problems),
			)
			s.writeLogEntry(path, rec.LineNumber, rec.Paper.DOI, problems)
		} else if _, err := s.im.paperService.Create(s.ctx, p); err != nil {
			errs++
			s.errors++
			s.im.log.Warn("create paper",
				slog.String("file", path),
				slog.Int("line", rec.LineNumber),
				slog.String("doi", rec.Paper.DOI),
				slog.String("error", err.Error()),
			)
			s.writeLogEntry(path, rec.LineNumber, rec.Paper.DOI, err.Error())
		} else {
			records++
			s.records++
		}

		return s.ctx.Err()
	})
	if err != nil {
		return records, errs, fmt.Errorf("read file: %w", err)
	}
	return records, errs, nil
}

// writeLogEntry appends one line to the run's log file describing a
// record that could not be imported, e.g.:
//
//	time=2025-03-15T14:30:05.123+01:00 file=dump/0.jsonl.gz line=42 doi=10.1234/abc error=missing title
func (s *runStats) writeLogEntry(path string, line int, doi, msg string) {
	now := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
	if _, err := fmt.Fprintf(s.logFile, "time=%s file=%s line=%d doi=%s error=%s\n",
		now, path, line, doi, msg); err != nil {
		s.im.log.Error("write import log", slog.String("error", err.Error()))
	}
}

// invalidFields explains why p can't be imported, or returns "" if it can.
func invalidFields(p paper.Paper) string {
	var problems []string
	if strings.TrimSpace(p.DOI) == "" {
		problems = append(problems, "missing DOI")
	}
	if strings.TrimSpace(p.Title) == "" {
		problems = append(problems, "missing title")
	}
	if !p.PublicationDate.IsValid() {
		d := p.PublicationDate
		problems = append(problems, fmt.Sprintf("invalid publication date (year=%d month=%d day=%d)", d.Year, d.Month, d.Day))
	}
	if !p.Type.IsValid() {
		problems = append(problems, fmt.Sprintf("invalid type %q", p.Type))
	}
	if ts := p.Metadata.DataSourceTimestamp; ts != "" {
		if _, err := time.Parse(time.RFC3339, ts); err != nil {
			problems = append(problems, fmt.Sprintf("invalid data source timestamp %q", ts))
		}
	}
	return strings.Join(problems, "; ")
}

func toPaper(source *crossref.Paper) paper.Paper {
	target := paper.Paper{
		DOI:             source.DOI,
		Title:           source.Title,
		Abstract:        source.Abstract,
		Type:            paper.PublicationType(source.Type),
		PublicationDate: paper.Date{Year: source.Issued.Year, Month: source.Issued.Month, Day: source.Issued.Day},
		Metadata: paper.Metadata{
			Publisher:     source.Publisher,
			JournalTitle:  source.ContainerTitle,
			JournalAbbrev: source.ShortContainerTitle,
			Pages:         source.Page,
			Volume:        source.Volume,
			Issue:         source.Issue,
			DataSource:    source.Source,
		},
	}
	if target.Type == "proceedings-article" {
		target.Type = paper.PublicationTypeConferenceArticle
	}
	if !target.Type.IsValid() {
		target.Type = ""
	}
	if !source.Indexed.IsZero() {
		target.Metadata.DataSourceTimestamp = source.Indexed.Format(time.RFC3339)
	}

	for _, author := range source.Authors {
		target.Authors = append(target.Authors, paper.Author{
			NameFirst:   author.GivenName,
			NameLast:    author.FamilyName,
			Affiliation: strings.Join(author.Affiliations, "; "),
			ORCID:       author.ORCID,
		})
	}
	for _, issn := range source.ISSNs {
		target.Metadata.ISSN = append(target.Metadata.ISSN, issn.Value)
	}
	for _, isbn := range source.ISBNs {
		target.Metadata.ISBN = append(target.Metadata.ISBN, isbn.Value)
	}
	return target
}
