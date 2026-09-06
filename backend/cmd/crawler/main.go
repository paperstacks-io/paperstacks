// Command crawler imports Crossref public-data-export dump files
// (*.jsonl or *.jsonl.gz) from a directory, parsing every record and
// reporting import statistics.
//
// Usage:
//
//	go run ./cmd/crawler -dir /path/to/dump
//
// Persistence isn't wired up yet — this currently reports what would be
// imported (record and error counts) without writing anywhere.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/paperstacks.io/paperstacks/internal/importer"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// run parses flags, builds the importer, and runs it until completion
// or until ctx is cancelled (e.g. by Ctrl+C, via main's signal.NotifyContext).
func run(ctx context.Context) error {
	dir := flag.String("dir", "", "Directory containing *.jsonl[.gz] dump files (required)")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	if *dir == "" {
		return fmt.Errorf("flag -dir is required")
	}

	level := slog.LevelInfo
	if *debug {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	imp := importer.New(importer.Config{
		Dir: *dir,
	}, log)

	return imp.Run(ctx)
}
