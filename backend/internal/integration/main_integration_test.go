// testintegration contains all integration tests against the http API
package integration

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
	"github.com/paperstacks.io/paperstacks/internal/paper/application"
	paperHttp "github.com/paperstacks.io/paperstacks/internal/paper/http"
	"github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
	"github.com/paperstacks.io/paperstacks/internal/server"
)

const (
	testHost      = "localhost"
	testPort      = "9999"
	testAPIPath   = "http://" + testHost + ":" + testPort
	clientTimeout = 10 * time.Second
)

var client *http.Client

func startApplication() bool {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	root := http.NewServeMux()
	api := http.NewServeMux()
	server.AddRoute(root, context.Background(), logger)

	paperService := application.NewPaperService(memory.NewRepository())
	paperHttp.AddPaperRoute(api, logger, paperService)
	root.Handle("/api/", http.StripPrefix("/api", api))

	httpServer := &http.Server{
		Addr:         net.JoinHostPort(testHost, testPort),
		Handler:      root,
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

	if !startApplication() {
		os.Exit(1)
	}

	client = &http.Client{Timeout: clientTimeout}

	os.Exit(m.Run())
}
