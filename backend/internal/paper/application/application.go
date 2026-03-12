package application

import (
	"github.com/paperstacks.io/paperstacks/internal/paper/application/command"
	"github.com/paperstacks.io/paperstacks/internal/paper/application/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	CreatePaper command.CreatePaperHandler
}

type Queries struct {
	ReadPapers query.ReadPapersHandler
}
