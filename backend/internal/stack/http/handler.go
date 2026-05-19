package http

import (
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	"github.com/paperstacks.io/paperstacks/internal/stack/application"
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
