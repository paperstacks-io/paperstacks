package importer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/importer/crossref/api"
	"github.com/paperstacks.io/paperstacks/internal/importer/crossref/parser"
	"github.com/paperstacks.io/paperstacks/internal/importer/syncstate"
)

// PageFetcher fetches one cursor page of works matching filter on day.
// *api.Client implements it.
type PageFetcher interface {
	FetchPage(ctx context.Context, filter api.Filter, day time.Time, cursor string) (api.Page, error)
}

// UpdaterConfig holds the parameters of a Crossref API sync run.
type UpdaterConfig struct {
	// Filter selects the Crossref date the day windows match against.
	Filter api.Filter
	// From is the first day to crawl when the filter's state is empty.
	From time.Time
	// Until is the last day (inclusive) to crawl.
	Until time.Time
	// StateDir holds the per-filter state and lock files.
	StateDir string
	// LogDir is the directory the per-run log of rejected records is created in.
	LogDir string
}

// Updater fetches works changed in Crossref one day window at a time,
// resuming after the last completed day recorded in the filter's state,
// and hands every valid work to a PaperWriter.
type Updater struct {
	cfg     UpdaterConfig
	log     *slog.Logger
	fetcher PageFetcher
	writer  PaperWriter
}

// NewUpdater constructs an Updater for the given configuration.
func NewUpdater(cfg UpdaterConfig, log *slog.Logger, fetcher PageFetcher, writer PaperWriter) *Updater {
	return &Updater{cfg: cfg, log: log, fetcher: fetcher, writer: writer}
}

// Run crawls every open day up to cfg.Until and appends one state entry
// per day. It stops at the first day that fails or is interrupted, so the
// next run retries that day.
func (u *Updater) Run(ctx context.Context) error {
	filter := string(u.cfg.Filter)
	unlock, err := syncstate.Lock(u.cfg.StateDir, filter)
	if err != nil {
		return err
	}
	defer func() {
		if err := unlock(); err != nil {
			u.log.Error("remove lock file", slog.String("error", err.Error()))
		}
	}()

	store := syncstate.NewStore(u.cfg.StateDir, filter)
	from, err := store.NextFrom(u.cfg.From)
	if err != nil {
		return err
	}
	if from.After(u.cfg.Until) {
		u.log.Info("crossref sync up to date", slog.String("filter", filter), slog.String("until", u.cfg.Until.Format(time.DateOnly)))
		return nil
	}

	start := time.Now()
	logFile, err := createLogFile(u.cfg.LogDir, start, filter)
	if err != nil {
		return err
	}
	defer logFile.Close()

	runID := start.Format("20060102-150405")
	u.log.Info("crossref sync starting",
		slog.String("filter", filter),
		slog.String("from", from.Format(time.DateOnly)),
		slog.String("until", u.cfg.Until.Format(time.DateOnly)),
		slog.String("log_file", logFile.Name()),
	)

	for day := from; !day.After(u.cfg.Until); day = day.AddDate(0, 0, 1) {
		entry, dayErr := u.syncDay(ctx, day, logFile)
		entry.RunID = runID
		if err := store.Append(entry); err != nil {
			return errors.Join(dayErr, err)
		}
		u.log.Info("window complete",
			slog.String("filter", filter),
			slog.String("day", entry.WindowFrom),
			slog.String("status", string(entry.Status)),
			slog.Int("total_results", entry.TotalResults),
			slog.Int("written", entry.Written),
			slog.Int("invalid", entry.Invalid),
			slog.Int("parse_errors", entry.ParseErrors),
			slog.Int("requests", entry.Requests),
		)
		if dayErr != nil {
			return fmt.Errorf("window %s: %w", entry.WindowFrom, dayErr)
		}
	}

	u.log.Info("crossref sync finished", slog.String("filter", filter), slog.Duration("elapsed", time.Since(start)))
	return nil
}

// syncDay pages through all works of one day and returns its state entry.
// The returned error is also recorded in the entry.
func (u *Updater) syncDay(ctx context.Context, day time.Time, logFile io.Writer) (syncstate.Entry, error) {
	d := day.Format(time.DateOnly)
	e := syncstate.Entry{
		Filter:      string(u.cfg.Filter),
		WindowFrom:  d,
		WindowUntil: d,
		StartedAt:   time.Now().UTC(),
	}

	err := u.fetchDay(ctx, day, &e, logFile)
	e.FinishedAt = time.Now().UTC()
	switch {
	case err == nil:
		e.Status = syncstate.StatusCompleted
	case ctx.Err() != nil:
		e.Status = syncstate.StatusInterrupted
		e.Error = err.Error()
	default:
		e.Status = syncstate.StatusFailed
		e.Error = err.Error()
	}
	return e, err
}

func (u *Updater) fetchDay(ctx context.Context, day time.Time, e *syncstate.Entry, logFile io.Writer) error {
	cursor := api.FirstCursor
	for {
		page, err := u.fetcher.FetchPage(ctx, u.cfg.Filter, day, cursor)
		e.Requests++
		if err != nil {
			return err
		}
		if cursor == api.FirstCursor {
			e.TotalResults = page.TotalResults
		}
		if len(page.Items) == 0 {
			return nil
		}

		for _, item := range page.Items {
			e.Fetched++
			if err := u.writeItem(ctx, item, e, logFile); err != nil {
				return err
			}
		}

		if page.NextCursor == "" {
			// Without a cursor the next request would restart at page one.
			return fmt.Errorf("crossref returned no next-cursor after %d items", e.Fetched)
		}
		cursor = page.NextCursor
		if err := ctx.Err(); err != nil {
			return err
		}
	}
}

// writeItem parses, validates and writes one work. Unusable records are
// counted and logged; only a writer error is returned.
func (u *Updater) writeItem(ctx context.Context, item json.RawMessage, e *syncstate.Entry, logFile io.Writer) error {
	rec, err := parser.Parse(item)
	if err != nil {
		e.ParseErrors++
		u.log.Warn("parse error", slog.String("day", e.WindowFrom), slog.String("error", err.Error()))
		u.writeLogEntry(logFile, e.WindowFrom, "", err.Error())
		return nil
	}

	p := toPaper(rec)
	if problems := invalidFields(p); problems != "" {
		e.Invalid++
		u.log.Debug("invalid paper", slog.String("day", e.WindowFrom), slog.String("doi", rec.DOI), slog.String("error", problems))
		u.writeLogEntry(logFile, e.WindowFrom, rec.DOI, problems)
		return nil
	}

	if err := u.writer.Upsert(ctx, p); err != nil {
		return fmt.Errorf("write paper %s: %w", rec.DOI, err)
	}
	e.Written++
	return nil
}

// writeLogEntry appends one line to the run's log file describing a
// record that could not be written, e.g.:
//
//	time=2026-10-13T03:00:05.123+02:00 filter=update-date day=2026-03-01 doi=10.1234/abc error=missing title
func (u *Updater) writeLogEntry(w io.Writer, day, doi, msg string) {
	now := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
	if _, err := fmt.Fprintf(w, "time=%s filter=%s day=%s doi=%s error=%s\n",
		now, u.cfg.Filter, day, doi, msg); err != nil {
		u.log.Error("write sync log", slog.String("error", err.Error()))
	}
}
