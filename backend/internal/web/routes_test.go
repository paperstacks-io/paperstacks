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

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	for _, want := range []string{
		`id="sidebar-backdrop"`,
		`class="fixed inset-0 z-40 hidden bg-foreground/40 lg:hidden"`,
		"Menu",
		`aria-controls="mobile-sidebar"`,
		"mobile-sidebar",
		`class="fixed inset-y-0 left-0 z-50 flex w-72 -translate-x-full flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground shadow-lg transition-transform lg:static lg:z-auto lg:translate-x-0 lg:shadow-none"`,
		`class="flex items-center gap-3 border-b border-border bg-background px-4 py-4 sm:px-6 lg:hidden"`,
		"Home",
		"Papers",
		"Search",
		"Settings",
		"Collect and revisit the papers that matter.",
		"Contributor information",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("response body missing %q", want)
		}
	}
}
