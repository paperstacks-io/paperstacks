package http

import (
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	"github.com/paperstacks.io/paperstacks/internal/stack/application"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userService "github.com/paperstacks.io/paperstacks/internal/user/application"
	"github.com/paperstacks.io/paperstacks/internal/web/auth"
)

func handleListUserStacks(
	logger *slog.Logger,
	service *application.StackService,
	userService *userService.UserService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		session, ok := auth.SessionFromContext(ctx)
		if !ok || session == nil || !session.IsValid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := userService.GetByExternalID(ctx, session.UserID)
		if err != nil {
			logger.Error("get user by external id", "user_id", session.UserID, "error", err)
			http.Error(w, "failed to get user", http.StatusInternalServerError)
			return
		}

		stacks, err := service.List(ctx, user)
		if err != nil {
			logger.Error("list user stacks", "user_id", user.ExternalID, "error", err)
			http.Error(w, "failed to list stacks", http.StatusInternalServerError)
			return
		}

		resp := NewStackResponses(stacks)

		if err := server.Encode(w, r, http.StatusOK, resp); err != nil {
			logger.Error("encode stacks", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func handleCreateUserStack(
	logger *slog.Logger,
	service *application.StackService,
	userService *userService.UserService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		session, ok := auth.SessionFromContext(ctx)
		if !ok || session == nil || !session.IsValid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := userService.GetByExternalID(ctx, session.UserID)
		if err != nil {
			logger.Error("get user by external id", "user_id", session.UserID, "error", err)
			http.Error(w, "failed to get user", http.StatusInternalServerError)
			return
		}

		req, err := server.Decode[StackRequest](r)
		if err != nil {
			http.Error(w, "invalid req body", http.StatusBadRequest)
			return
		}

		s := req.toDomain()
		stack := domain.NewStack(s.Name, user)
		createdStack, err := service.Create(ctx, *stack)
		if err != nil {
			logger.Error("create stack", "user_id", user.ExternalID, "error", err)
			http.Error(w, "failed to create stack", http.StatusInternalServerError)
			return
		}

		resp := NewStackResponse(createdStack)

		if err := server.Encode(w, r, http.StatusCreated, resp); err != nil {
			logger.Error("encode stack", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
