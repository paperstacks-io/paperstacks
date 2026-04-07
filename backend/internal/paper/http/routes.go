package http

import (
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/paper/application"
	"github.com/paperstacks.io/paperstacks/internal/server/middleware"
)

func AddPaperRoute(
	mux *http.ServeMux,
	logger *slog.Logger,
	paperService *application.PaperService,
) {
	defaultMiddle := middleware.NewDefault(logger)

	mux.Handle(http.MethodGet+" /papers", defaultMiddle(handleListOrSearchPapers(logger, paperService)))
	mux.Handle(http.MethodGet+" /papers/uuid/{uuid...}", defaultMiddle(handleGetPaperByUUID(logger, paperService)))
	mux.Handle(http.MethodGet+" /papers/doi/{doi...}", defaultMiddle(handleGetPaperByDOI(logger, paperService)))

	mux.Handle(http.MethodDelete+" /papers/doi/{doi...}", defaultMiddle(handleDeletePaper(logger, paperService)))
	mux.Handle(http.MethodPost+" /papers", defaultMiddle(handleSavePaper(logger, paperService)))
	mux.Handle(http.MethodPut+" /papers/doi/{doi...}", defaultMiddle(handleUpdatePaper(logger, paperService)))
}
