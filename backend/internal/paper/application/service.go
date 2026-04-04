package application

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

type PaperService struct {
	repo domain.Repository
}

func NewPaperService(repo domain.Repository) *PaperService {
	return &PaperService{repo: repo}
}

func (s *PaperService) List(ctx context.Context) ([]domain.Paper, error) {
	return s.repo.List(ctx)
}

func (s *PaperService) GetByUUID(ctx context.Context, uuid string) (domain.Paper, error) {
	return s.repo.GetByUUID(ctx, strings.TrimSpace(uuid))
}

func (s *PaperService) GetByDOI(ctx context.Context, doi string) (domain.Paper, error) {
	return s.repo.GetByDOI(ctx, strings.TrimSpace(doi))
}

func (s *PaperService) Create(ctx context.Context, paper domain.Paper) error {
	paper.UUID = uuid.NewString()
	paper = paper.Normalize()
	if err := paper.Validate(); err != nil {
		return err
	}

	return s.repo.Save(ctx, paper)
}

func (s *PaperService) Update(ctx context.Context, doi string, paper domain.Paper) error {
	doi = strings.TrimSpace(doi)
	paper = paper.Normalize()

	if doi == "" {
		return domain.ErrInvalidPaper
	}

	if err := paper.Validate(); err != nil {
		return err
	}

	return s.repo.Update(ctx, doi, paper)
}

func (s *PaperService) Delete(ctx context.Context, doi string) error {
	return s.repo.Delete(ctx, strings.TrimSpace(doi))
}

func (s *PaperService) Search(ctx context.Context, title, keyword string, sortBy string, orderDesc bool) ([]domain.Paper, error) {
	keyword = strings.ToLower(keyword)
	title = strings.ToLower(title)

	return s.repo.Search(ctx, title, keyword, sortBy, orderDesc)
}
