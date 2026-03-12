// Package main runs the paperstacks HTTP server_ binary.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/doi_"
	paperhttp "github.com/paperstacks.io/paperstacks/internal/paper/infrastructure/http"
	"github.com/paperstacks.io/paperstacks/internal/paper/infrastructure/persistence/memory"
	"github.com/paperstacks.io/paperstacks/internal/paper/service"
	"github.com/paperstacks.io/paperstacks/internal/paper_"
	"github.com/paperstacks.io/paperstacks/internal/server_"
)

func run(
	ctx context.Context,
	getenv func(string) string,
) error {
	host := getenv("HOST")
	if host == "" {
		host = "localhost"
	}
	port := getenv("PORT")
	if port == "" {
		port = "8080"
	}
	portTest := getenv("PORT")
	if portTest == "" {
		portTest = "8090"
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	doiService := doi_.NewService(nil)

	paperService, err := paper_.NewService(
		paper_.MemoryRepository(paper_.NewMemoryRepo()),
	)
	if err != nil {
		return err
	}

	paperApplication := service.NewApplication(memory.NewMemoryPaperRepository())
	handleTest := paperhttp.AddRouteTest(
		http.NewServeMux(),
		ctx,
		logger,
		paperApplication,
	)
	httpServerTest := &http.Server{
		Addr:         net.JoinHostPort("localhost", "8090"),
		Handler:      handleTest,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	handle := server_.AddRoute(
		http.NewServeMux(),
		ctx,
		logger,
		doiService,
		paperService,
	)
	httpServer := &http.Server{
		Addr:         net.JoinHostPort("localhost", "8080"),
		Handler:      handle,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		slog.Info("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		httpServer.SetKeepAlivesEnabled(false)
		if err := httpServer.Shutdown(ctx); err != nil {
			slog.Error("Could not gracefully shutdown the server_", slog.String("error", err.Error()))
		}
		close(done)
	}()

	slog.Info("Starting server_:", slog.String("host", host), slog.String("port", portTest))
	if err := httpServerTest.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Could not listen", slog.String("port", port), slog.String("error", err.Error()))
		return err
	}

	slog.Info("Starting server_:", slog.String("host", host), slog.String("port", port))
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Could not listen", slog.String("port", port), slog.String("error", err.Error()))
		return err
	}

	<-done
	slog.Info("Server stopped")

	return nil
}

func main() {
	if err := run(context.Background(), os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "%s:n", err)
		os.Exit(1)
	}
}
