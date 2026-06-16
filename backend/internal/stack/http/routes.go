package http

import (
	"log/slog"
	"net/http"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	"github.com/paperstacks.io/paperstacks/internal/common/server/middleware"
	"github.com/paperstacks.io/paperstacks/internal/stack/application"
)

func AddStackRoute(
	mux *http.ServeMux,
	logger *slog.Logger,
	stackService *application.StackService,
	sessionService commonauth.SessionService,
) {
	defaultMiddle := middleware.NewDefault(logger, sessionService)
	requireAuthMiddle := commonauth.RequireAuthAPIMiddleware()

	mux.Handle(http.MethodGet+" /stacks", defaultMiddle(handleSearchStacks(
		logger,
		stackService,
	)))
	mux.Handle(http.MethodPost+" /stacks", defaultMiddle(requireAuthMiddle(handleCreateStack(
		logger,
		stackService,
	))))
	mux.Handle(http.MethodGet+" /stacks/{uuid}", defaultMiddle(handleGetStack(
		logger,
		stackService,
	)))
	mux.Handle(http.MethodDelete+" /stacks/{uuid}", defaultMiddle(requireAuthMiddle(handleDeleteStack(
		logger,
		stackService,
	))))
	mux.Handle(http.MethodGet+" /stacks/{uuid}/papers", defaultMiddle(handleListPapersInStack(
		logger,
		stackService,
	)))
	mux.Handle(http.MethodPost+" /stacks/{uuid}/papers/{paperUuid}", defaultMiddle(requireAuthMiddle(handleAddPaperInStack(
		logger,
		stackService,
	))))
	mux.Handle(http.MethodDelete+" /stacks/{uuid}/papers/{paperUuid}", defaultMiddle(requireAuthMiddle(handleDeletePaperInStack(
		logger,
		stackService,
	))))
}
