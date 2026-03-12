package http

import (
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/paper/application"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/server/middleware"
)

const domainPath = "/v2/papers"

func AddPaperRoute(
	mux *http.ServeMux,
	logger *slog.Logger,
	paperRepo domain.Repository,
) {
	defaultMiddle := middleware.NewDefault(logger)
	paperService := application.NewService(paperRepo)

	mux.Handle(http.MethodGet+" "+domainPath, defaultMiddle(handleListPapers(logger, paperService)))
	mux.Handle(http.MethodGet+" "+domainPath+"/doi/{doi...}", defaultMiddle(handleGetPaperByDOI(logger, paperService)))

	mux.Handle(http.MethodDelete+" "+domainPath+"/doi/{doi...}", defaultMiddle(handleDeletePaper(logger, paperService)))
	mux.Handle(http.MethodPost+" "+domainPath, defaultMiddle(handleCreatePaper(logger, paperService)))
	mux.Handle(http.MethodPut+" "+domainPath+"/doi/{doi...}", defaultMiddle(HandleUpdatePaper(logger, paperService)))
}
