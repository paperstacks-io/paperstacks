package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/common/server/middleware"
	"github.com/paperstacks.io/paperstacks/internal/paper/infrastructure/http/handlers"
	"github.com/paperstacks.io/paperstacks/internal/paper/service"
)

func AddRouteTest(
	mux *http.ServeMux,
	ctx context.Context,
	logger *slog.Logger,
	paperApplication *service.Application,

) http.Handler {
	defaultMiddle := middleware.NewDefault(logger)
	mux.Handle(http.MethodGet+" /papers/", defaultMiddle(handlers.HandleReadPapers(ctx, logger, paperApplication)))
	return mux
}
