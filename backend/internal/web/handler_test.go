package web

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	paperApp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	paperMemory "github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
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

func TestHandleStackPublicSettingUpdateSetsPublic(t *testing.T) {
	t.Parallel()

	stackService := newTestStackService()
	user := userDomain.NewUser("owner-public-setting-true", "public-true@example.com")
	stack := stackDomain.NewStack("Public Setting Stack", user)
	stack.IsPublic = false
	if err := stackService.Create(t.Context(), *stack); err != nil {
		t.Fatalf("seed stack: %v", err)
	}

	handler := handleStackPublicSettingUpdate(testLogger(), testWebTemplate(t), stackService)
	req := newStackPublicSettingRequest(t, stack.UUID, user.ExternalID, user.Email, true)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assertStackPublicSettingSuccess(t, rr)

	updated, err := stackService.GetByUUID(req.Context(), stack.UUID)
	if err != nil {
		t.Fatalf("GetByUUID() error = %v", err)
	}
	if !updated.IsPublic {
		t.Fatalf("updated IsPublic = false, want true")
	}
}

func TestHandleStackPublicSettingUpdateClearsPublic(t *testing.T) {
	t.Parallel()

	stackService := newTestStackService()
	user := userDomain.NewUser("owner-public-setting-false", "public-false@example.com")
	stack := stackDomain.NewStack("Private Setting Stack", user)
	stack.IsPublic = true
	if err := stackService.Create(t.Context(), *stack); err != nil {
		t.Fatalf("seed stack: %v", err)
	}

	handler := handleStackPublicSettingUpdate(testLogger(), testWebTemplate(t), stackService)
	req := newStackPublicSettingRequest(t, stack.UUID, user.ExternalID, user.Email, false)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assertStackPublicSettingSuccess(t, rr)

	updated, err := stackService.GetByUUID(req.Context(), stack.UUID)
	if err != nil {
		t.Fatalf("GetByUUID() error = %v", err)
	}
	if updated.IsPublic {
		t.Fatalf("updated IsPublic = true, want false")
	}
}

