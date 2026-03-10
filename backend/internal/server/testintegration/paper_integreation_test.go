package testintegration

import (
	"bytes"
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
		t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var papers []server.PaperResponse
	err = json.NewDecoder(resp.Body).Decode(&papers)
	if err != nil {
		t.Fatalf("failed to decode response body as []PaperResponse: %v", err)
	}

	if len(papers) < 3 {
		t.Errorf("expected at least %d papers, got %d", 3, len(papers))
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
		t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var paper server.PaperResponse
	err = json.NewDecoder(resp.Body).Decode(&paper)
	if err != nil {
		t.Fatalf("failed to decode response body as PaperResponse: %v", err)
	}

	if paper.DOI != "1" {
		t.Fatalf("expected paper to have DOI %s, got %s", "1", paper.DOI)
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
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if string(body) != "paper not found\n" {
		t.Errorf("expected body '%s', got '%s'", "paper not found\n", string(body))
	}
}

func TestPapersCreate(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	paperBody := server.CreatePaperRequest{
		DOI:   "4",
		Title: "Created Paper",
	}
	body, _ := json.Marshal(paperBody)

	endpoint := testAPIPath() + "/papers/"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to prepare request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	// Verify it was created
	endpoint = testAPIPath() + "/papers/4"
	req, _ = http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d after creation, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestPapersUpdate(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	paperBody := server.UpdatePaperRequest{
		Title: "Updated Paper Two",
	}
	body, _ := json.Marshal(paperBody)

	endpoint := testAPIPath() + "/papers/2"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, endpoint, bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to prepare request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Verify it was updated
	req, _ = http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()
	var paper server.PaperResponse
	json.NewDecoder(resp.Body).Decode(&paper)
	if paper.Title != "Updated Paper Two" {
		t.Errorf("expected title %s, got %s", "Updated Paper Two", paper.Title)
	}
}

func TestPapersDelete(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	doiToDelete := "to-be-deleted"
	paperBody := server.CreatePaperRequest{
		DOI:   doiToDelete,
		Title: "To Be Deleted",
	}

	body, _ := json.Marshal(paperBody)
	createEndpoint := testAPIPath() + "/papers/"
	createReq, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, createEndpoint, bytes.NewBuffer(body))
	createReq.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(createReq)
	if err != nil {
		t.Fatalf("failed to create paper for deletion: %v", err)
	}
	resp.Body.Close()

	endpoint := testAPIPath() + "/papers/" + doiToDelete
	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, endpoint, nil)
	if err != nil {
		t.Fatalf("failed to prepare request: %v", err)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status code %d, got %d", http.StatusNoContent, resp.StatusCode)
	}

	req, _ = http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status code %d after deletion, got %d", http.StatusNotFound, resp.StatusCode)
	}
}
