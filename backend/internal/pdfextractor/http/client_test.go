package http

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/pdfextractor/domain"
)

func TestClient_ExtractMetadata(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mock := &mockExtractor{
		metadataFn: func(_ context.Context, _ io.Reader, opts domain.Options) (domain.ExtractedMetadata, error) {
			if opts.HeadPages != 3 {
				t.Errorf("expected HeadPages 3, got %d", opts.HeadPages)
			}
			return domain.ExtractedMetadata{
				Title: "Client Test Paper",
				DOI:   "10.1234/test.doi",
			}, nil
		},
	}

	apiMux := http.NewServeMux()
	AddExtractorRoute(apiMux, logger, mock)

	rootMux := http.NewServeMux()
	rootMux.Handle("/api/", http.StripPrefix("/api", apiMux))

	ts := httptest.NewServer(rootMux)
	defer ts.Close()

	client := NewClient(ts.URL, ts.Client())

	pdfData := strings.NewReader("%PDF-1.4 test dummy")
	resp, err := client.ExtractMetadata(context.Background(), pdfData, "paper.pdf", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Title != "Client Test Paper" {
		t.Errorf("expected Title 'Client Test Paper', got %q", resp.Title)
	}
	if resp.DOI != "10.1234/test.doi" {
		t.Errorf("expected DOI '10.1234/test.doi', got %q", resp.DOI)
	}
}

func TestClient_ExtractText(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mock := &mockExtractor{
		textFn: func(_ context.Context, _ io.Reader, opts domain.Options) (domain.ExtractedText, error) {
			return domain.ExtractedText{
				FullText:   "Full extracted text from client call",
				TotalPages: 10,
				IsScanned:  false,
			}, nil
		},
	}

	apiMux := http.NewServeMux()
	AddExtractorRoute(apiMux, logger, mock)

	rootMux := http.NewServeMux()
	rootMux.Handle("/api/", http.StripPrefix("/api", apiMux))

	ts := httptest.NewServer(rootMux)
	defer ts.Close()

	client := NewClient(ts.URL, ts.Client())

	pdfData := strings.NewReader("%PDF-1.4 text content")
	resp, err := client.ExtractText(context.Background(), pdfData, "paper.pdf", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.FullText != "Full extracted text from client call" {
		t.Errorf("unexpected full text: %q", resp.FullText)
	}
}
