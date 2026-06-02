package http

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	"github.com/paperstacks.io/paperstacks/internal/stack/application"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userApp "github.com/paperstacks.io/paperstacks/internal/user/application"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

func handleListAllPublicStacks(
	logger *slog.Logger,
	service *application.StackService,
	userService *userApp.UserService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		users, err := userService.List(ctx)
		if err != nil {
			logger.Error("Failed to list users", "error", err)
			http.Error(w, "Failed to list users", http.StatusInternalServerError)
			return
		}

		stacks := make([]domain.Stack, 0)
		for _, user := range users {
			userStacks, err := service.ListPublic(ctx, user.ExternalID)
			if err != nil {
				logger.Error("Failed to list public stacks for user", "userExternalID", user.ExternalID, "error", err)
				http.Error(w, "Failed to list public stacks for user", http.StatusInternalServerError)
				return
			}

			stacks = append(stacks, userStacks...)
		}

		resp := NewStackResponses(stacks)

		if err := server.Encode(w, r, http.StatusOK, resp); err != nil {
			logger.Error("encode stack response", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func handleCreateStack(
	logger *slog.Logger,
	service *application.StackService,
	userService *userApp.UserService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}

		user, err := userService.ResolveByAuthToken(ctx, token)
		if err != nil {
			if errors.Is(err, userDomain.ErrInvalidAuthToken) {
				http.Error(w, userDomain.ErrInvalidAuthToken.Error(), http.StatusUnauthorized)
				return
			}

			logger.Error("read current user", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		req, err := server.Decode[StackRequest](r)
		if err != nil {
			logger.Error("decode stack request", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		stack := req.toDomain()
		created := domain.NewStack(stack.Name, user)
		created.IsPublic = stack.IsPublic

		if err := service.Create(ctx, *created); err != nil {
			logger.Error("create stack", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := NewStackResponse(*created)
		if err := server.Encode(w, r, http.StatusCreated, resp); err != nil {
			logger.Error("encode stack response", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func bearerToken(header string) (string, bool) {
	token, ok := strings.CutPrefix(strings.TrimSpace(header), "Bearer ")
	if !ok {
		return "", false
	}

	token = strings.TrimSpace(token)
	return token, token != ""
}
