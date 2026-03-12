package server_

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/doi_"
	"github.com/paperstacks.io/paperstacks/internal/paper_"
	"github.com/paperstacks.io/paperstacks/internal/server_/middleware_"
)

func AddRoute(
	mux *http.ServeMux,
	ctx context.Context,
	logger *slog.Logger,
	doiService *doi_.Service,
	paperService *paper_.Service,
) http.Handler {
	defaultMiddle := middleware_.NewDefault(logger)
	mux.Handle(http.MethodGet+" /healthz", defaultMiddle(handleHealthz()))
	mux.Handle(http.MethodGet+" /example", defaultMiddle(handleExample()))
	mux.Handle(http.MethodGet+" /doi_/{doi_...}", defaultMiddle(handleDOI(logger, doiService)))
	mux.Handle(http.MethodGet+" /papers/", defaultMiddle(handleReadPapers(logger, paperService)))
	mux.Handle(http.MethodGet+" /papers/{paperId}", defaultMiddle(handleReadPaper(logger, paperService)))
	mux.Handle(http.MethodDelete+" /papers/{paperId}", defaultMiddle(handleDeletePaper(logger, paperService)))
	mux.Handle(http.MethodPost+" /papers/", defaultMiddle(handleCreatePaper(logger, paperService)))
	mux.Handle(http.MethodPut+" /papers/{paperId}", defaultMiddle(handleUpdatePaper(logger, paperService)))
	return mux
}
