// Command crossref-sync fetches works changed in Crossref since the last
// run from the Crossref REST API, one day window at a time, and passes
// them to a PaperWriter. Progress is kept per filter in
// <state-dir>/<filter>.jsonl, so each run resumes after the last completed
// day; the first run starts at -from. It is meant to be run from cron,
// e.g. update-date daily and index-date weekly:
//
//	0 3 * * *  cd /path/to/backend && go run ./cmd/crossref-sync -filter update-date
//	0 4 * * 0  cd /path/to/backend && go run ./cmd/crossref-sync -filter index-date
//
// Set CROSSREF_MAILTO (or -mailto) to use Crossref's polite pool.
// Works are upserted into the PostgreSQL database at DATABASE_URL, which
// may also be set in a .env file.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"

	"github.com/paperstacks.io/paperstacks/internal/importer"
	"github.com/paperstacks.io/paperstacks/internal/importer/crossref/api"
	paperapp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	paperpostgres "github.com/paperstacks.io/paperstacks/internal/paper/repository/postgres"
)

// Default directories, relative to the backend root the command is run
// from (see Usage above).
const (
	stateDir = "tmp/crossref-sync/state"
	logDir   = "tmp/crossref-sync/logs"
)

// defaultFrom is the first day crawled when a filter has no state yet:
// a few days before the cut-off of the March 2026 Crossref public data
// export the database was seeded from.
const defaultFrom = "2026-02-25"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// run parses flags, builds the updater, and runs it until all open days
// are crawled or ctx is cancelled (e.g. by Ctrl+C, via main's signal.NotifyContext).
func run(ctx context.Context) error {
	yesterday := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -1)

	filter := flag.String("filter", string(api.FilterUpdateDate), "Crossref date to sync by: update-date or index-date")
	from := flag.String("from", defaultFrom, "First day (YYYY-MM-DD) to crawl when the filter has no state yet")
	until := flag.String("until", yesterday.Format(time.DateOnly), "Last day (YYYY-MM-DD) to crawl, inclusive")
	mailto := flag.String("mailto", os.Getenv("CROSSREF_MAILTO"), "Contact address for Crossref's polite pool (default $CROSSREF_MAILTO)")
	state := flag.String("state-dir", stateDir, "Directory for per-filter state and lock files")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	f := api.Filter(*filter)
	if !f.IsValid() {
		return fmt.Errorf("flag -filter: unknown filter %q", *filter)
	}
	fromDay, err := time.Parse(time.DateOnly, *from)
	if err != nil {
		return fmt.Errorf("flag -from: %w", err)
	}
	untilDay, err := time.Parse(time.DateOnly, *until)
	if err != nil {
		return fmt.Errorf("flag -until: %w", err)
	}

	level := slog.LevelInfo
	if *debug {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	cfg := importer.UpdaterConfig{
		Filter:   f,
		From:     fromDay,
		Until:    untilDay,
		StateDir: *state,
		LogDir:   logDir,
	}
	printConfig(os.Stdout, cfg)

	db, err := openDatabase(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	writer := importer.NewServiceWriter(paperapp.NewPaperService(paperpostgres.NewRepository(db), nil))
	client := api.NewClient(api.DefaultBaseURL, *mailto, nil, log)
	return importer.NewUpdater(cfg, log, client, writer).Run(ctx)
}

// openDatabase connects to the PostgreSQL database at DATABASE_URL.
func openDatabase(ctx context.Context) (*pgxpool.Pool, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	db, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}

func printConfig(w io.Writer, cfg importer.UpdaterConfig) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	fmt.Fprintln(tw)
	fmt.Fprintln(tw, "  Filter:\t"+cfg.Filter)
	fmt.Fprintln(tw, "  From:\t"+cfg.From.Format("2006-01-02"))
	fmt.Fprintln(tw, "  Until:\t"+cfg.Until.Format("2006-01-02"))
	fmt.Fprintln(tw, "  StateDir:\t"+cfg.StateDir)
	fmt.Fprintln(tw, "  LogDir:\t"+cfg.LogDir)
	fmt.Fprintln(tw)

	_ = tw.Flush()
}
