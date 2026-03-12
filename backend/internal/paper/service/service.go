package service

import (
	"context"

	"github.com/paperstacks.io/paperstacks/internal/paper/application"
	"github.com/paperstacks.io/paperstacks/internal/paper/application/command"
	"github.com/paperstacks.io/paperstacks/internal/paper/application/query"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

func NewApplication(ctx context.Context, repository domain.Repository) application.Application {
	_ = ctx
	return application.Application{
		Commands: application.Commands{
			CreatePaper: command.NewCreatePaperHandler(repository),
		},
		Queries: application.Queries{
			ReadPapers: query.NewReadPapersHandler(repository),
		},
	}
}
