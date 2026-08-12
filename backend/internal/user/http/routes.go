// Package http provides HTTP routes and handlers for user resources.
package http

import (
	"log/slog"
	"net/http"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	"github.com/paperstacks.io/paperstacks/internal/common/server/middleware"
	stackApplication "github.com/paperstacks.io/paperstacks/internal/stack/application"
	"github.com/paperstacks.io/paperstacks/internal/user/application"
)

func AddUserRoute(
	mux *http.ServeMux,
	logger *slog.Logger,
	userService *application.UserService,
	userProvisioner *application.UserProvisioner,
	stackService *stackApplication.StackService,
	sessionService commonauth.SessionService,
) {
	defaultMiddle := middleware.NewDefault(logger, sessionService)

	mux.Handle(http.MethodGet+" /users/me", defaultMiddle(handleGetCurrentUser(logger, userProvisioner)))
	mux.Handle(http.MethodGet+" /users/me/stacks", defaultMiddle(handleListCurrentUserStacks(logger, userProvisioner, stackService)))
	mux.Handle(http.MethodGet+" /users/{userId}", defaultMiddle(handleGetUserByID(logger, userService)))
	mux.Handle(http.MethodGet+" /users/{userId}/stacks", defaultMiddle(handleListUserStacks(logger, userService, stackService)))
}
