package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/server/middleware"
)

var domainPath string = " /v2/papers/"

func AddPaperRoute(
	mux *http.ServeMux,
	ctx context.Context,
	logger *slog.Logger,
	paperRepo domain.Repository,
) {
	defaultMiddle := middleware.NewDefault(logger)

	mux.Handle(http.MethodGet+domainPath, defaultMiddle(HandleReadPapers(logger, paperRepo)))
	mux.Handle(http.MethodGet+domainPath+"/doi/{doi...}", defaultMiddle(HandleReadPaper(logger, paperRepo)))
	mux.Handle(http.MethodDelete+domainPath+"/doi/{doi...}", defaultMiddle(HandleDeletePaper(logger, paperRepo)))
	mux.Handle(http.MethodPost+domainPath, defaultMiddle(HandleCreatePaper(logger, paperRepo)))
	mux.Handle(http.MethodPut+domainPath+"/doi/{doi...}", defaultMiddle(HandleUpdatePaper(logger, paperRepo)))
}
