package http

import (
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/stacks/application"
)

func AddStackRoute(
	mux *http.ServeMux,
	logger *slog.Logger,
	stackService *application.StackService,
) {
}
