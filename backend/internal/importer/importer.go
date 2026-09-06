// Package importer drives the Crossref dump import: walking a directory
// of dump files, parsing every record, and reporting progress.
package importer

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/importer/crossref/reader"
)

// Config holds the parameters needed to run a Crossref dump import.
type Config struct {
	// Dir is the directory containing *.jsonl[.gz] dump files.
	Dir string
}

// Importer walks a directory of Crossref dump files, parses every
// record, and reports import statistics.
type Importer struct {
	cfg Config
	log *slog.Logger
}

// New constructs an Importer for the given configuration.
func New(cfg Config, log *slog.Logger) *Importer {
	return &Importer{cfg: cfg, log: log}
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

	bar.Done()
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
	ch, err := reader.ReadFile(path)
	if err != nil {
		return 0, 0, fmt.Errorf("open file: %w", err)
	}

	for rec := range ch {
		if rec.Err != nil {
			errs++
			s.errors++
			s.im.log.Warn("parse error",
				slog.String("file", path),
				slog.Int("line", rec.LineNumber),
				slog.String("error", rec.Err.Error()),
			)
		} else {
			records++
			s.records++
		}

		s.bar.Set(s.records + s.errors)

		if err := s.ctx.Err(); err != nil {
			return records, errs, err
		}
	}

	return records, errs, nil
}
