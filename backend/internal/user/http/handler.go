package http

import (
	"errors"
	"log/slog"
	nethttp "net/http"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
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

func handleListCurrentUserStacks(
	logger *slog.Logger,
	stackService *stackApplication.StackService,
) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		session, _ := commonauth.SessionFromContext(r.Context())
		stacks, err := stackService.List(r.Context(), session.UserID)
		if err != nil {
			logger.Error("read current user stacks", "userId", session.UserID, "error", err)
			nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
			return
		}

		resp := NewStackResponses(stacks)
		if err := server.Encode(w, r, nethttp.StatusOK, resp); err != nil {
			logger.Error("encode current user stacks response", "error", err)
			nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
			return
		}
	})
}

func handleGetCurrentUser(logger *slog.Logger, service *application.UserService) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		session, _ := commonauth.SessionFromContext(r.Context())
		user, err := service.GetByExternalID(r.Context(), session.UserID)
		if err != nil {
			logger.Error("read current user", "userId", session.UserID, "error", err)
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
