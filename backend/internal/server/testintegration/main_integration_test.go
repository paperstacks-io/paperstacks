// testintegration contains all integration tests against the http API
package testintegration

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/common/tests"
	"github.com/paperstacks.io/paperstacks/internal/paper"
	"github.com/paperstacks.io/paperstacks/internal/server"
)

const (
	testHost      = "localhost"
	testPort      = "9999"
	clientTimeout = 10 * time.Second
)

var client *http.Client

func testAPIPath() string {
	return "http://" + testHost + ":" + testPort
}

func startService() bool {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	paperService, err := paper.NewService(
		paper.MemoryRepository(paper.NewMemoryRepo()),
	)
	if err != nil {
		return false
	}
	handle := server.AddRoute(http.NewServeMux(), context.Background(), logger, nil, paperService)
	httpServer := &http.Server{
		Addr:         net.JoinHostPort(testHost, testPort),
		Handler:      handle,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	go httpServer.ListenAndServe()

	ok := tests.WaitForPort(testHost + ":" + testPort)

	if !ok {
		log.Println("Timed out waiting for trainings HTTP to come up")
	}
	return ok
}

func TestMain(m *testing.M) {
	if !tests.IsIntegrationTest() {
		fmt.Println("skipping integration tests: set INTEGRATION environment variable")
		return
	}

	if !startService() {
		os.Exit(1)
	}

	client = &http.Client{Timeout: clientTimeout}

	os.Exit(m.Run())
}
