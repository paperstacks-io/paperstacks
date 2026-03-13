// Package main runs the paperstacks HTTP server binary.
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

	doiApp "github.com/paperstacks.io/paperstacks/internal/doi/application"
	doiHttp "github.com/paperstacks.io/paperstacks/internal/doi/http"
	paperApp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	phttp "github.com/paperstacks.io/paperstacks/internal/paper/http"
	"github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
	"github.com/paperstacks.io/paperstacks/internal/server"
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

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	paperService := paperApp.NewPaperService(memory.NewRepository())
	doiService := doiApp.NewDOIService(nil)

	handle := http.NewServeMux()
	server.AddRoute(
		handle,
		ctx,
		logger,
	)
	phttp.AddPaperRoute(
		handle,
		logger,
		paperService,
	)
	doiHttp.AddDOIRoute(handle, logger, doiService)

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
			slog.Error("Could not gracefully shutdown the server", slog.String("error", err.Error()))
		}
		close(done)
	}()

	slog.Info("Starting server:", slog.String("host", host), slog.String("port", port))
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
