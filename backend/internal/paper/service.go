package paper

import (
	"net/http"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/domain"
)

type Service struct {
	paperRepo Repository
	client    *http.Client
}

func NewService(client *http.Client, repo Repository) *Service {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	return &Service{
		paperRepo: repo,
		client:    client,
	}
}

func (r *Service) ReadAll() (map[string]domain.Paper, error) {
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
