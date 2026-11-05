package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Client provides an HTTP client to communicate with the PDF extractor microservice.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Client configured with the extractor service baseURL.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 60 * time.Second}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

// ExtractMetadata sends the PDF to the extractor service and returns the extracted metadata.
func (c *Client) ExtractMetadata(ctx context.Context, file io.Reader, filename string, headPages int) (MetadataResponse, error) {
	req, err := c.newMultipartRequest(ctx, "/api/extract/metadata", file, filename, headPages)
	if err != nil {
		return MetadataResponse{}, fmt.Errorf("create extract metadata request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return MetadataResponse{}, fmt.Errorf("do extract metadata request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return MetadataResponse{}, fmt.Errorf("extractor service returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var metadataResp MetadataResponse
	if err := json.NewDecoder(resp.Body).Decode(&metadataResp); err != nil {
		return MetadataResponse{}, fmt.Errorf("decode metadata response: %w", err)
	}

	return metadataResp, nil
}

// ExtractText sends the PDF to the extractor service and returns the extracted text.
func (c *Client) ExtractText(ctx context.Context, file io.Reader, filename string, headPages int) (TextResponse, error) {
	req, err := c.newMultipartRequest(ctx, "/api/extract/text", file, filename, headPages)
	if err != nil {
		return TextResponse{}, fmt.Errorf("create extract text request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return TextResponse{}, fmt.Errorf("do extract text request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return TextResponse{}, fmt.Errorf("extractor service returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var textResp TextResponse
	if err := json.NewDecoder(resp.Body).Decode(&textResp); err != nil {
		return TextResponse{}, fmt.Errorf("decode text response: %w", err)
	}

	return textResp, nil
}

func (c *Client) newMultipartRequest(ctx context.Context, path string, file io.Reader, filename string, headPages int) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(FormFieldFile, filename)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("copy file data: %w", err)
	}

	if headPages > 0 {
		if err := writer.WriteField(FormFieldHeadPages, strconv.Itoa(headPages)); err != nil {
			return nil, fmt.Errorf("write %s field: %w", FormFieldHeadPages, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}
