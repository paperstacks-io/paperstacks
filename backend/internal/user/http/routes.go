// Package http provides HTTP routes and handlers for user resources.
package http

import (
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/common/server/middleware"
	stackApplication "github.com/paperstacks.io/paperstacks/internal/stack/application"
	"github.com/paperstacks.io/paperstacks/internal/user/application"
)

func AddUserRoute(
	mux *http.ServeMux,
	logger *slog.Logger,
	userService *application.UserService,
	stackService *stackApplication.StackService,
) {
	defaultMiddle := middleware.NewDefault(logger)

	mux.Handle(http.MethodGet+" /users/me", defaultMiddle(handleGetCurrentUser(logger, userService)))
	mux.Handle(http.MethodGet+" /users/me/stacks", defaultMiddle(handleListCurrentUserStacks(logger, userService, stackService)))
	mux.Handle(http.MethodGet+" /users/{userId}", defaultMiddle(handleGetUserByID(logger, userService)))
	mux.Handle(http.MethodGet+" /users/{userId}/stacks", defaultMiddle(handleListUserStacks(logger, userService, stackService)))
}
