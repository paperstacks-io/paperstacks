package importer

import (
	"context"

	paper "github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

// PaperWriter persists a changed Crossref work. Implementations must be
// idempotent: the same DOI can arrive repeatedly (overlapping windows,
// re-runs of a failed day, update-date and index-date both reporting it).
type PaperWriter interface {
	Upsert(ctx context.Context, p paper.Paper) error
}

// NoopWriter discards every paper. It stands in until changes are written
// to the database.
type NoopWriter struct{}

func (NoopWriter) Upsert(context.Context, paper.Paper) error { return nil }
