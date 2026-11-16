package importer

import (
	"context"
	"errors"

	paperapp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	paper "github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

// PaperWriter persists a changed Crossref work. Implementations must be
// idempotent: the same DOI can arrive repeatedly (overlapping windows,
// re-runs of a failed day, update-date and index-date both reporting it).
type PaperWriter interface {
	Upsert(ctx context.Context, p paper.Paper) error
}

// NoopWriter discards every paper.
type NoopWriter struct{}

func (NoopWriter) Upsert(context.Context, paper.Paper) error { return nil }

// ServiceWriter upserts papers by DOI through a PaperService.
type ServiceWriter struct {
	papers *paperapp.PaperService
}

// NewServiceWriter returns a ServiceWriter writing to papers.
func NewServiceWriter(papers *paperapp.PaperService) *ServiceWriter {
	return &ServiceWriter{papers: papers}
}

// Upsert creates p if no paper with its DOI exists, and otherwise updates
// the existing paper's Crossref-sourced fields, keeping everything else
// (e.g. keywords, publication status, license) as it is. A DOI held by a
// soft-deleted paper yields paper.ErrPaperAlreadyExists.
func (w *ServiceWriter) Upsert(ctx context.Context, p paper.Paper) error {
	existing, err := w.papers.GetByDOI(ctx, p.DOI)
	if errors.Is(err, paper.ErrPaperNotFound) {
		_, err := w.papers.Create(ctx, p)
		return err
	}
	if err != nil {
		return err
	}
	return w.papers.Update(ctx, existing.UUID, withCrossrefFields(existing, p))
}

// withCrossrefFields returns existing with every field toPaper fills
// replaced by the value from p.
func withCrossrefFields(existing, p paper.Paper) paper.Paper {
	existing.Title = p.Title
	existing.Abstract = p.Abstract
	existing.Type = p.Type
	existing.PublicationDate = p.PublicationDate
	existing.Authors = p.Authors

	m := &existing.Metadata
	m.Publisher = p.Metadata.Publisher
	m.JournalTitle = p.Metadata.JournalTitle
	m.JournalAbbrev = p.Metadata.JournalAbbrev
	m.Pages = p.Metadata.Pages
	m.Volume = p.Metadata.Volume
	m.Issue = p.Metadata.Issue
	m.ISSN = p.Metadata.ISSN
	m.ISBN = p.Metadata.ISBN
	m.DataSource = p.Metadata.DataSource
	m.DataSourceTimestamp = p.Metadata.DataSourceTimestamp
	return existing
}
