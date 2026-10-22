package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var day = time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

func TestFetchPageBuildsQuery(t *testing.T) {
	tests := []struct {
		name       string
		filter     Filter
		mailto     string
		wantFilter string
	}{
		{"update-date public pool", FilterUpdateDate, "", "from-update-date:2026-03-01,until-update-date:2026-03-01"},
		{"index-date polite pool", FilterIndexDate, "dev@example.org", "from-index-date:2026-03-01,until-index-date:2026-03-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				if r.URL.Path != "/works" {
					t.Errorf("path = %q, want /works", r.URL.Path)
				}
				if got := q.Get("filter"); got != tt.wantFilter {
					t.Errorf("filter = %q, want %q", got, tt.wantFilter)
				}
				if got := q.Get("rows"); got != "1000" {
					t.Errorf("rows = %q, want 1000", got)
				}
				if got := q.Get("cursor"); got != "abc" {
					t.Errorf("cursor = %q, want abc", got)
				}
				if got := q.Get("mailto"); got != tt.mailto {
					t.Errorf("mailto = %q, want %q", got, tt.mailto)
				}
				if ua := r.UserAgent(); tt.mailto != "" && !strings.Contains(ua, tt.mailto) {
					t.Errorf("User-Agent %q does not contain mailto", ua)
				}
				w.Write([]byte(`{"message":{"total-results":2,"next-cursor":"def","items":[{"DOI":"10.1/a"},{"DOI":"10.1/b"}]}}`))
			}))
			defer srv.Close()

			page, err := NewClient(srv.URL, tt.mailto, nil).FetchPage(context.Background(), tt.filter, day, "abc")
			if err != nil {
				t.Fatalf("FetchPage: %v", err)
			}
			if page.TotalResults != 2 || page.NextCursor != "def" || len(page.Items) != 2 {
				t.Errorf("page = %+v, want 2 results, cursor def, 2 items", page)
			}
		})
	}
}

func TestFetchPageRetriesRateLimited(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(`{"message":{"total-results":0,"next-cursor":"x","items":[]}}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "", nil)
	c.backoff = time.Millisecond
	if _, err := c.FetchPage(context.Background(), FilterUpdateDate, day, FirstCursor); err != nil {
		t.Fatalf("FetchPage: %v", err)
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2", calls)
	}
}

func TestFetchPageDoesNotRetryClientError(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, "invalid filter", http.StatusBadRequest)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "", nil)
	c.backoff = time.Millisecond
	_, err := c.FetchPage(context.Background(), FilterUpdateDate, day, FirstCursor)
	if err == nil || !strings.Contains(err.Error(), "invalid filter") {
		t.Fatalf("err = %v, want status 400 with body", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}
