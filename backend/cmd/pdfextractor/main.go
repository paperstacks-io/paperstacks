// Command pdfextractor scans object storage for PDF files without a matching
// XML file, sends each missing PDF to GROBID to extract fulltext TEI-XML, and
// stores the resulting XML back in object storage.
//
// Usage:
//
//	go run ./cmd/pdfextractor [-grobid-url http://localhost:8070] [-prefix pdf/] [-debug]
//
// The GROBID URL defaults to GROBID_URL, falling back to http://localhost:8070.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/joho/godotenv/autoload"
	"github.com/paperstacks.io/paperstacks/internal/common/config"
	"github.com/paperstacks.io/paperstacks/internal/common/objectstorage"
	paperapp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	"github.com/paperstacks.io/paperstacks/internal/pdfextractor"
)

const storagePrefix = "pdf/"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg := config.New()
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	level := slog.LevelInfo
	if *debug {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	s3Store, err := objectstorage.NewS3Store(cfg.ObjectStorage, log)
	if err != nil {
		return fmt.Errorf("init object storage: %w", err)
	}
	if err := s3Store.Check(ctx); err != nil {
		return err
	}

	grobidClient := pdfextractor.NewGrobidClient(cfg.GrobidURL, nil)
	if err := grobidClient.Check(ctx); err != nil {
		return err
	}

	paperService := paperapp.NewPaperService(nil, s3Store)
	paperService.SetXMLExtractor(grobidClient)

	return paperService.SyncPDFXMLs(ctx, storagePrefix, log)
}
