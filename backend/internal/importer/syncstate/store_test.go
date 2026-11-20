package syncstate

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var baseline = time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC)

func TestNextDayAfterEmptyStateReturnsBaseline(t *testing.T) {
	got, err := NewFilterStorage(t.TempDir(), "update-date").NextDayAfter(baseline)
	if err != nil {
		t.Fatalf("NextDayAfter: %v", err)
	}
	if !got.Equal(baseline) {
		t.Errorf("NextDayAfter = %s, want %s", got, baseline)
	}
}

func TestNextDayAfterResumesAfterLastCompletedWindow(t *testing.T) {
	s := NewFilterStorage(t.TempDir(), "update-date")
	for _, e := range []Entry{
		{WindowFrom: "2026-03-01", WindowUntil: "2026-03-01", Status: StatusCompleted},
		{WindowFrom: "2026-03-02", WindowUntil: "2026-03-02", Status: StatusCompleted},
		{WindowFrom: "2026-03-03", WindowUntil: "2026-03-03", Status: StatusFailed},
	} {
		if err := s.Append(e); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	got, err := s.NextDayAfter(baseline)
	if err != nil {
		t.Fatalf("NextDayAfter: %v", err)
	}
	if want := time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("NextDayAfter = %s, want %s", got, want)
	}

	entries, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(entries) != 3 || entries[2].Status != StatusFailed {
		t.Errorf("Load = %+v, want 3 entries ending with failed", entries)
	}
}

func TestAppendWritesSlogLineThatLoadsBack(t *testing.T) {
	dir := t.TempDir()
	want := Entry{
		RunID:        "20261111-030000",
		Filter:       "update-date",
		WindowFrom:   "2026-03-01",
		WindowUntil:  "2026-03-01",
		Status:       StatusFailed,
		StartedAt:    time.Date(2026, 11, 11, 3, 0, 0, 0, time.UTC),
		FinishedAt:   time.Date(2026, 11, 11, 3, 2, 41, 0, time.UTC),
		TotalResults: 65382,
		Fetched:      12000,
		Written:      11991,
		Invalid:      8,
		ParseErrors:  1,
		Requests:     13,
		Error:        "crossref unavailable",
	}
	s := NewFilterStorage(dir, "update-date")
	if err := s.Append(want); err != nil {
		t.Fatalf("Append: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "update-date.jsonl"))
	if err != nil {
		t.Fatalf("read state file: %v", err)
	}
	var line map[string]any
	if err := json.Unmarshal(raw, &line); err != nil {
		t.Fatalf("state line is not JSON: %v", err)
	}
	if line["level"] != "INFO" || line["msg"] != "window complete" || line["time"] == nil {
		t.Errorf("state line %s lacks slog time/level/msg", raw)
	}

	entries, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(entries) != 1 || entries[0] != want {
		t.Errorf("Load = %+v, want %+v", entries, want)
	}
}

func TestFilterStoragesAreSeparatePerFilter(t *testing.T) {
	dir := t.TempDir()
	if err := NewFilterStorage(dir, "update-date").Append(Entry{WindowUntil: "2026-03-10", Status: StatusCompleted}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	got, err := NewFilterStorage(dir, "index-date").NextDayAfter(baseline)
	if err != nil {
		t.Fatalf("NextDayAfter: %v", err)
	}
	if !got.Equal(baseline) {
		t.Errorf("index-date NextDayAfter = %s, want baseline %s", got, baseline)
	}
}

func TestLockIsExclusivePerFilter(t *testing.T) {
	dir := t.TempDir()
	unlock, err := Lock(dir, "index-date")
	if err != nil {
		t.Fatalf("Lock: %v", err)
	}

	if _, err := Lock(dir, "index-date"); !errors.Is(err, ErrLocked) {
		t.Errorf("second Lock err = %v, want ErrLocked", err)
	}
	other, err := Lock(dir, "update-date")
	if err != nil {
		t.Fatalf("Lock other filter: %v", err)
	}
	other()

	if err := unlock(); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	again, err := Lock(dir, "index-date")
	if err != nil {
		t.Fatalf("Lock after unlock: %v", err)
	}
	again()
}
