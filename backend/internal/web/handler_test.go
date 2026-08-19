package web

import (
	"context"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	stackApp "github.com/paperstacks.io/paperstacks/internal/stack/application"
	stackDomain "github.com/paperstacks.io/paperstacks/internal/stack/domain"
	stackMemory "github.com/paperstacks.io/paperstacks/internal/stack/repository/memory"
	userApplication "github.com/paperstacks.io/paperstacks/internal/user/application"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
	userMemory "github.com/paperstacks.io/paperstacks/internal/user/repository/memory"
)

func newTestStackService() *stackApp.StackService {
	return stackApp.NewStackService(stackMemory.NewRepository(), nil)
}

func newTestUserService(users ...userDomain.User) *userApplication.UserService {
	repo := userMemory.NewRepository()
	for _, user := range users {
		_, _ = repo.SaveIfNotExist(context.Background(), user)
	}

	return userApplication.NewUserService(repo)
}

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
		newTestStackService(),
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
		newTestStackService(),
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

func TestHandleStackPaperRemoveRemovesPaperAndRedirectsToDetail(t *testing.T) {
	t.Parallel()

	stackService := newTestStackService()
	user := userDomain.NewUser("owner-paper-remove", "paper-remove@example.com")
	stack := stackDomain.NewStack("Paper Remove Stack", user)
	removedPaperUUID := "11111111-1111-4111-8111-111111111111"
	retainedPaperUUID := "22222222-2222-4222-8222-222222222222"
	stack.Papers = []paperDomain.Paper{
		{UUID: removedPaperUUID, Title: "Removed Paper"},
		{UUID: retainedPaperUUID, Title: "Retained Paper"},
	}
	if err := stackService.Create(t.Context(), *stack); err != nil {
		t.Fatalf("seed stack: %v", err)
	}

	handler := handleStackPaperRemove(testLogger(), testWebTemplate(t), stackService)
	req := newStackPaperRemoveRequest(t, stack.UUID, removedPaperUUID, user.ExternalID, user.Email)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if got := rr.Header().Get("HX-Redirect"); got != "/app/stacks/detail/"+stack.UUID {
		t.Fatalf("HX-Redirect = %q, want detail page redirect", got)
	}

	updated, err := stackService.GetByUUID(req.Context(), stack.UUID)
	if err != nil {
		t.Fatalf("GetByUUID() error = %v", err)
	}
	if len(updated.Papers) != 1 {
		t.Fatalf("updated papers length = %d, want 1", len(updated.Papers))
	}
	if updated.Papers[0].UUID != retainedPaperUUID {
		t.Fatalf("remaining paper UUID = %q, want %q", updated.Papers[0].UUID, retainedPaperUUID)
	}
}

func TestHandleSidebarStackCreateRedirectsToDetail(t *testing.T) {
	t.Parallel()

	user := userDomain.NewUser("owner-sidebar-create", "sidebar@example.com")
	stackService := newTestStackService()
	handler := handleSidebarStackCreate(testLogger(), testWebTemplate(t), newTestUserService(user), stackService)

	req := newSidebarStackCreateRequest(t, user.ExternalID, user.Email, "Sidebar Stack")
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
	if created.Owner != user {
		t.Fatalf("created stack owner = %#v, want %#v", created.Owner, user)
	}
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

func newStackPaperRemoveRequest(t *testing.T, stackUUID string, paperUUID string, userID string, email string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/app/stacks/detail/"+stackUUID+"/papers/"+paperUUID+"/remove", nil)
	req.Header.Set("HX-Request", "true")
	req.SetPathValue("uuid", stackUUID)
	req.SetPathValue("paperUUID", paperUUID)
	req = req.WithContext(commonauth.ContextWithSession(req.Context(), &commonauth.Session{
		UserID:  userID,
		Email:   email,
		IsValid: true,
	}))

	return req
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
