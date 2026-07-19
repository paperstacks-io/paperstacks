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
	paperRepo := paperMemory.NewRepository()
	_, err := paperRepo.Save(context.Background(), paperDomain.Paper{
		UUID:  "960ae542-c8ee-4454-9ad3-536ffbbacde6",
		DOI:   "10.1145/1234567.1234568",
		Title: "Mock Paper",
	})
	if err != nil {
		t.Fatalf("failed to seed mock paper: %v", err)
	}
	paperService := paperApplication.NewPaperService(paperRepo)

	docRepo := documentMemory.NewRepository()
	docStorage := documentMemory.NewStorage()
	docService := application.NewDocumentService(docRepo, docStorage, paperService)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	mux := http.NewServeMux()
	documentHttp.UploadDocumentRoute(
		mux,
		logger,
		docService,
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

	err := writer.WriteField("paper_uuid", "960ae542-c8ee-4454-9ad3-536ffbbacde6")
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

	if response.PaperUUID != "960ae542-c8ee-4454-9ad3-536ffbbacde6" {
		t.Errorf("PaperUUID = %q, want %q", response.PaperUUID, "960ae542-c8ee-4454-9ad3-536ffbbacde6")
	}

	if response.FileName != "test.pdf" {
		t.Errorf("FileName = %q, want %q", response.FileName, "test.pdf")
	}

	if response.Size != int64(len(pdfContent)) {
		t.Errorf("Size = %d, want %d", response.Size, len(pdfContent))
	}

	if !strings.HasPrefix(response.Key, "paper/960ae542-c8ee-4454-9ad3-536ffbbacde6/") || !strings.HasSuffix(response.Key, ".pdf") {
		t.Errorf("Key = %q, expected format 'paper/960ae542-c8ee-4454-9ad3-536ffbbacde6/{uuid}.pdf'", response.Key)
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

	writer.WriteField("paper_uuid", "960ae542-c8ee-4454-9ad3-536ffbbacde6")
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

	const fileSizeExceedingLimit = 16 * 1024 * 1024
	largeData := make([]byte, fileSizeExceedingLimit)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("paper_uuid", "960ae542-c8ee-4454-9ad3-536ffbbacde6")
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
		t.Errorf("status = %d, want %d for request exceeding max file size limit (10MB)", res.Code, http.StatusBadRequest)
	}
}
