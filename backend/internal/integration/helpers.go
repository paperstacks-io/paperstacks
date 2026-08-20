package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// assertStatusCode checks if the response status code matches the expected value.
func assertStatusCode(t *testing.T, resp *http.Response, expected int) {
	t.Helper()

	if resp.StatusCode != expected {
		t.Errorf("expected status code %d, got %d", expected, resp.StatusCode)
	}
}

// assertStatusCodeFatal checks if the response status code matches the expected value,
// and fails the test if it doesn't.
func assertStatusCodeFatal(t *testing.T, resp *http.Response, expected int) {
	t.Helper()

	if resp.StatusCode != expected {
		t.Fatalf("expected status code %d, got %d", expected, resp.StatusCode)
	}
}

// assertBody checks if the response body matches the expected string.
func assertBody(t *testing.T, resp *http.Response, expected string) {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if string(body) != expected {
		t.Errorf("expected body '%s', got '%s'", expected, string(body))
	}
}

// decodeJSON unmarshals the response body into the given value.
func decodeJSON(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()

	err := json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
}
