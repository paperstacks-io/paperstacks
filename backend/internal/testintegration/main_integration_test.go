package testintegration

import (
	"context"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/common/tests"
	"github.com/paperstacks.io/paperstacks/internal/server"
)

func startService() bool {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	handle := server.AddRoute(http.NewServeMux(), context.Background(), logger, nil)
	httpServer := &http.Server{
		Addr:         net.JoinHostPort("localhost", "9000"),
		Handler:      handle,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	go httpServer.ListenAndServe()

	ok := tests.WaitForPort("localhost:9000")

	if !ok {
		log.Println("Timed out waiting for trainings HTTP to come up")
	}
	return ok
}

func TestMain(m *testing.M) {
	if !startService() {
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestHealthz(t *testing.T) {
	t.Log("TestHealthz...")
	client := &http.Client{Timeout: 10 * time.Second}

	endpoint := "http://localhost:9000/healthz"
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
