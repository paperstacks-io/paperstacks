// Package syncstate records which day windows of a Crossref sync have
// been crawled. Each filter keeps its own append-only JSONL file with one
// Entry per processed window, written by slog's JSON handler like the
// repository's other logs, so completed, failed and still-open days can
// be read straight from the file, and a crash never corrupts earlier
// entries.
package syncstate

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// Status is the outcome of crawling one window.
type Status string

const (
	StatusCompleted   Status = "completed"
	StatusFailed      Status = "failed"
	StatusInterrupted Status = "interrupted"
)

// Entry describes one crawled window. WindowFrom and WindowUntil are
// inclusive days formatted as YYYY-MM-DD.
type Entry struct {
	RunID        string    `json:"run_id"`
	Filter       string    `json:"filter"`
	WindowFrom   string    `json:"window_from"`
	WindowUntil  string    `json:"window_until"`
	Status       Status    `json:"status"`
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   time.Time `json:"finished_at"`
	TotalResults int       `json:"total_results"`
	Fetched      int       `json:"fetched"`
	Written      int       `json:"written"`
	Invalid      int       `json:"invalid"`
	ParseErrors  int       `json:"parse_errors"`
	Requests     int       `json:"requests"`
	Error        string    `json:"error"`
}

// LogValue implements slog.LogValuer. Keys must match the json tags above,
// since Load decodes the logged lines back into Entry.
func (e Entry) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("run_id", e.RunID),
		slog.String("filter", e.Filter),
		slog.String("window_from", e.WindowFrom),
		slog.String("window_until", e.WindowUntil),
		slog.String("status", string(e.Status)),
		slog.Time("started_at", e.StartedAt),
		slog.Time("finished_at", e.FinishedAt),
		slog.Int("total_results", e.TotalResults),
		slog.Int("fetched", e.Fetched),
		slog.Int("written", e.Written),
		slog.Int("invalid", e.Invalid),
		slog.Int("parse_errors", e.ParseErrors),
		slog.Int("requests", e.Requests),
		slog.String("error", e.Error),
	)
}

// FilterStorage is the state file of one filter, <dir>/<filter>.jsonl.
type FilterStorage struct {
	path string
}

// NewFilterStorage returns the FilterStorage for filter in dir.
func NewFilterStorage(dir, filter string) *FilterStorage {
	return &FilterStorage{path: filepath.Join(dir, filter+".jsonl")}
}

// Append logs e as a new JSON line, creating dir and the file if needed.
// The entry's fields sit at the top level next to slog's time, level and
// msg keys, which Load ignores.
func (s *FilterStorage) Append(e Entry) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open state file: %w", err)
	}
	// Handle is called directly instead of via slog.Logger, which would
	// silently drop write errors.
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "window complete", 0)
	r.AddAttrs(slog.Attr{Key: "", Value: e.LogValue()})
	if err := slog.NewJSONHandler(f, nil).Handle(context.Background(), r); err != nil {
		f.Close()
		return fmt.Errorf("write state file: %w", err)
	}
	return f.Close()
}

// Load returns all entries in file order, or none if the file does not
// exist yet.
func (s *FilterStorage) Load() ([]Entry, error) {
	f, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open state file: %w", err)
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	for line := 1; scanner.Scan(); line++ {
		var e Entry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			return nil, fmt.Errorf("state file %s line %d: %w", s.path, line, err)
		}
		entries = append(entries, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read state file: %w", err)
	}
	return entries, nil
}

// NextDayAfter returns the day after the latest completed window, or baseline
// if no window has been completed yet. Completed windows are appended in
// ascending order, so the last completed entry is the latest one.
func (s *FilterStorage) NextDayAfter(baseline time.Time) (time.Time, error) {
	entries, err := s.Load()
	if err != nil {
		return time.Time{}, err
	}
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if e.Status != StatusCompleted {
			continue
		}
		until, err := time.Parse(time.DateOnly, e.WindowUntil)
		if err != nil {
			return time.Time{}, fmt.Errorf("state entry window_until %q: %w", e.WindowUntil, err)
		}
		return until.AddDate(0, 0, 1), nil
	}
	return baseline, nil
}
