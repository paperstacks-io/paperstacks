package command

import (
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

func (p *CreatePaperHandler) Handle(paper domain.Paper) error {
	return nil
}
