// Package http provides HTTP routes and handlers for paper resources.
package http

import (
	"log/slog"
	"net/http"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	"github.com/paperstacks.io/paperstacks/internal/common/server/middleware"
	"github.com/paperstacks.io/paperstacks/internal/paper/application"
)

func AddPaperRoute(
	mux *http.ServeMux,
	logger *slog.Logger,
	paperService *application.PaperService,
	sessionService commonauth.SessionService,
) {
	defaultMiddle := middleware.NewDefault(logger, sessionService)

	mux.Handle(http.MethodGet+" /papers", defaultMiddle(handleSearchPapers(logger, paperService)))
	mux.Handle(http.MethodGet+" /papers/{uuid}", defaultMiddle(handleGetPaperByUUID(logger, paperService)))
	mux.Handle(http.MethodGet+" /papers/doi/{doi...}", defaultMiddle(handleGetPaperByDOI(logger, paperService)))

	mux.Handle(http.MethodPost+" /papers", defaultMiddle(handleSavePaper(logger, paperService)))
	mux.Handle(http.MethodDelete+" /papers/{uuid}", defaultMiddle(handleDeletePaper(logger, paperService)))
	mux.Handle(http.MethodPut+" /papers/{uuid}", defaultMiddle(handleUpdatePaper(logger, paperService)))
}
