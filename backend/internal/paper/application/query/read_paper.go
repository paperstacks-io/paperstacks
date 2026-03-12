package query

import (
	"context"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

type ReadPaperHandler struct {
	repository domain.Repository
}

func NewReadPaperHandler(repository domain.Repository) ReadPaperHandler {
	return ReadPaperHandler{
		repository: repository,
	}
}

func (p *ReadPaperHandler) Handle(ctx context.Context, id string) (domain.Paper, error) {
	return p.repository.Read(ctx, id)
}
