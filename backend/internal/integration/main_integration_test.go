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
	"sync"
	"testing"
	"time"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	"github.com/paperstacks.io/paperstacks/internal/common/tests"
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

const (
	testHost      = "localhost"
	testPort      = "9999"
	testAPIPath   = "http://" + testHost + ":" + testPort
	clientTimeout = 10 * time.Second
)

var client *http.Client
var testRepo *memory.Repository
var integrationTestMu sync.Mutex

type noopSessionService struct{}

func (noopSessionService) ResolveSession(context.Context, string) (*commonauth.Session, error) {
	return nil, nil
}

func (noopSessionService) LogoutSession(context.Context, string) error {
	return nil
}

func startApplication() bool {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	root := http.NewServeMux()
	api := http.NewServeMux()
	testRepo = memory.NewRepository()
	paperService := application.NewPaperService(testRepo)
	userService := userApplication.NewUserService(userMemory.NewRepository(), "", nil)
	stackService := stackApplication.NewStackService(stackMemory.NewRepository(), userService, paperService)
	sessionService := noopSessionService{}
	server.AddRoute(root, context.Background(), logger, sessionService)
	paperHttp.AddPaperRoute(api, logger, paperService, sessionService)
	userHttp.AddUserRoute(api, logger, userService, stackService, sessionService)
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
