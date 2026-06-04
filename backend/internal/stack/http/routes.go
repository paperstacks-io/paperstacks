package http

import (
	"log/slog"
	"net/http"

	paperApp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	"github.com/paperstacks.io/paperstacks/internal/server/middleware"
	"github.com/paperstacks.io/paperstacks/internal/stack/application"
	userApp "github.com/paperstacks.io/paperstacks/internal/user/application"
)

func AddStackRoute(
	mux *http.ServeMux,
	logger *slog.Logger,
	stackService *application.StackService,
	userService *userApp.UserService,
	paperService *paperApp.PaperService,
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
	mux.Handle(http.MethodGet+" /stacks/{uuid}", defaultMiddle(handleGetStack(
		logger,
		stackService,
		userService,
	)))
	mux.Handle(http.MethodDelete+" /stacks/{uuid}", defaultMiddle(handleDeleteStack(
		logger,
		stackService,
		userService,
	)))
	mux.Handle(http.MethodGet+" /stacks/{uuid}/papers", defaultMiddle(handleListPapersInStack(
		logger,
		stackService,
		userService,
	)))
	mux.Handle(http.MethodPost+" /stacks/{uuid}/papers/{paperUuid}", defaultMiddle(handleAddPaperInStack(
		logger,
		stackService,
		userService,
		paperService,
	)))
	mux.Handle(http.MethodDelete+" /stacks/{uuid}/papers/{paperUuid}", defaultMiddle(handleDeletePaperInStack(
		logger,
		stackService,
		userService,
	)))
}
