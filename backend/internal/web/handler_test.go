package web

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/common/objectstorage"
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

type uploadPDFTestStore struct {
	objectstorage.Store
	input objectstorage.PutObjectInput
	body  []byte
	err   error
}

func (s *uploadPDFTestStore) Put(_ context.Context, input objectstorage.PutObjectInput) (objectstorage.ObjectInfo, error) {
	s.input = input
	var err error
	s.body, err = io.ReadAll(input.Body)
	if err != nil {
		return objectstorage.ObjectInfo{}, err
	}
	return objectstorage.ObjectInfo{}, s.err
}

func TestHandlePaperUploadPDF(t *testing.T) {
	const paperUUID = "11111111-1111-4111-8111-111111111111"
	const stackURL = "/app/stacks/detail/22222222-2222-4222-8222-222222222222"
	tests := []struct {
		name         string
		content      string
		field        string
		missing      bool
		noAuth       bool
		forbidden    bool
		notInStack   bool
		missingStack bool
		storeErr     error
		want         string
	}{
		{name: "success", field: "file", content: "%PDF-1.4\ncontent", want: ""},
		{name: "invalid PDF", field: "file", content: "not a PDF", want: "not a valid PDF"},
		{name: "missing file", field: "other", want: "Choose a PDF file"},
		{name: "missing paper", field: "file", missing: true, want: "paper could not be found"},
		{name: "unauthenticated", noAuth: true, want: "Please log in"},
		{name: "forbidden", forbidden: true, want: "not allowed"},
		{name: "paper not in stack", notInStack: true, want: "does not belong"},
		{name: "missing stack", missingStack: true, want: "stack could not be found"},
		{name: "storage error", field: "file", content: "%PDF-1.4\ncontent", storeErr: errors.New("storage unavailable"), want: "could not be uploaded"},
		{name: "file too large", field: "file", content: "%PDF-" + strings.Repeat("x", paperDomain.PDFMaxFileSize), want: "no larger than 10 MiB"},
		{name: "request too large", field: "file", content: strings.Repeat("x", paperDomain.PDFMaxFileSize+int(oneMiB)), want: "no larger than 10 MiB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := paperMemory.NewRepository()
			if !tt.missing {
				if _, err := repo.Save(t.Context(), paperDomain.Paper{UUID: paperUUID, Title: "Test paper"}); err != nil {
					t.Fatal(err)
				}
			}
			store := &uploadPDFTestStore{err: tt.storeErr}
			stackService := stackApp.NewStackService(stackMemory.NewRepository(nil), nil)
			stack := stackDomain.Stack{
				UUID:   strings.TrimPrefix(stackURL, "/app/stacks/detail/"),
				Name:   "Upload stack",
				Owner:  userDomain.User{ExternalID: "owner"},
				Papers: []paperDomain.Paper{{UUID: paperUUID}},
			}
			if tt.notInStack {
				stack.Papers = nil
			}
			if !tt.missingStack {
				if err := stackService.Create(t.Context(), stack); err != nil {
					t.Fatal(err)
				}
			}
			handler := handlePaperUploadPDF(testLogger(), testWebTemplate(t), paperApp.NewPaperService(repo, store), stackService)
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			file, err := writer.CreateFormFile(tt.field, "paper.pdf")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := io.WriteString(file, tt.content); err != nil {
				t.Fatal(err)
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(http.MethodPost, stackURL+"/papers/"+paperUUID+"/upload-pdf", &body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.SetPathValue("stackUUID", stack.UUID)
			req.SetPathValue("paperUUID", paperUUID)
			if !tt.noAuth {
				req = req.WithContext(commonauth.ContextWithSession(req.Context(), &commonauth.Session{UserID: "owner", IsValid: true}))
			}
			if tt.forbidden {
				req = req.WithContext(commonauth.ContextWithSession(req.Context(), &commonauth.Session{UserID: "other", IsValid: true}))
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), tt.want) {
				t.Fatalf("response = %d %s, want toast containing %q", rr.Code, rr.Body.String(), tt.want)
			}
			paper, _ := repo.GetByUUID(t.Context(), paperUUID)
			wantLocation := ""
			if tt.name == "success" {
				wantLocation = stackURL
			}
			if got := rr.Header().Get("HX-Location"); got != wantLocation {
				t.Fatalf("HX-Location = %q, want %q", got, wantLocation)
			}
			if tt.name == "success" {
				if paper.PDFStorageKey != "pdf/"+paperUUID+".pdf" || string(store.body) != tt.content ||
					store.input.Metadata["filename"] != "paper.pdf" || store.input.Size != int64(len(tt.content)) ||
					store.input.ContentType != paperDomain.PDFContentType {
					t.Fatalf("PDF upload did not preserve content, metadata, or attachment: paper=%+v, input=%+v", paper, store.input)
				}
			} else if paper.PDFStorageKey != "" {
				t.Fatalf("failed upload attached PDF: %q", paper.PDFStorageKey)
			}
			if (tt.noAuth || tt.forbidden || tt.notInStack || tt.missingStack) && store.input.Key != "" {
				t.Fatal("unauthorized upload reached object storage")
			}
		})
	}
}

