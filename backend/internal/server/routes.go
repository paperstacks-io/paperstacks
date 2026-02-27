package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/doi"
	"github.com/paperstacks.io/paperstacks/internal/server/middleware"
)

func AddRoute(
	mux *http.ServeMux,
	ctx context.Context,
	logger *slog.Logger,
	doiService *doi.Service,
) http.Handler {
	defaultMiddle := middleware.NewDefault(logger)
	mux.Handle(http.MethodGet+" /healthz", defaultMiddle(handleHealthz()))
	mux.Handle(http.MethodGet+" /example", defaultMiddle(handleExample()))
	mux.Handle(http.MethodGet+" /doi/{doi...}", defaultMiddle(handleDOI(logger, doiService)))
	return mux
}
