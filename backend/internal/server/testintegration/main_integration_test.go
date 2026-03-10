package testintegration

import (
	"context"
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

var (
	testHost = "localhost"
	testPort = "9999"
)

func testAPIPath() string {
	return "http://" + testHost + ":" + testPort
}

func startService() bool {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	paperService := paper.NewService(paper.NewMemoryRepo())
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
	if !startService() {
		os.Exit(1)
	}
	os.Exit(m.Run())
}
