package paper

import (
	"context"

	"github.com/paperstacks.io/paperstacks/internal/domain"
)

type Service struct {
	paperRepo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		paperRepo: repo,
	}
}

func (r *Service) ResolveMetadata(ctx context.Context) {
}

func (r *Service) ReadAll() ([]domain.Paper, error) {
	return r.paperRepo.ReadAll()
}

func (r *Service) Read(id string) (domain.Paper, error) {
	return r.paperRepo.Read(id)
}

func (r *Service) Delete(id string) error {
	return r.paperRepo.Delete(id)
}

func (r *Service) Create(paper domain.Paper) error {
	return r.paperRepo.Create(paper)
}

func (r *Service) Update(id string, paper domain.Paper) error {
	return r.paperRepo.Update(id, paper)
}
