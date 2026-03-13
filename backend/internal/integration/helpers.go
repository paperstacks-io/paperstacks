package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// doRequest makes an HTTP request and returns the response.
// The caller is responsible for closing resp.Body.
func doRequest(t *testing.T, method, endpoint string, body io.Reader, headers map[string]string) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), method, endpoint, body)
	if err != nil {
		t.Fatalf("failed to prepare request: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	return resp
}

// doGetRequest makes a GET request and returns the response.
func doGetRequest(t *testing.T, endpoint string) *http.Response {
	t.Helper()

	headers := map[string]string{"Accept": "application/json"}
	return doRequest(t, http.MethodGet, endpoint, nil, headers)
}

// doPostRequest makes a POST request with JSON body and returns the response.
func doPostRequest(t *testing.T, endpoint string, body interface{}) *http.Response {
	t.Helper()

	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	headers := map[string]string{"Content-Type": "application/json"}
	return doRequest(t, http.MethodPost, endpoint, bytes.NewBuffer(jsonBody), headers)
}

// doPutRequest makes a PUT request with JSON body and returns the response.
func doPutRequest(t *testing.T, endpoint string, body interface{}) *http.Response {
	t.Helper()

	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	headers := map[string]string{"Content-Type": "application/json"}
	return doRequest(t, http.MethodPut, endpoint, bytes.NewBuffer(jsonBody), headers)
}

// doDeleteRequest makes a DELETE request and returns the response.
func doDeleteRequest(t *testing.T, endpoint string) *http.Response {
	t.Helper()

	headers := map[string]string{"Accept": "application/json"}
	return doRequest(t, http.MethodDelete, endpoint, nil, headers)
}

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
