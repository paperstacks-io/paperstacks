package paper

import (
	"context"

	"github.com/paperstacks.io/paperstacks/internal/old/domain"
)

type ServiceConfiguration func(service *Service) error

type Service struct {
	paperRepo Repository
}

func NewService(cfgs ...ServiceConfiguration) (*Service, error) {
	service := &Service{}

	for _, cfg := range cfgs {
		err := cfg(service)

		if err != nil {
			return nil, err
		}
	}

	return service, nil
}

func (r *Service) ResolveMetadata(ctx context.Context) {
}

// This is an in-memory repository implementation.
// It stores data only in memory and is mainly intended for development or testing.
// It can easily be replaced with another implementation, such as a database-backed repository.
func MemoryRepository(memoryRepo Repository) ServiceConfiguration {
	return func(service *Service) error {
		service.paperRepo = memoryRepo
		return nil
	}
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
