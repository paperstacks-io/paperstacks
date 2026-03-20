package application

import (
	"cmp"
	"context"
	"sort"
	"strings"

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

func (s *PaperService) GetByDOI(ctx context.Context, doi string) (domain.Paper, error) {
	return s.repo.GetByDOI(ctx, strings.TrimSpace(doi))
}

func (s *PaperService) Create(ctx context.Context, paper domain.Paper) error {
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

func (s *PaperService) Delete(ctx context.Context, doi string) error {
	return s.repo.Delete(ctx, strings.TrimSpace(doi))
}

func (s *PaperService) Search(ctx context.Context, title, keyword string, sortBy *string, desc bool) ([]domain.Paper, error) {
	keyword = strings.ToLower(keyword)
	title = strings.ToLower(title)
	
	papers, err := s.repo.Search(ctx, title, keyword)
	if err != nil { return nil, err }

	if sortBy != nil { return sortPapersByOrder(papers, *sortBy, desc) }

	return papers, nil
}

func sortPapersByOrder(papers []domain.Paper, sortBy string, desc bool) ([]domain.Paper, error) {
	if len(papers) == 1 {
		return papers, nil
	}

	sort.Slice(papers, func(i, j int) bool {
		switch sortBy {
		case "title":
			if desc {
				return compare(papers[i].Title, papers[j].Title, true)
			}
			return compare(papers[i].Title, papers[j].Title, false)

		case "year":
			if desc {
				return compare(papers[i].PublicationYear, papers[j].PublicationYear, true)
			}
			return compare(papers[i].PublicationYear, papers[j].PublicationYear, false)

		default:
			return false
		}
	})

	return papers, nil
}

func compare[T cmp.Ordered](a, b T, desc bool) bool {
	if desc {
		return a > b
	}

	return a < b
}
