package command

import (
	"context"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

type UpdatePaperHandler struct {
	repository domain.Repository
}

func NewUpdatePaperHandler(repository domain.Repository) UpdatePaperHandler {
	return UpdatePaperHandler{
		repository: repository,
	}
}

func (p *UpdatePaperHandler) Handle(ctx context.Context, id string, paper domain.Paper) error {
	return p.repository.Update(ctx, id, paper)
}
