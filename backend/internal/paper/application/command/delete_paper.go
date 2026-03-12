package command

import (
	"context"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

type DeletePaperHandler struct {
	repository domain.Repository
}

func NewDeletePaperHandler(repository domain.Repository) DeletePaperHandler {
	return DeletePaperHandler{
		repository: repository,
	}
}

func (p *DeletePaperHandler) Handle(ctx context.Context, id string) error {
	return p.repository.Delete(ctx, id)
}
