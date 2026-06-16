package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	"github.com/paperstacks.io/paperstacks/internal/document/application"
	documentHttp "github.com/paperstacks.io/paperstacks/internal/document/http"
	documentMemory "github.com/paperstacks.io/paperstacks/internal/document/repository/memory"
	paperApplication "github.com/paperstacks.io/paperstacks/internal/paper/application"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	paperMemory "github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
	userApplication "github.com/paperstacks.io/paperstacks/internal/user/application"
	userMemory "github.com/paperstacks.io/paperstacks/internal/user/repository/memory"
)

type mockSessionService struct{}

func (m mockSessionService) ResolveSession(ctx context.Context, token string) (*commonauth.Session, error) {
	return &commonauth.Session{
		Token:   token,
		UserID:  "user-1",
		Email:   "user@example.com",
		IsValid: true,
	}, nil
}

func (m mockSessionService) LogoutSession(ctx context.Context, token string) error {
	return nil
}

func setupTestRouter(t *testing.T) (http.Handler, *application.DocumentService) {
	docRepo := documentMemory.NewRepository()
	docStorage := documentMemory.NewStorage()
	docService := application.NewDocumentService(docRepo, docStorage)

	userRepo := userMemory.NewRepository()
	userService := userApplication.NewUserService(userRepo, "", nil)

	paperRepo := paperMemory.NewRepository()
	_, err := paperRepo.Save(context.Background(), paperDomain.Paper{
		UUID:  "paper-uuid-123",
		DOI:   "10.1145/1234567.1234568",
		Title: "Mock Paper",
	})
	if err != nil {
		t.Fatalf("failed to seed mock paper: %v", err)
	}
	paperService := paperApplication.NewPaperService(paperRepo)

	_, err = userService.CreateIfNotExist(context.Background(), "user-1", "user@example.com")
	if err != nil {
		t.Fatalf("failed to create mock user: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	mux := http.NewServeMux()
	documentHttp.UploadDocumentRoute(
		mux,
		logger,
		docService,
		userService,
		paperService,
		mockSessionService{},
	)
	return mux, docService
}

func TestUploadDocument(t *testing.T) {
	t.Parallel()

	mux, _ := setupTestRouter(t)

	pdfContent := []byte("%PDF-1.4\n%...\nEOF")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	err := writer.WriteField("paper_uuid", "paper-uuid-123")
	if err != nil {
		t.Fatalf("failed to write field: %v", err)
	}

	part, err := writer.CreateFormFile("file", "test.pdf")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	_, err = part.Write(pdfContent)
	if err != nil {
		t.Fatalf("failed to write file content: %v", err)
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/document", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: "hanko", Value: "valid-session-token"})

	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d. Body: %s", res.Code, http.StatusCreated, res.Body.String())
	}

	var response documentHttp.DocumentResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.PaperUUID != "paper-uuid-123" {
		t.Errorf("PaperUUID = %q, want %q", response.PaperUUID, "paper-uuid-123")
	}

	if response.FileName != "test.pdf" {
		t.Errorf("FileName = %q, want %q", response.FileName, "test.pdf")
	}

	if response.Size != int64(len(pdfContent)) {
		t.Errorf("Size = %d, want %d", response.Size, len(pdfContent))
	}

	if response.StorageURI == "" {
		t.Errorf("expected StorageURI to be populated, got empty string")
	}
}

func TestUploadDocumentValidationRequiredFields(t *testing.T) {
	t.Parallel()

	mux, _ := setupTestRouter(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test.pdf")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	_, err = part.Write([]byte("%PDF-1.4"))
	if err != nil {
		t.Fatalf("failed to write content: %v", err)
	}

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/document", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: "hanko", Value: "valid-session-token"})

	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d for missing paper_uuid", res.Code, http.StatusBadRequest)
	}

	if !strings.Contains(res.Body.String(), "paper_uuid is required") {
		t.Errorf("unexpected body error message: %q", res.Body.String())
	}
}

func TestUploadDocumentValidationInvalidFileType(t *testing.T) {
	t.Parallel()

	mux, _ := setupTestRouter(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("paper_uuid", "paper-uuid-123")
	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	_, err = part.Write([]byte("hello world"))
	if err != nil {
		t.Fatalf("failed to write content: %v", err)
	}

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/document", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: "hanko", Value: "valid-session-token"})

	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d for invalid file type", res.Code, http.StatusBadRequest)
	}
}

func TestUploadDocumentMaxBytesLimit(t *testing.T) {
	t.Parallel()

	mux, _ := setupTestRouter(t)

	largeData := make([]byte, 16*1024*1024)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("paper_uuid", "paper-uuid-123")
	part, err := writer.CreateFormFile("file", "large.pdf")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	_, err = part.Write(largeData)
	if err != nil {
		t.Fatalf("failed to write content: %v", err)
	}

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/document", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: "hanko", Value: "valid-session-token"})

	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d for request exceeding 15MB", res.Code, http.StatusBadRequest)
	}
}
