// Package syncstate records which day windows of a Crossref sync have
// been crawled. Each filter keeps its own append-only JSONL file with one
// Entry per processed window, so completed, failed and still-open days can
// be read straight from the file, and a crash never corrupts earlier
// entries.
package syncstate

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
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

// Store is the state file of one filter, <dir>/<filter>.jsonl.
type Store struct {
	path string
}

// NewStore returns the Store for filter in dir.
func NewStore(dir, filter string) *Store {
	return &Store{path: filepath.Join(dir, filter+".jsonl")}
}

// Append adds e as a new line, creating dir and the file if needed.
func (s *Store) Append(e Entry) error {
	line, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("encode state entry: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open state file: %w", err)
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		f.Close()
		return fmt.Errorf("write state file: %w", err)
	}
	return f.Close()
}

// Load returns all entries in file order, or none if the file does not
// exist yet.
func (s *Store) Load() ([]Entry, error) {
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

// NextFrom returns the day after the latest completed window, or baseline
// if no window has been completed yet.
func (s *Store) NextFrom(baseline time.Time) (time.Time, error) {
	entries, err := s.Load()
	if err != nil {
		return time.Time{}, err
	}
	var last time.Time
	for _, e := range entries {
		if e.Status != StatusCompleted {
			continue
		}
		until, err := time.Parse(time.DateOnly, e.WindowUntil)
		if err != nil {
			return time.Time{}, fmt.Errorf("state entry window_until %q: %w", e.WindowUntil, err)
		}
		if until.After(last) {
			last = until
		}
	}
	if last.IsZero() {
		return baseline, nil
	}
	return last.AddDate(0, 0, 1), nil
}
