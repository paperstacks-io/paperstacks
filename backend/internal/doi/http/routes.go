package http

import (
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/doi/application"
	"github.com/paperstacks.io/paperstacks/internal/server/middleware"
)

func AddDOIRoute(
	mux *http.ServeMux,
	logger *slog.Logger,
	doiService *application.DOIService,
) {
	defaultMiddle := middleware.NewDefault(logger)

	mux.Handle(http.MethodGet+" /doi/{doi...}", defaultMiddle(handleDOI(logger, doiService)))
}
