package application

import (
	"context"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) Service {
	return Service{repo: repo}
}

func (s Service) List(ctx context.Context) ([]domain.Paper, error) {
	return s.repo.List(ctx)
}

func (s Service) GetByDOI(ctx context.Context, doi string) (domain.Paper, error) {
	return s.repo.GetByDOI(ctx, strings.TrimSpace(doi))
}

func (s Service) Create(ctx context.Context, paper domain.Paper) error {
	paper = paper.Normalize()
	if err := paper.Validate(); err != nil {
		return err
	}

	return s.repo.Save(ctx, paper)
}

func (s Service) Update(ctx context.Context, doi string, paper domain.Paper) error {
	doi = strings.TrimSpace(doi)
	paper = paper.Normalize()

	if doi == "" {
		return domain.ErrInvalidPaper
	}

	if paper.DOI == "" {
		paper.DOI = doi
	} else if paper.DOI != doi {
		return domain.ErrDOIMismatch
	}

	if err := paper.Validate(); err != nil {
		return err
	}

	return s.repo.Update(ctx, doi, paper)
}

func (s Service) Delete(ctx context.Context, doi string) error {
	return s.repo.Delete(ctx, strings.TrimSpace(doi))
}
