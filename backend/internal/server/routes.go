package server

import (
	"context"
	"log/slog"
	"net/http"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	"github.com/paperstacks.io/paperstacks/internal/common/server/middleware"
)

func AddRoute(
	mux *http.ServeMux,
	ctx context.Context,
	logger *slog.Logger,
	sessionService commonauth.SessionService,
) {
	defaultMiddle := middleware.NewDefault(logger, sessionService)

	mux.Handle(http.MethodGet+" /{$}", defaultMiddle(handleRoot()))
	mux.Handle(http.MethodGet+" /healthz", defaultMiddle(handleHealthz()))
}
