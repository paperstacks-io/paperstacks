package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/common/server/middleware"
)

func AddRouteTest(
	mux *http.ServeMux,
	ctx context.Context,
	logger *slog.Logger,
	paperHttpServer *Server,
) http.Handler {
	defaultMiddle := middleware.NewDefault(logger)
	mux.Handle(http.MethodGet+" /papers/", defaultMiddle(paperHttpServer.HandleReadPapers(ctx, logger)))
	mux.Handle(http.MethodGet+" /papers/doi/{doi...}", defaultMiddle(paperHttpServer.HandleReadPaper(ctx, logger)))
	mux.Handle(http.MethodPost+" /papers/", defaultMiddle(paperHttpServer.HandleCreatePaper(ctx, logger)))
	mux.Handle(http.MethodDelete+" /papers/doi/{doi...}", defaultMiddle(paperHttpServer.HandleDeletePaper(ctx, logger)))
	mux.Handle(http.MethodPut+" /papers/doi/{doi...}", defaultMiddle(paperHttpServer.HandleUpdatePaper(ctx, logger)))
	return mux
}
