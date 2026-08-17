package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/paperstacks.io/paperstacks/internal/paper/bibliography"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

// BibliographyService coordinates bibliography operations with Paper storage.
type BibliographyService struct {
	papers *PaperService
}

type BibLaTeXImportResult struct {
	Created  []bibliography.ImportedPaper
	Existing []bibliography.ImportedPaper
	Rejected []RejectedBibLaTeXEntry
}

type RejectedBibLaTeXEntry struct {
	SourceKey string
	Warnings  []bibliography.Diagnostic
	Errors    []bibliography.Diagnostic
}

func NewBibliographyService(papers *PaperService) *BibliographyService {
	return &BibliographyService{papers: papers}
}

// ExportBibLaTeX loads Papers by UUID and exports them in request order.
func (s *BibliographyService) ExportBibLaTeX(ctx context.Context, uuids []string) ([]byte, error) {
	papers := make([]domain.Paper, 0, len(uuids))
	for _, uuid := range uuids {
		paper, err := s.papers.GetByUUID(ctx, uuid)
		if err != nil {
			return nil, err
		}
		papers = append(papers, paper)
	}

	return bibliography.ExportBibLaTeX(papers)
}

// ImportBibLaTeX parses source and persists every valid, previously unknown Paper.
func (s *BibliographyService) ImportBibLaTeX(ctx context.Context, source []byte) (BibLaTeXImportResult, error) {
	parsed, err := bibliography.ImportBibLaTeX(source)
	if err != nil {
		return BibLaTeXImportResult{}, err
	}

	result := BibLaTeXImportResult{
		Created:  make([]bibliography.ImportedPaper, 0, len(parsed.Entries)),
		Existing: make([]bibliography.ImportedPaper, 0),
		Rejected: make([]RejectedBibLaTeXEntry, 0, len(parsed.Errors)),
	}
	errorsByKey := make(map[string][]bibliography.Diagnostic, len(parsed.Errors))
	for _, diagnostic := range parsed.Errors {
		errorsByKey[diagnostic.EntryKey] = append(errorsByKey[diagnostic.EntryKey], diagnostic)
	}

	for _, entry := range parsed.Entries {
		if diagnostics := errorsByKey[entry.SourceKey]; len(diagnostics) > 0 {
			result.Rejected = append(result.Rejected, RejectedBibLaTeXEntry{
				SourceKey: entry.SourceKey,
				Warnings:  entry.Warnings,
				Errors:    diagnostics,
			})
			continue
		}

		existing, err := s.papers.GetByDOI(ctx, entry.Paper.DOI)
		switch {
		case err == nil:
			entry.Paper = existing
			result.Existing = append(result.Existing, entry)
			continue
		case !errors.Is(err, domain.ErrPaperNotFound):
			return result, fmt.Errorf("check existing paper for BibLaTeX entry %q: %w", entry.SourceKey, err)
		}

		created, err := s.papers.Create(ctx, entry.Paper)
		if err != nil {
			return result, fmt.Errorf("create paper for BibLaTeX entry %q: %w", entry.SourceKey, err)
		}
		entry.Paper = created
		result.Created = append(result.Created, entry)
	}

	return result, nil
}
