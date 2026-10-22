package syncstate

import (
	"errors"
	"testing"
	"time"
)

var baseline = time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC)

func TestNextFromEmptyStateReturnsBaseline(t *testing.T) {
	got, err := NewStore(t.TempDir(), "update-date").NextFrom(baseline)
	if err != nil {
		t.Fatalf("NextFrom: %v", err)
	}
	if !got.Equal(baseline) {
		t.Errorf("NextFrom = %s, want %s", got, baseline)
	}
}

func TestNextFromResumesAfterLastCompletedWindow(t *testing.T) {
	s := NewStore(t.TempDir(), "update-date")
	for _, e := range []Entry{
		{WindowFrom: "2026-03-01", WindowUntil: "2026-03-01", Status: StatusCompleted},
		{WindowFrom: "2026-03-02", WindowUntil: "2026-03-02", Status: StatusCompleted},
		{WindowFrom: "2026-03-03", WindowUntil: "2026-03-03", Status: StatusFailed},
	} {
		if err := s.Append(e); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	got, err := s.NextFrom(baseline)
	if err != nil {
		t.Fatalf("NextFrom: %v", err)
	}
	if want := time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("NextFrom = %s, want %s", got, want)
	}

	entries, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(entries) != 3 || entries[2].Status != StatusFailed {
		t.Errorf("Load = %+v, want 3 entries ending with failed", entries)
	}
}

func TestStoresAreSeparatePerFilter(t *testing.T) {
	dir := t.TempDir()
	if err := NewStore(dir, "update-date").Append(Entry{WindowUntil: "2026-03-10", Status: StatusCompleted}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	got, err := NewStore(dir, "index-date").NextFrom(baseline)
	if err != nil {
		t.Fatalf("NextFrom: %v", err)
	}
	if !got.Equal(baseline) {
		t.Errorf("index-date NextFrom = %s, want baseline %s", got, baseline)
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
