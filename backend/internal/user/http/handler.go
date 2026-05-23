package http

import (
	"errors"
	"log/slog"
	nethttp "net/http"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	stackApplication "github.com/paperstacks.io/paperstacks/internal/stack/application"
	"github.com/paperstacks.io/paperstacks/internal/user/application"
	"github.com/paperstacks.io/paperstacks/internal/user/domain"
)

func handleGetUserByID(logger *slog.Logger, service *application.UserService) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		userID := strings.TrimSpace(r.PathValue("userId"))
		if userID == "" {
			nethttp.Error(w, "missing user id", nethttp.StatusBadRequest)
			return
		}

		user, err := service.GetByExternalID(r.Context(), userID)
		if err != nil {
			if errors.Is(err, domain.ErrUserNotFound) {
				nethttp.Error(w, err.Error(), nethttp.StatusNotFound)
				return
			}

			logger.Error("read user", "userId", userID, "error", err)
			nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
			return
		}

		resp := NewUserResponse(user)
		if err := server.Encode(w, r, nethttp.StatusOK, resp); err != nil {
			logger.Error("encode user response", "error", err)
			nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
			return
		}
	})
}

func handleListUserStacks(
	logger *slog.Logger,
	userService *application.UserService,
	stackService *stackApplication.StackService,
) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		userID := strings.TrimSpace(r.PathValue("userId"))
		if userID == "" {
			nethttp.Error(w, "missing user id", nethttp.StatusBadRequest)
			return
		}

		user, err := userService.GetByExternalID(r.Context(), userID)
		if err != nil {
			if errors.Is(err, domain.ErrUserNotFound) {
				nethttp.Error(w, err.Error(), nethttp.StatusNotFound)
				return
			}

			logger.Error("read user", "userId", userID, "error", err)
			nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
			return
		}

		stacks, err := stackService.ListPublic(r.Context(), user.ExternalID)
		if err != nil {
			logger.Error("read user stacks", "userId", userID, "error", err)
			nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
			return
		}

		resp := NewStackResponses(stacks)
		if err := server.Encode(w, r, nethttp.StatusOK, resp); err != nil {
			logger.Error("encode user stacks response", "error", err)
			nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
			return
		}
	})
}

func handleGetCurrentUser(logger *slog.Logger, service *application.UserService) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			nethttp.Error(w, "missing bearer token", nethttp.StatusUnauthorized)
			return
		}

		user, err := service.ResolveByAuthToken(r.Context(), token)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidAuthToken) {
				nethttp.Error(w, domain.ErrInvalidAuthToken.Error(), nethttp.StatusUnauthorized)
				return
			}

			logger.Error("read current user", "error", err)
			nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
			return
		}

		resp := NewUserResponse(user)
		if err := server.Encode(w, r, nethttp.StatusOK, resp); err != nil {
			logger.Error("encode user response", "error", err)
			nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
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
