package application

import (
	"context"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/paper/bibliography"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

// BibliographyService coordinates bibliography operations with Paper storage.
type BibliographyService struct {
	repo domain.Repository
}

func NewBibliographyService(repo domain.Repository) *BibliographyService {
	return &BibliographyService{repo: repo}
}

// ExportBibLaTeX loads Papers by UUID and exports them in request order.
func (s *BibliographyService) ExportBibLaTeX(ctx context.Context, uuids []string) ([]byte, error) {
	papers := make([]domain.Paper, 0, len(uuids))
	for _, uuid := range uuids {
		paper, err := s.repo.GetByUUID(ctx, strings.TrimSpace(uuid))
		if err != nil {
			return nil, err
		}
		papers = append(papers, paper)
	}

	return bibliography.ExportBibLaTeX(papers)
}