func TestHandleStackPublicSettingUpdateRejectsNonOwner(t *testing.T) {
	t.Parallel()

	stackService := newTestStackService()
	owner := userDomain.NewUser("owner-public-setting-forbidden", "owner-public@example.com")
	stack := stackDomain.NewStack("Forbidden Public Setting Stack", owner)
	stack.IsPublic = false
	if err := stackService.Create(t.Context(), *stack); err != nil {
		t.Fatalf("seed stack: %v", err)
	}

	handler := handleStackPublicSettingUpdate(testLogger(), testWebTemplate(t), stackService)
	req := newStackPublicSettingRequest(t, stack.UUID, "other-public-setting-forbidden", "other-public@example.com", true)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assertSidebarStackCreateError(t, rr, http.StatusOK, "You are not allowed to update this stack.")

	updated, err := stackService.GetByUUID(req.Context(), stack.UUID)
	if err != nil {
		t.Fatalf("GetByUUID() error = %v", err)
	}
	if updated.IsPublic {
		t.Fatalf("updated IsPublic = true, want unchanged false")
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

func TestHandleStackPaperRemoveRejectsNonOwner(t *testing.T) {
	t.Parallel()

	stackService := newTestStackService()
	owner := userDomain.NewUser("owner-paper-remove-forbidden", "owner-remove@example.com")
	stack := stackDomain.NewStack("Forbidden Paper Remove Stack", owner)
	paperUUID := "33333333-3333-4333-8333-333333333333"
	stack.Papers = []paperDomain.Paper{{UUID: paperUUID, Title: "Kept Paper"}}
	if err := stackService.Create(t.Context(), *stack); err != nil {
		t.Fatalf("seed stack: %v", err)
	}

	handler := handleStackPaperRemove(testLogger(), testWebTemplate(t), stackService)
	req := newStackPaperRemoveRequest(t, stack.UUID, paperUUID, "other-paper-remove-forbidden", "other-remove@example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assertSidebarStackCreateError(t, rr, http.StatusOK, "You are not allowed to remove papers from this stack.")

	updated, err := stackService.GetByUUID(req.Context(), stack.UUID)
	if err != nil {
		t.Fatalf("GetByUUID() error = %v", err)
	}
	if len(updated.Papers) != 1 {
		t.Fatalf("updated papers length = %d, want unchanged 1", len(updated.Papers))
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

func TestHandleSidebarStackCreateRendersToastOnEmptyName(t *testing.T) {
	t.Parallel()

	user := userDomain.NewUser("owner-empty-sidebar-create", "empty@example.com")
	handler := handleSidebarStackCreate(
		testLogger(),
		testWebTemplate(t),
		newTestUserService(user),
		newTestStackService(),
	)

	req := newSidebarStackCreateRequest(t, user.ExternalID, user.Email, "   ")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assertSidebarStackCreateError(t, rr, http.StatusOK, "Stack name must be 1-80 characters")
}

func TestHandleSidebarStackCreateRendersToastOnDuplicateName(t *testing.T) {
	t.Parallel()

	user := userDomain.NewUser("owner-duplicate-sidebar-create", "duplicate@example.com")
	stackService := newTestStackService()
	existing := stackDomain.NewStack("Duplicate Stack", user)
	if err := stackService.Create(t.Context(), *existing); err != nil {
		t.Fatalf("seed duplicate stack: %v", err)
	}

	handler := handleSidebarStackCreate(testLogger(), testWebTemplate(t), newTestUserService(user), stackService)
	req := newSidebarStackCreateRequest(t, user.ExternalID, user.Email, "duplicate stack")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assertSidebarStackCreateError(t, rr, http.StatusOK, "A stack with this name already exists.")
}

func TestHandleSidebarStackCreateRendersToastWithoutSession(t *testing.T) {
	t.Parallel()

	handler := handleSidebarStackCreate(
		testLogger(),
		testWebTemplate(t),
		newTestUserService(),
		newTestStackService(),
	)
	req := httptest.NewRequest(http.MethodPost, "/app/stacks/sidebar/create", strings.NewReader(url.Values{
		"name": {"Sidebar Stack"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assertSidebarStackCreateError(t, rr, http.StatusOK, "Unauthorized. Please log in to create a stack.")
}

func TestHandleSettingsPageRendersUserORCID(t *testing.T) {
	t.Parallel()

	userService := userApplication.NewUserService(userMemory.NewRepository())
	user, err := userService.CreateIfNotExist(t.Context(), "settings-page-user", "settings-page@example.com")
	if err != nil {
		t.Fatalf("CreateIfNotExist() error = %v", err)
	}
	user.ORCID = "0000-0002-1825-0097"
	if err := userService.Update(t.Context(), user.ExternalID, user); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	handler := handleSettingsPage(testLogger(), testPageTemplate(t, "settings/page"), "", userService)
	req := newSettingsRequest(t, user.ExternalID, user.Email)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	body := rr.Body.String()
	for _, want := range []string{
		`value="0000-0002-1825-0097"`,
		`data-initial-value="0000-0002-1825-0097"`,
		`data-user-settings-actions`,
		`hidden`,
		`hx-post="/app/settings/user"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("settings page missing %q: %s", want, body)
		}
	}
}

func TestHandleUserSettingsUpdateRendersToastOnInvalidORCID(t *testing.T) {
	t.Parallel()

	userService := userApplication.NewUserService(userMemory.NewRepository())
	user, err := userService.CreateIfNotExist(t.Context(), "settings-invalid-user", "settings-invalid@example.com")
	if err != nil {
		t.Fatalf("CreateIfNotExist() error = %v", err)
	}

	handler := handleUserSettingsUpdate(testLogger(), testWebTemplate(t), userService)
	req := newUserSettingsUpdateRequest(t, user.ExternalID, user.Email, "invalid")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assertSidebarStackCreateError(t, rr, http.StatusOK, "ORCID must use the 0000-0000-0000-0000 format.")

	stored, err := userService.GetByExternalID(t.Context(), user.ExternalID)
	if err != nil {
		t.Fatalf("GetByExternalID() error = %v", err)
	}
	if stored.ORCID != "" {
		t.Fatalf("stored ORCID = %q, want unchanged empty", stored.ORCID)
	}
}

func TestHandleStackBibLaTeXImportDisplaysCandidatesWithoutAddingToStack(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("../paper/bibliography/testdata/bauer.bib")
	if err != nil {
		t.Fatalf("read BibLaTeX fixture: %v", err)
	}
	paperService := paperApp.NewPaperService(paperMemory.NewRepository())
	stackService := stackApp.NewStackService(stackMemory.NewRepository(), paperService)
	owner := userDomain.NewUser("biblatex-import-owner", "biblatex-import@example.com")
	stack := stackDomain.NewStack("BibLaTeX Import", owner)
	if err := stackService.Create(t.Context(), *stack); err != nil {
		t.Fatalf("seed stack: %v", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("biblatex", "bauer.bib")
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}
	if _, err := part.Write(source); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/app/stacks/"+stack.UUID+"/import/biblatex", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetPathValue("uuid", stack.UUID)
	req = req.WithContext(commonauth.ContextWithSession(req.Context(), &commonauth.Session{
		UserID:  owner.ExternalID,
		Email:   owner.Email,
		IsValid: true,
	}))
	rr := httptest.NewRecorder()

	handleStackBibLaTeXImport(
		testLogger(),
		testWebTemplate(t),
		paperApp.NewBibliographyService(paperService),
		stackService,
	).ServeHTTP(rr, req)

	if !strings.Contains(rr.Body.String(), "When GUI-Based Testing of Web Applications Meets Code Review") {
		t.Fatalf("response does not contain imported title: %s", rr.Body.String())
	}
	updated, err := stackService.GetByUUID(t.Context(), stack.UUID)
	if err != nil {
		t.Fatalf("get imported stack: %v", err)
	}
	if len(updated.Papers) != 0 {
		t.Fatalf("stack papers length = %d, want 0", len(updated.Papers))
	}
}

func newSettingsRequest(t *testing.T, userID string, email string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/app/settings", nil)
	req = req.WithContext(commonauth.ContextWithSession(req.Context(), &commonauth.Session{
		UserID:  userID,
		Email:   email,
		IsValid: true,
	}))

	return req
}

func newUserSettingsUpdateRequest(t *testing.T, userID string, email string, orcid string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/app/settings/user", strings.NewReader(url.Values{
		"orcid": {orcid},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	req = req.WithContext(commonauth.ContextWithSession(req.Context(), &commonauth.Session{
		UserID:  userID,
		Email:   email,
		IsValid: true,
	}))

	return req
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

func newStackPublicSettingRequest(t *testing.T, stackUUID string, userID string, email string, isPublic bool) *http.Request {
	t.Helper()

	values := url.Values{}
	if isPublic {
		values.Set("is_public", "on")
	}

	req := httptest.NewRequest(http.MethodPost, "/app/stacks/detail/"+stackUUID+"/settings/is-public", strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("uuid", stackUUID)
	req = req.WithContext(commonauth.ContextWithSession(req.Context(), &commonauth.Session{
		UserID:  userID,
		Email:   email,
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

func assertStackPublicSettingSuccess(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if body := rr.Body.String(); !strings.Contains(body, "Stack changes saved") {
		t.Fatalf("body does not contain success toast: %s", body)
	}
}

func testPageTemplate(t *testing.T, pageTemplate string) *template.Template {
	t.Helper()

	tmpl, err := pageTemplateSet(testWebTemplate(t), pageTemplate)
	if err != nil {
		t.Fatalf("pageTemplateSet(%q) error = %v", pageTemplate, err)
	}

	return tmpl
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
