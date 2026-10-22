package importer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/importer/crossref/api"
	"github.com/paperstacks.io/paperstacks/internal/importer/syncstate"
	paper "github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

// fakeFetcher serves pages keyed by "<day>/<cursor>" and records the days
// it was asked for. Days listed in fail return an error.
type fakeFetcher struct {
	pages map[string]api.Page
	fail  map[string]bool
	days  []string
}

func (f *fakeFetcher) FetchPage(_ context.Context, _ api.Filter, day time.Time, cursor string) (api.Page, error) {
	d := day.Format(time.DateOnly)
	if cursor == api.FirstCursor {
		f.days = append(f.days, d)
	}
	if f.fail[d] {
		return api.Page{}, errors.New("crossref unavailable")
	}
	if page, ok := f.pages[d+"/"+cursor]; ok {
		return page, nil
	}
	return api.Page{NextCursor: "end"}, nil
}

type recordingWriter struct {
	dois []string
}

func (w *recordingWriter) Upsert(_ context.Context, p paper.Paper) error {
	w.dois = append(w.dois, p.DOI)
	return nil
}

func items(raw ...string) []json.RawMessage {
	out := make([]json.RawMessage, len(raw))
	for i, r := range raw {
		out[i] = json.RawMessage(r)
	}
	return out
}

func date(s string) time.Time {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}
	return t
}

func newTestUpdater(t *testing.T, stateDir string, fetcher PageFetcher, writer PaperWriter, until string) *Updater {
	t.Helper()
	cfg := UpdaterConfig{
		Filter:   api.FilterUpdateDate,
		From:     date("2026-03-01"),
		Until:    date(until),
		StateDir: stateDir,
		LogDir:   filepath.Join(t.TempDir(), "logs"),
	}
	return NewUpdater(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), fetcher, writer)
}

func TestUpdaterSyncsEachDayAndRecordsState(t *testing.T) {
	fetcher := &fakeFetcher{pages: map[string]api.Page{
		"2026-03-01/*": {TotalResults: 3, NextCursor: "c1", Items: items(
			`{"DOI":"10.1/valid-1","title":["Valid 1"],"type":"journal-article"}`,
			`{"DOI":"10.1/no-title"}`,
		)},
		"2026-03-01/c1": {TotalResults: 3, NextCursor: "c2", Items: items(
			`{"title":["No DOI"]}`,
		)},
		"2026-03-02/*": {TotalResults: 1, NextCursor: "c1", Items: items(
			`{"DOI":"10.1/valid-2","title":["Valid 2"],"type":"journal-article"}`,
		)},
	}}
	writer := &recordingWriter{}
	stateDir := t.TempDir()

	if err := newTestUpdater(t, stateDir, fetcher, writer, "2026-03-02").Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(writer.dois) != 2 || writer.dois[0] != "10.1/valid-1" || writer.dois[1] != "10.1/valid-2" {
		t.Errorf("written DOIs = %v, want [10.1/valid-1 10.1/valid-2]", writer.dois)
	}

	entries, err := syncstate.NewStore(stateDir, "update-date").Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d state entries, want 2", len(entries))
	}
	first := entries[0]
	if first.WindowFrom != "2026-03-01" || first.Status != syncstate.StatusCompleted ||
		first.TotalResults != 3 || first.Fetched != 3 || first.Written != 1 ||
		first.Invalid != 1 || first.ParseErrors != 1 || first.Requests != 3 {
		t.Errorf("first entry = %+v", first)
	}
	if entries[0].RunID == "" || entries[0].RunID != entries[1].RunID {
		t.Errorf("run IDs = %q, %q, want equal and non-empty", entries[0].RunID, entries[1].RunID)
	}
}

func TestUpdaterResumesAtFailedDay(t *testing.T) {
	stateDir := t.TempDir()
	failing := &fakeFetcher{fail: map[string]bool{"2026-03-02": true}}

	err := newTestUpdater(t, stateDir, failing, NoopWriter{}, "2026-03-03").Run(context.Background())
	if err == nil {
		t.Fatal("Run: want error for failed day")
	}
	if len(failing.days) != 2 {
		t.Errorf("fetched days = %v, want stop after 2026-03-02", failing.days)
	}

	entries, err := syncstate.NewStore(stateDir, "update-date").Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(entries) != 2 || entries[1].Status != syncstate.StatusFailed || entries[1].Error == "" {
		t.Fatalf("entries = %+v, want completed + failed with error", entries)
	}

	retry := &fakeFetcher{}
	if err := newTestUpdater(t, stateDir, retry, NoopWriter{}, "2026-03-03").Run(context.Background()); err != nil {
		t.Fatalf("second Run: %v", err)
	}
	if len(retry.days) != 2 || retry.days[0] != "2026-03-02" || retry.days[1] != "2026-03-03" {
		t.Errorf("second run fetched days = %v, want [2026-03-02 2026-03-03]", retry.days)
	}
}

func TestUpdaterRefusesWhenLocked(t *testing.T) {
	stateDir := t.TempDir()
	unlock, err := syncstate.Lock(stateDir, "update-date")
	if err != nil {
		t.Fatalf("Lock: %v", err)
	}
	defer unlock()

	fetcher := &fakeFetcher{}
	err = newTestUpdater(t, stateDir, fetcher, NoopWriter{}, "2026-03-01").Run(context.Background())
	if !errors.Is(err, syncstate.ErrLocked) {
		t.Errorf("Run err = %v, want ErrLocked", err)
	}
	if len(fetcher.days) != 0 {
		t.Errorf("fetched days = %v, want none", fetcher.days)
	}
}
