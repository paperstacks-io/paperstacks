package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/doi"
	"github.com/paperstacks.io/paperstacks/internal/old/paper"
	"github.com/paperstacks.io/paperstacks/internal/paper/db/memory"
	phttp "github.com/paperstacks.io/paperstacks/internal/paper/http"
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
	mux.Handle(http.MethodGet+" /papers/doi/{doi...}", defaultMiddle(handleReadPaper(logger, paperService)))
	mux.Handle(http.MethodDelete+" /papers/doi/{doi...}", defaultMiddle(handleDeletePaper(logger, paperService)))
	mux.Handle(http.MethodPost+" /papers/", defaultMiddle(handleCreatePaper(logger, paperService)))
	mux.Handle(http.MethodPut+" /papers/doi/{doi...}", defaultMiddle(handleUpdatePaper(logger, paperService)))

	paperRepo := memory.NewMemoryRepo()
	mux.Handle(http.MethodGet+" /v2/papers/", defaultMiddle(phttp.HandleReadPapers(logger, paperRepo)))
	mux.Handle(http.MethodGet+" /v2/papers/doi/{doi...}", defaultMiddle(phttp.HandleReadPaper(logger, paperRepo)))
	mux.Handle(http.MethodDelete+" /v2/papers/doi/{doi...}", defaultMiddle(phttp.HandleDeletePaper(logger, paperRepo)))
	mux.Handle(http.MethodPost+" /v2/papers/", defaultMiddle(phttp.HandleCreatePaper(logger, paperRepo)))
	mux.Handle(http.MethodPut+" /v2/papers/doi/{doi...}", defaultMiddle(phttp.HandleUpdatePaper(logger, paperRepo)))

	return mux
}
