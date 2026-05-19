package http

import (
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/server/middleware"
	"github.com/paperstacks.io/paperstacks/internal/stack/application"
	userApp "github.com/paperstacks.io/paperstacks/internal/user/application"
	"github.com/paperstacks.io/paperstacks/internal/web/auth"
)

func AddStackRoute(
	mux *http.ServeMux,
	logger *slog.Logger,
	stackService *application.StackService,
	userService *userApp.UserService,
	sessionService auth.SessionService,
) {
	defaultMiddle := middleware.NewDefault(logger)

	mux.Handle("GET /stacks",
		auth.SessionMiddleware(sessionService)(
			defaultMiddle(handleListUserStacks(
				logger,
				stackService,
				userService,
			)),
		),
	)
}
