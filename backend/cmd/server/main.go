package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/paperstacks.io/paperstacks/internal/server"
)

func main() {
	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.SetLogLoggerLevel(slog.LevelDebug)

	handle := server.AddRoute(http.NewServeMux(), context.Background())
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: handle,
	}

	slog.Info("Starting server:", slog.String("host", host), slog.String("port", port))
	err := httpServer.ListenAndServe()
	if err != nil {
		slog.Error(err.Error())
	}
}
