package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/doi"
	"github.com/paperstacks.io/paperstacks/internal/paper"
	"github.com/paperstacks.io/paperstacks/internal/server/middleware"
)

func AddRoute(
	mux *http.ServeMux,
	ctx context.Context,
	logger *slog.Logger,
	doiService *doi.Service,
	paperService *paper.Service,
) http.Handler {
	defaultMiddle := middleware.NewDefault(logger)
	mux.Handle(http.MethodGet+" /healthz", defaultMiddle(handleHealthz()))
	mux.Handle(http.MethodGet+" /example", defaultMiddle(handleExample()))
	mux.Handle(http.MethodGet+" /doi/{doi...}", defaultMiddle(handleDOI(logger, doiService)))
	mux.Handle(http.MethodGet+" /papers/", defaultMiddle(handleReadPapers(logger, paperService)))
	mux.Handle(http.MethodGet+" /papers/{paperId}", defaultMiddle(handleReadPapers(logger, paperService)))

	// e.g curl -X DELETE http://localhost:8080/papers/delete/paper-1
	mux.Handle(http.MethodDelete+" /papers/delete/{paperId}", defaultMiddle(handleDeletePaper(logger, paperService)))

	return mux
}
