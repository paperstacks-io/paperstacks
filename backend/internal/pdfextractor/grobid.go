// Package pdfextractor provides a client for the GROBID service, which extracts
// structured TEI-XML from scientific PDFs, and a parser that maps that TEI-XML
// to the paper domain model.
package pdfextractor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

const (
	defaultTimeout      = 5 * time.Minute
	defaultMaxRetries   = 3
	defaultRetryDelay   = 2 * time.Second
	isAlivePath         = "/api/isalive"
	fulltextProcessPath = "/api/processFulltextDocument"
	formFieldInput      = "input"
	defaultFileName     = "document.pdf"
	pdfSignature        = "%PDF-"
)

var (
	ErrNilPDF          = errors.New("pdf reader is nil")
	ErrInvalidPDF      = errors.New("invalid or unreadable PDF: missing %PDF- header")
	ErrEmptyExtraction = errors.New("grobid returned empty XML content")
	ErrServiceBusy     = errors.New("grobid service is busy (HTTP 503)")
)

// GrobidClient communicates with a GROBID HTTP service to extract
// structured TEI-XML from scientific PDFs.
type GrobidClient struct {
	baseURL    string
	client     *http.Client
	maxRetries int
	retryDelay time.Duration
}

// NewGrobidClient constructs a GrobidClient for the given base URL
// (e.g. "http://localhost:8070"). If client is nil, a default client
// with a 5-minute timeout is used.
func NewGrobidClient(baseURL string, client *http.Client) *GrobidClient {
	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}
	return &GrobidClient{
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		client:     client,
		maxRetries: defaultMaxRetries,
		retryDelay: defaultRetryDelay,
	}
}

// Check verifies that the GROBID service is reachable and alive.
func (c *GrobidClient) Check(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+isAlivePath, nil)
	if err != nil {
		return fmt.Errorf("create grobid health request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("check grobid service at %q: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("grobid health check returned status %d", resp.StatusCode)
	}

	return nil
}

// ExtractFulltext validates that pdf begins with a valid %PDF- header, sends
// it to GROBID's /api/processFulltextDocument endpoint (retrying on HTTP 503
// when GROBID's worker pool is busy), and returns the resulting TEI-XML bytes.
func (c *GrobidClient) ExtractFulltext(ctx context.Context, pdf io.Reader) ([]byte, error) {
	if pdf == nil {
		return nil, ErrNilPDF
	}

	header := make([]byte, len(pdfSignature))
	if _, err := io.ReadFull(pdf, header); err != nil || string(header) != pdfSignature {
		return nil, ErrInvalidPDF
	}

	pdfData, err := io.ReadAll(io.MultiReader(bytes.NewReader(header), pdf))
	if err != nil {
		return nil, fmt.Errorf("read pdf stream: %w", err)
	}

	attempts := c.maxRetries + 1
	if attempts < 1 {
		attempts = 1
	}

	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			timer := time.NewTimer(c.retryDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
		}

		xmlData, err := c.sendFulltextRequest(ctx, pdfData)
		if !errors.Is(err, ErrServiceBusy) {
			return xmlData, err
		}
	}

	return nil, ErrServiceBusy
}

func (c *GrobidClient) sendFulltextRequest(ctx context.Context, pdf []byte) ([]byte, error) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile(formFieldInput, defaultFileName)
	if err != nil {
		return nil, fmt.Errorf("create multipart form file: %w", err)
	}
	if _, err := part.Write(pdf); err != nil {
		return nil, fmt.Errorf("write pdf to multipart form: %w", err)
	}
	if err := mw.Close(); err != nil {
		return nil, fmt.Errorf("close multipart form: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+fulltextProcessPath, &body)
	if err != nil {
		return nil, fmt.Errorf("create grobid request: %w", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/xml")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request grobid fulltext extraction: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// Handled below.
	case http.StatusNoContent:
		return nil, ErrEmptyExtraction
	case http.StatusServiceUnavailable:
		return nil, ErrServiceBusy
	default:
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		msg := strings.TrimSpace(string(snippet))
		if msg != "" {
			return nil, fmt.Errorf("grobid returned status %d: %s", resp.StatusCode, msg)
		}
		return nil, fmt.Errorf("grobid returned status %d", resp.StatusCode)
	}

	xmlData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read grobid response: %w", err)
	}
	if len(bytes.TrimSpace(xmlData)) == 0 {
		return nil, ErrEmptyExtraction
	}

	return xmlData, nil
}
