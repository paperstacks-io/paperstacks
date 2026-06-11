package http

import (
	"log/slog"
	"net/http"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/auth"
	"github.com/paperstacks.io/paperstacks/internal/server/middleware"
	"github.com/paperstacks.io/paperstacks/internal/stack/application"
)

func AddStackRoute(
	mux *http.ServeMux,
	logger *slog.Logger,
	stackService *application.StackService,
	sessionService commonauth.SessionService,
) {
	defaultMiddle := middleware.NewDefault(logger)
	sessionMiddle := commonauth.SessionMiddleware(sessionService)
	requireAuthMiddle := commonauth.RequireAuthAPIMiddleware()

	mux.Handle(http.MethodGet+" /stacks", defaultMiddle(handleListAllPublicStacks(
		logger,
		stackService,
	)))
	mux.Handle(http.MethodPost+" /stacks", defaultMiddle(sessionMiddle(requireAuthMiddle(handleCreateStack(
		logger,
		stackService,
	)))))
	mux.Handle(http.MethodGet+" /stacks/{uuid}", defaultMiddle(sessionMiddle(handleGetStack(
		logger,
		stackService,
	))))
	mux.Handle(http.MethodDelete+" /stacks/{uuid}", defaultMiddle(sessionMiddle(requireAuthMiddle(handleDeleteStack(
		logger,
		stackService,
	)))))
	mux.Handle(http.MethodGet+" /stacks/{uuid}/papers", defaultMiddle(sessionMiddle(handleListPapersInStack(
		logger,
		stackService,
	))))
	mux.Handle(http.MethodPost+" /stacks/{uuid}/papers/{paperUuid}", defaultMiddle(sessionMiddle(requireAuthMiddle(handleAddPaperInStack(
		logger,
		stackService,
	)))))
	mux.Handle(http.MethodDelete+" /stacks/{uuid}/papers/{paperUuid}", defaultMiddle(sessionMiddle(requireAuthMiddle(handleDeletePaperInStack(
		logger,
		stackService,
	)))))
}
