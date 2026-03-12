package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/common/server/middleware"
	"github.com/paperstacks.io/paperstacks/internal/paper/ports"
)

func AddRouteTest(
	mux *http.ServeMux,
	ctx context.Context,
	logger *slog.Logger,
	paperHttpServer *ports.HttpServer,
) http.Handler {
	defaultMiddle := middleware.NewDefault(logger)
	mux.Handle(http.MethodGet+" /papers/", defaultMiddle(paperHttpServer.HandleReadPapers(ctx, logger)))
	return mux
}
