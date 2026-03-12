package service

import (
	"context"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

type Application struct {
	repository domain.Repository
}

func NewApplication(repository domain.Repository) *Application {
	return &Application{
		repository: repository,
	}
}

func (r *Application) ReadAll(ctx context.Context) ([]domain.Paper, error) {
	return r.repository.ReadAll(ctx)
}
