package query

import (
	"context"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

type ReadPapersHandler struct {
	repository domain.Repository
}

func NewReadPapersHandler(repository domain.Repository) ReadPapersHandler {
	return ReadPapersHandler{
		repository: repository,
	}
}

func (p *ReadPapersHandler) Handle(ctx context.Context) ([]domain.Paper, error) {
	return p.repository.ReadAll(ctx)
}
