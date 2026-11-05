package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/pdfextractor/domain"
)

type mockExtractor struct {
	metadataFn func(ctx context.Context, r io.Reader, opts domain.Options) (domain.ExtractedMetadata, error)
	textFn     func(ctx context.Context, r io.Reader, opts domain.Options) (domain.ExtractedText, error)
}

func (m *mockExtractor) ExtractMetadata(ctx context.Context, r io.Reader, opts domain.Options) (domain.ExtractedMetadata, error) {
	if m.metadataFn != nil {
		return m.metadataFn(ctx, r, opts)
	}
	return domain.ExtractedMetadata{}, nil
}

func (m *mockExtractor) ExtractText(ctx context.Context, r io.Reader, opts domain.Options) (domain.ExtractedText, error) {
	if m.textFn != nil {
		return m.textFn(ctx, r, opts)
	}
	return domain.ExtractedText{}, nil
}

func createMultipartRequest(t *testing.T, target string, fieldName, filename string, content []byte) *http.Request {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("failed to write part content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, target, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestHandleExtractMetadata(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mock := &mockExtractor{
		metadataFn: func(_ context.Context, _ io.Reader, _ domain.Options) (domain.ExtractedMetadata, error) {
			return domain.ExtractedMetadata{
				Title:    "Test Paper Title",
				DOI:      "10.1145/1234567.1234568",
				Abstract: "This is a test abstract.",
				Keywords: []string{"Testing", "Go"},
			}, nil
		},
	}

	handler := handleExtractMetadata(logger, mock)

	t.Run("successful metadata extraction", func(t *testing.T) {
		req := createMultipartRequest(t, "/api/extract/metadata", "file", "test.pdf", []byte("%PDF-1.4 test"))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp MetadataResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.DOI != "10.1145/1234567.1234568" {
			t.Errorf("expected DOI 10.1145/1234567.1234568, got %q", resp.DOI)
		}
		if resp.Title != "Test Paper Title" {
			t.Errorf("expected Title 'Test Paper Title', got %q", resp.Title)
		}
	})

	t.Run("missing file returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/extract/metadata", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})
}

func TestHandleExtractText(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mock := &mockExtractor{
		textFn: func(_ context.Context, _ io.Reader, _ domain.Options) (domain.ExtractedText, error) {
			return domain.ExtractedText{
				FullText:   "Full paper text content for embeddings...",
				TotalPages: 5,
				IsScanned:  false,
			}, nil
		},
	}

	handler := handleExtractText(logger, mock)

	t.Run("successful text extraction", func(t *testing.T) {
		req := createMultipartRequest(t, "/api/extract/text", "file", "paper.pdf", []byte("%PDF-1.4 paper content"))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp TextResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.FullText != "Full paper text content for embeddings..." {
			t.Errorf("unexpected text content: %q", resp.FullText)
		}
	})
}

func TestHandleHealthz(t *testing.T) {
	handler := handleHealthz()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "OK\n" {
		t.Errorf("expected 'OK\\n', got %q", rec.Body.String())
	}
}

