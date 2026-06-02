package http

import (
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/server/middleware"
	"github.com/paperstacks.io/paperstacks/internal/stack/application"
	userApp "github.com/paperstacks.io/paperstacks/internal/user/application"
)

func AddStackRoute(
	mux *http.ServeMux,
	logger *slog.Logger,
	stackService *application.StackService,
	userService *userApp.UserService,
) {
	defaultMiddle := middleware.NewDefault(logger)

	mux.Handle(http.MethodGet+" /stacks", defaultMiddle(handleListAllPublicStacks(
		logger,
		stackService,
		userService,
	)))
	mux.Handle(http.MethodPost+" /stacks", defaultMiddle(handleCreateStack(
		logger,
		stackService,
		userService,
	)))
}
