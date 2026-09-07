// Package importer drives the Crossref dump import: walking a directory
// of dump files, parsing every record, and reporting progress.
package importer

import (
	"context"
	"fmt"
	"log/slog"
	"os"
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

// Run walks cfg.Dir, parses every record in every dump file, and shows
// a progress bar of records processed against the total record count.
// A single bad file (e.g. corrupt gzip) does not abort the run; only
// ctx cancellation does.
func (im *Importer) Run(ctx context.Context) error {
	im.log.Info("crawler import starting", slog.String("dir", im.cfg.Dir))
	start := time.Now()

	total, err := reader.CountRecords(ctx, im.cfg.Dir)
	if err != nil {
		return fmt.Errorf("count records: %w", err)
	}
	im.log.Info("counted records", slog.Int("total", total), slog.Duration("elapsed", time.Since(start)))

	bar := newProgressBar(os.Stderr, total)
	bar.Set(0)

	stats := &runStats{im: im, ctx: ctx, bar: bar, start: start}
	walkErr := reader.WalkDumpDir(im.cfg.Dir, stats.visitFile)

	if bar.tty {
		fmt.Fprintln(bar.out)
	}
	im.log.Info("import finished",
		slog.Int("files", stats.files),
		slog.Int("records", stats.records),
		slog.Int("errors", stats.errors),
		slog.Duration("elapsed", time.Since(start)),
	)

	return walkErr
}

// runStats tracks the running totals for a single Importer.Run call and
// drives the per-file processing that reader.WalkDumpDir invokes. It's
// its own type, separate from Importer, because these counters are
// scoped to one run rather than to the Importer's lifetime.
type runStats struct {
	im    *Importer
	ctx   context.Context
	bar   *progressBar
	start time.Time

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

	s.im.log.Debug("file complete",
		slog.String("file", path),
		slog.Int("records", fileRecords),
		slog.Int("errors", fileErrors),
		slog.Int("total_records", s.records),
		slog.Duration("elapsed", time.Since(s.start)),
	)
	return nil
}

// importFile drains one dump file's records, tallying successes and
// per-line parse errors into s.records/s.errors and redrawing the
// progress bar after every line. A non-nil error means either the file
// couldn't be opened/decompressed, or the run was interrupted.
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
		} else if _, err := s.im.paperService.Create(s.ctx, toPaper(rec.Paper)); err != nil {
			errs++
			s.errors++
			s.im.log.Warn("create paper",
				slog.String("file", path),
				slog.Int("line", rec.LineNumber),
				slog.String("doi", rec.Paper.DOI),
				slog.String("error", err.Error()),
			)
		} else {
			records++
			s.records++
		}

		s.bar.Set(s.records + s.errors)
		return s.ctx.Err()
	})
	if err != nil {
		return records, errs, fmt.Errorf("read file: %w", err)
	}
	return records, errs, nil
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
