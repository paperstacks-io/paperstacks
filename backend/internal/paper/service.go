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

func (r *Service) Create(paper domain.Paper) {
	_, err := r.paperRepo.Create(paper)
	if err != nil {
		return
	}
}
