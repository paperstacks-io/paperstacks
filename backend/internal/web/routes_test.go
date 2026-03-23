package web

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAddRouteRendersResponsiveLandingShell(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	if err := AddRoute(mux, logger); err != nil {
		t.Fatalf("AddRoute() error = %v", err)
	}

	pages := []struct {
		path   string
		title  string
		active string
		body   string
	}{
		{path: "/app/", title: "<title>Paperstacks · Paperstacks</title>", active: `href="/app/"`, body: "Collect and revisit the papers that matter."},
		{path: "/app/papers", title: "<title>Papers · Paperstacks</title>", active: `href="/app/papers"`, body: "Paper page here"},
		{path: "/app/search", title: "<title>Search · Paperstacks</title>", active: `href="/app/search"`, body: "Search page here"},
		{path: "/app/settings", title: "<title>Settings · Paperstacks</title>", active: `href="/app/settings"`, body: "Settings page here"},
	}

	for _, page := range pages {
		page := page
		t.Run(page.path, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, page.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
			}

			body := rec.Body.String()
			if !strings.Contains(body, page.active) || !strings.Contains(body, `bg-sidebar-accent font-medium text-sidebar-accent-foreground`) {
				t.Errorf("response body missing active navigation styling for %q", page.path)
			}

			for _, want := range []string{
				`href="/app/assets/styles.css"`,
				`src="/app/assets/htmx.min.js"`,
				`id="page-content"`,
				`id="sidebar-backdrop"`,
				`class="fixed inset-0 z-40 hidden bg-foreground/40 lg:hidden"`,
				"Menu",
				`aria-controls="mobile-sidebar"`,
				"mobile-sidebar",
				`class="fixed inset-y-0 left-0 z-50 flex w-72 -translate-x-full flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground shadow-lg transition-transform lg:static lg:z-auto lg:translate-x-0 lg:shadow-none"`,
				`class="flex items-center gap-3 border-b border-border bg-background px-4 py-4 sm:px-6 lg:hidden"`,
				`href="/app/"`,
				`hx-get="/app/"`,
				`hx-target="#page-content"`,
				`hx-select="#page-content"`,
				`hx-swap="outerHTML"`,
				`hx-push-url="true"`,
				`href="/app/papers"`,
				`href="/app/search"`,
				`href="/app/settings"`,
				page.body,
				page.title,
			} {
				if !strings.Contains(body, want) {
					t.Errorf("response body missing %q", want)
				}
			}

		})
	}
}

func TestAddRouteRendersPageContentFragmentForHTMXRequests(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	if err := AddRoute(mux, logger); err != nil {
		t.Fatalf("AddRoute() error = %v", err)
	}

	tests := []struct {
		path string
		want string
	}{
		{path: "/app/", want: "Collect and revisit the papers that matter."},
		{path: "/app/papers", want: "Paper page here"},
		{path: "/app/search", want: "Search page here"},
		{path: "/app/settings", want: "Settings page here"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.path, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, test.path, nil)
			req.Header.Set("HX-Request", "true")
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
			}

			body := rec.Body.String()
			if !strings.Contains(body, `id="page-content"`) {
				t.Fatalf("HTMX response missing page content wrapper")
			}
			if !strings.Contains(body, test.want) {
				t.Fatalf("HTMX response missing page body %q", test.want)
			}
			for _, unwanted := range []string{"<!doctype html>", `id="mobile-sidebar"`, `href="/app/assets/styles.css"`} {
				if strings.Contains(body, unwanted) {
					t.Fatalf("HTMX response unexpectedly included %q", unwanted)
				}
			}
		})
	}
}

func TestAddRouteServesAssetsUnderAppNamespace(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	if err := AddRoute(mux, logger); err != nil {
		t.Fatalf("AddRoute() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/app/assets/styles.css", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "text/css") {
		t.Fatalf("content type = %q, want CSS content type", contentType)
	}

	if body := rec.Body.String(); !strings.Contains(body, "tailwindcss") {
		t.Fatalf("asset body missing expected stylesheet content")
	}
}

func TestAddRouteDoesNotHandleUnknownOrAPINamespacePaths(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	if err := AddRoute(mux, logger); err != nil {
		t.Fatalf("AddRoute() error = %v", err)
	}

	for _, path := range []string{"/missing", "/app/missing", "/papers"} {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("status code = %d, want %d", rec.Code, http.StatusNotFound)
			}

			if strings.Contains(rec.Body.String(), "Home") {
				t.Fatalf("unexpectedly rendered shell page for %q", path)
			}
		})
	}
}
