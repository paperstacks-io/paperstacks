package testintegration

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestHealthz(t *testing.T) {
	client := &http.Client{Timeout: 10 * time.Second}

	endpoint := testAPIPath() + "/healthz"
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status code %d, got %d", http.StatusNoContent, resp.StatusCode)
	}

	if string(body) != "" {
		t.Errorf("expected empty body '', got '%s'", string(body))
	}
}
