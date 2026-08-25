// testintegration contains all integration tests against the http API
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	"github.com/paperstacks.io/paperstacks/internal/paper/application"
	paperHttp "github.com/paperstacks.io/paperstacks/internal/paper/http"
	"github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
	"github.com/paperstacks.io/paperstacks/internal/server"
	stackApplication "github.com/paperstacks.io/paperstacks/internal/stack/application"
	stackMemory "github.com/paperstacks.io/paperstacks/internal/stack/repository/memory"
	userApplication "github.com/paperstacks.io/paperstacks/internal/user/application"
	userHttp "github.com/paperstacks.io/paperstacks/internal/user/http"
	userMemory "github.com/paperstacks.io/paperstacks/internal/user/repository/memory"
)

const clientTimeout = 10 * time.Second

type testApplication struct {
	baseURL string
	client  *http.Client
}

type noopSessionService struct{}

func (noopSessionService) ResolveSession(context.Context, string) (*commonauth.Session, error) {
	return nil, nil
}

func (noopSessionService) LogoutSession(context.Context, string) error {
	return nil
}

func startApplication(t *testing.T) testApplication {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	root := http.NewServeMux()
	api := http.NewServeMux()
	paperService := application.NewPaperService(memory.NewRepository())
	userRepo := userMemory.NewRepository()
	stackService := stackApplication.NewStackService(stackMemory.NewRepository(), paperService)
	userService := userApplication.NewUserService(userRepo)
	sessionService := noopSessionService{}
	server.AddRoute(root, context.Background(), logger, sessionService)
	paperHttp.AddPaperRoute(api, logger, paperService, sessionService)
	userHttp.AddUserRoute(api, logger, userService, stackService, sessionService)
	root.Handle("/api/", http.StripPrefix("/api", api))

	testServer := httptest.NewServer(root)
	t.Cleanup(testServer.Close)

	client := testServer.Client()
	client.Timeout = clientTimeout

	return testApplication{
		baseURL: testServer.URL,
		client:  client,
	}
}

// doRequest makes an HTTP request and returns the response.
// The caller is responsible for closing resp.Body.
func (a testApplication) doRequest(t *testing.T, method, endpoint string, body io.Reader, headers map[string]string) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), method, endpoint, body)
	if err != nil {
		t.Fatalf("failed to prepare request: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	return resp
}

// doGetRequest makes a GET request and returns the response.
func (a testApplication) doGetRequest(t *testing.T, endpoint string) *http.Response {
	t.Helper()

	headers := map[string]string{"Accept": "application/json"}
	return a.doRequest(t, http.MethodGet, endpoint, nil, headers)
}

// doPostRequest makes a POST request with JSON body and returns the response.
func (a testApplication) doPostRequest(t *testing.T, endpoint string, body any) *http.Response {
	t.Helper()

	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	headers := map[string]string{"Content-Type": "application/json"}
	return a.doRequest(t, http.MethodPost, endpoint, bytes.NewBuffer(jsonBody), headers)
}

// doPutRequest makes a PUT request with JSON body and returns the response.
func (a testApplication) doPutRequest(t *testing.T, endpoint string, body any) *http.Response {
	t.Helper()

	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	headers := map[string]string{"Content-Type": "application/json"}
	return a.doRequest(t, http.MethodPut, endpoint, bytes.NewBuffer(jsonBody), headers)
}

// doDeleteRequest makes a DELETE request and returns the response.
func (a testApplication) doDeleteRequest(t *testing.T, endpoint string) *http.Response {
	t.Helper()

	headers := map[string]string{"Accept": "application/json"}
	return a.doRequest(t, http.MethodDelete, endpoint, nil, headers)
}
