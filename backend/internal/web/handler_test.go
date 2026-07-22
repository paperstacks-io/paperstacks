package web

import (
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	stackApp "github.com/paperstacks.io/paperstacks/internal/stack/application"
	stackDomain "github.com/paperstacks.io/paperstacks/internal/stack/domain"
	stackMemory "github.com/paperstacks.io/paperstacks/internal/stack/repository/memory"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

func TestPageNameFromPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "root path",
			path: "/",
			want: "home",
		},
		{
			name: "empty path",
			path: "",
			want: "home",
		},
		{
			name: "single segment",
			path: "/papers",
			want: "papers",
		},
		{
			name: "nested path",
			path: "/stacks/my",
			want: "stacks/my",
		},
		{
			name: "trailing slash",
			path: "/stacks/",
			want: "stacks",
		},
		{
			name: "no leading slash",
			path: "settings",
			want: "settings",
		},
		{
			name: "no query param",
			path: "/stacks/my?sort=desc",
			want: "stacks/my",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := pageNameFromPath(tt.path)
			if got != tt.want {
				t.Fatalf("pageNameFromPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestHandleStacksDetailPageRedirectsToStacksPageWhenStackNotFoundHTMX(t *testing.T) {
	t.Parallel()

	handler := handleStacksDetailPage(
		testLogger(),
		testWebTemplate(t),
		"",
		stackApp.NewStackService(stackMemory.NewRepository(), nil, nil),
	)
	req := newStackDetailRequest(t, "00000000-0000-4000-8000-000000000000")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if got := rr.Header().Get("HX-Redirect"); got != stacksPageURL {
		t.Fatalf("HX-Redirect = %q, want %q", got, stacksPageURL)
	}
}

func TestHandleStacksDetailPageRedirectsToStacksPageWhenStackNotFound(t *testing.T) {
	t.Parallel()

	handler := handleStacksDetailPage(
		testLogger(),
		testWebTemplate(t),
		"",
		stackApp.NewStackService(stackMemory.NewRepository(), nil, nil),
	)
	req := newStackDetailRequest(t, "00000000-0000-4000-8000-000000000001")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != stacksPageURL {
		t.Fatalf("Location = %q, want %q", got, stacksPageURL)
	}
}

func TestHandleSidebarStackCreateRedirectsToDetail(t *testing.T) {
	t.Parallel()

	stackService := stackApp.NewStackService(stackMemory.NewRepository(), nil, nil)
	handler := handleSidebarStackCreate(testLogger(), testWebTemplate(t), stackService)

	req := newSidebarStackCreateRequest(t, "owner-sidebar-create", "sidebar@example.com", "Sidebar Stack")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	redirect := rr.Header().Get("HX-Redirect")
	if !strings.HasPrefix(redirect, "/app/stacks/detail/") {
		t.Fatalf("HX-Redirect = %q, want /app/stacks/detail/{uuid}", redirect)
	}

	uuid := strings.TrimPrefix(redirect, "/app/stacks/detail/")
	created, err := stackService.GetByUUID(req.Context(), uuid)
	if err != nil {
		t.Fatalf("created stack not found by redirect uuid %q: %v", uuid, err)
	}
	if created.Name != "Sidebar Stack" {
		t.Fatalf("created stack name = %q, want %q", created.Name, "Sidebar Stack")
	}
}
func TestHandleSidebarStackCreateRendersToastOnEmptyName(t *testing.T) {
	t.Parallel()

	handler := handleSidebarStackCreate(
		testLogger(),
		testWebTemplate(t),
		stackApp.NewStackService(stackMemory.NewRepository(), nil, nil),
	)

	req := newSidebarStackCreateRequest(t, "owner-empty-sidebar-create", "empty@example.com", "   ")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assertSidebarStackCreateError(t, rr, http.StatusUnprocessableEntity, "Stack name must be 1-80 characters")
}

func TestHandleSidebarStackCreateRendersToastOnDuplicateName(t *testing.T) {
	t.Parallel()

	stackService := stackApp.NewStackService(stackMemory.NewRepository(), nil, nil)
	user := userDomain.NewUser("owner-duplicate-sidebar-create", "duplicate@example.com")
	existing := stackDomain.NewStack("Duplicate Stack", user)
	if err := stackService.Create(t.Context(), *existing); err != nil {
		t.Fatalf("seed duplicate stack: %v", err)
	}

	handler := handleSidebarStackCreate(testLogger(), testWebTemplate(t), stackService)
	req := newSidebarStackCreateRequest(t, user.ExternalID, user.Email, "duplicate stack")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assertSidebarStackCreateError(t, rr, http.StatusConflict, "A stack with this name already exists.")
}

func TestHandleSidebarStackCreateRendersToastWithoutSession(t *testing.T) {
	t.Parallel()

	handler := handleSidebarStackCreate(
		testLogger(),
		testWebTemplate(t),
		stackApp.NewStackService(stackMemory.NewRepository(), nil, nil),
	)
	req := httptest.NewRequest(http.MethodPost, "/app/stacks/sidebar/create", strings.NewReader(url.Values{
		"name": {"Sidebar Stack"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assertSidebarStackCreateError(t, rr, http.StatusUnauthorized, "Unauthorized. Please log in to create a stack.")
}

func newSidebarStackCreateRequest(t *testing.T, userID string, email string, name string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/app/stacks/sidebar/create", strings.NewReader(url.Values{
		"name": {name},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(commonauth.ContextWithSession(req.Context(), &commonauth.Session{
		UserID:  userID,
		Email:   email,
		IsValid: true,
	}))

	return req
}

func newStackDetailRequest(t *testing.T, stackUUID string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/app/stacks/detail/"+stackUUID, nil)
	req.SetPathValue("uuid", stackUUID)
	req = req.WithContext(commonauth.ContextWithSession(req.Context(), &commonauth.Session{
		UserID:  "owner-detail",
		Email:   "detail@example.com",
		IsValid: true,
	}))

	return req
}

func assertSidebarStackCreateError(t *testing.T, rr *httptest.ResponseRecorder, status int, message string) {
	t.Helper()

	if rr.Code != status {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, status, rr.Body.String())
	}
	if got := rr.Header().Get("HX-Retarget"); got != "#toast-host" {
		t.Fatalf("HX-Retarget = %q, want #toast-host", got)
	}
	if got := rr.Header().Get("HX-Reswap"); got != "innerHTML" {
		t.Fatalf("HX-Reswap = %q, want innerHTML", got)
	}
	if body := rr.Body.String(); !strings.Contains(body, message) {
		t.Fatalf("body does not contain %q: %s", message, body)
	}
}

func testWebTemplate(t *testing.T) *template.Template {
	t.Helper()

	templateFiles, err := templateFiles(content)
	if err != nil {
		t.Fatalf("list templates: %v", err)
	}

	tmpl, err := template.ParseFS(content, templateFiles...)
	if err != nil {
		t.Fatalf("parse templates: %v", err)
	}

	return tmpl
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
