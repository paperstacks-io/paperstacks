package testintegration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/server"
)

func TestPapersGetAll(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	endpoint := testAPIPath() + "/papers/"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatalf("failed to prepare request: %v", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusNoContent, resp.StatusCode)
	}

	var papers []server.PaperResponse
	err = json.NewDecoder(resp.Body).Decode(&papers)
	if err != nil {
		t.Fatalf("failed to decode response body as []PaperResponse: %v", err)
	}

	if len(papers) != 3 {
		t.Errorf("expected length of PaperResponse %d, got %d", 3, len(papers))
	}
}

func TestPapersGetSingle(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	endpoint := testAPIPath() + "/papers/1"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatalf("failed to prepare request: %v", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusNoContent, resp.StatusCode)
	}

	var paper server.PaperResponse
	err = json.NewDecoder(resp.Body).Decode(&paper)
	if err != nil {
		t.Fatalf("failed to decode response body as PaperResponse: %v", err)
	}

	if paper.DOI != "1" {
		t.Fatalf("expected paper to have DOI %s, got %s", "3", paper.DOI)
	}
}

func TestPapersGetSingleUnkown(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	endpoint := testAPIPath() + "/papers/doesntexist"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatalf("failed to prepare request: %v", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d", http.StatusNoContent, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if string(body) != "paper not found\n" {
		t.Errorf("expected body '%s', got '%s'", "paper not found", string(body))
	}
}
