package command

import (
	"context"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

type CreatePaperHandler struct {
	repository domain.Repository
}

func NewCreatePaperHandler(repository domain.Repository) CreatePaperHandler {
	return CreatePaperHandler{
		repository: repository,
	}
}

func (p *CreatePaperHandler) Handle(ctx context.Context, paper domain.Paper) error {
	return p.repository.Create(ctx, paper)
}