func TestPaperInfoUploadURLIncludesStackUUID(t *testing.T) {
	const stackUUID = "22222222-2222-4222-8222-222222222222"
	paper := paperDomain.Paper{UUID: "11111111-1111-4111-8111-111111111111", Title: "Test paper"}
	stackService := newTestStackService(paper)
	if err := stackService.Create(t.Context(), stackDomain.Stack{
		UUID: stackUUID, Name: "Upload stack", Papers: []paperDomain.Paper{paper},
	}); err != nil {
		t.Fatal(err)
	}
	repo := paperMemory.NewRepository()
	if _, err := repo.Save(t.Context(), paper); err != nil {
		t.Fatal(err)
	}
	tmpl := testWebTemplate(t)
	detailTmpl, err := pageTemplateSet(tmpl, "stacks/detail")
	if err != nil {
		t.Fatal(err)
	}
	for name, handler := range map[string]http.Handler{
		"detail page":   handleStacksDetailPage(testLogger(), detailTmpl, "", stackService),
		"paper partial": handleStackPaperInfo(testLogger(), tmpl, paperApp.NewPaperService(repo, nil)),
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/app/stacks/detail/"+stackUUID, nil)
			req.SetPathValue("uuid", stackUUID)
			req.SetPathValue("stackUUID", stackUUID)
			req.SetPathValue("paperUUID", paper.UUID)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			want := `hx-post="/app/stacks/detail/` + stackUUID + `/papers/` + paper.UUID + `/upload-pdf"`
			if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), want) {
				t.Fatalf("response = %d %s, want upload URL %s", rr.Code, rr.Body.String(), want)
			}
		})
	}
}

func newTestStackService(papers ...paperDomain.Paper) *stackApp.StackService {
	repo := paperMemory.NewRepository()
	for _, paper := range papers {
		_, _ = repo.Save(context.Background(), paper)
	}
	paperService := paperApp.NewPaperService(repo, nil)
	return stackApp.NewStackService(stackMemory.NewRepository(paperService), paperService)
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

func TestHandleStackPaperRemoveRemovesPaperAndNavigatesToDetail(t *testing.T) {
	t.Parallel()

	user := userDomain.NewUser("owner-paper-remove", "paper-remove@example.com")
	stack := stackDomain.NewStack("Paper Remove Stack", user)
	removedPaperUUID := "11111111-1111-4111-8111-111111111111"
	retainedPaperUUID := "22222222-2222-4222-8222-222222222222"
	stack.Papers = []paperDomain.Paper{
		{UUID: removedPaperUUID, Title: "Removed Paper"},
		{UUID: retainedPaperUUID, Title: "Retained Paper"},
	}
	stackService := newTestStackService(stack.Papers...)
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
	if got := rr.Header().Get("HX-Location"); got != "/app/stacks/detail/"+stack.UUID {
		t.Fatalf("HX-Location = %q, want detail page redirect", got)
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

func TestHandleSidebarStackCreateNavigatesToDetail(t *testing.T) {
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

	location := rr.Header().Get("HX-Location")
	if !strings.HasPrefix(location, "/app/stacks/detail/") {
		t.Fatalf("HX-Location = %q, want /app/stacks/detail/{uuid}", location)
	}

	uuid := strings.TrimPrefix(location, "/app/stacks/detail/")
	created, err := stackService.GetByUUID(req.Context(), uuid)
	if err != nil {
		t.Fatalf("created stack not found by HX-Location UUID %q: %v", uuid, err)
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
