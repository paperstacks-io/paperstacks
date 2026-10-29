package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/common/build"
	extractorApp "github.com/paperstacks.io/paperstacks/internal/pdfextractor/application"
	extractorHttp "github.com/paperstacks.io/paperstacks/internal/pdfextractor/http"
)

const (
	host = "0.0.0.0"
	port = "8081"
)

func run(_ context.Context) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	if err := checkDependencies(); err != nil {
		slog.Error("Dependency check failed", slog.String("error", err.Error()))
		return err
	}

	extractorService := extractorApp.NewService()

	rootMux := http.NewServeMux()
	apiMux := http.NewServeMux()

	extractorHttp.AddExtractorRoute(apiMux, logger, extractorService)
	extractorHttp.AddHealthzRoute(rootMux)

	rootMux.Handle("/api/", http.StripPrefix("/api", apiMux))

	httpServer := &http.Server{
		Addr:         net.JoinHostPort(host, port),
		Handler:      rootMux,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		slog.Info("Extractor service is shutting down...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		httpServer.SetKeepAlivesEnabled(false)
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("Could not gracefully shutdown extractor server", slog.String("error", err.Error()))
		}
		close(done)
	}()

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "\tStarting PDF Extractor service:")
	fmt.Fprintln(tw, "\tAPI:      \thttp://"+host+":"+port+"/api/extract/")
	fmt.Fprintln(tw, "\tHealthz:  \thttp://"+host+":"+port+"/healthz")
	fmt.Fprintln(tw)
	_ = tw.Flush()

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Could not listen", slog.String("port", port), slog.String("error", err.Error()))
		return err
	}

	<-done
	slog.Info("Extractor service stopped")

	return nil
}

const bannerLogo = `
  ____  ____  _____   _____      _                  _             
 |  _ \|  _ \|  ___| | ____|_  _| |_ _ __ __ _  ___| |_ ___  _ __ 
 | |_) | | | | |_    |  _| \ \/ / __| '__/ _` + "`" + ` |/ __| __/ _ \| '__|
 |  __/| |_| |  _|   | |___ >  <| |_| | | (_| | (__| || (_) | |   
 |_|   |____/|_|     |_____/_/\_\\__|_|  \__,_|\___|\__\___/|_|   
`

func banner(w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	fmt.Fprint(tw, bannerLogo)
	fmt.Fprintln(tw)
	fmt.Fprintln(tw, "  Git hash:\t"+build.GitHash)
	fmt.Fprintln(tw, "  Listen:\t"+net.JoinHostPort(host, port))
	fmt.Fprintln(tw)

	_ = tw.Flush()
}

func main() {
	banner(os.Stdout)

	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func checkDependencies() error {
	// Add your dependency checks here
	return nil
}
