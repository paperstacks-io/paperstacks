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
	"time"

	"github.com/paperstacks.io/paperstacks/cmd/crawler/reader"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	dir := flag.String("dir", "", "Directory containing *.jsonl[.gz] dump files (required)")
	progressEvery := flag.Int("progress-every", 100_000, "Log a progress line every N records (0 disables)")
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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Info("crawler import starting", slog.String("dir", *dir))
	start := time.Now()

	var totalFiles, totalRecords, totalErrors int

	// importFile drains one dump file's records, tallying successes and
	// per-line parse errors. A non-nil error means either the file
	// couldn't be opened/decompressed, or the run was interrupted.
	importFile := func(path string) (records, errs int, err error) {
		ch, err := reader.ReadFile(path)
		if err != nil {
			return 0, 0, fmt.Errorf("open file: %w", err)
		}

		for rec := range ch {
			if rec.Err != nil {
				errs++
				log.Warn("parse error",
					slog.String("file", path),
					slog.Int("line", rec.LineNumber),
					slog.String("error", rec.Err.Error()),
				)
				continue
			}

			records++
			totalRecords++
			if *progressEvery > 0 && totalRecords%*progressEvery == 0 {
				log.Info("progress", slog.Int("total_records", totalRecords))
			}

			if err := ctx.Err(); err != nil {
				return records, errs, err
			}
		}

		return records, errs, nil
	}

	walkErr := reader.WalkDumpDir(*dir, func(path string) error {
		if err := ctx.Err(); err != nil {
			return err
		}

		fileRecords, fileErrors, err := importFile(path)
		totalFiles++
		totalErrors += fileErrors

		if err != nil {
			log.Error("file failed", slog.String("file", path), slog.String("error", err.Error()))
			return ctx.Err()
		}

		log.Info("file complete",
			slog.String("file", path),
			slog.Int("records", fileRecords),
			slog.Int("errors", fileErrors),
			slog.Int("total_records", totalRecords),
			slog.Duration("elapsed", time.Since(start)),
		)
		return nil
	})

	log.Info("import finished",
		slog.Int("files", totalFiles),
		slog.Int("records", totalRecords),
		slog.Int("errors", totalErrors),
		slog.Duration("elapsed", time.Since(start)),
	)

	return walkErr
}
