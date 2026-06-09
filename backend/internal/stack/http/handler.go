package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/application"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userApp "github.com/paperstacks.io/paperstacks/internal/user/application"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

func handleListAllPublicStacks(
	logger *slog.Logger,
	service *application.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		stacks, err := service.ListAllPublic(ctx)
		if err != nil {
			logger.Error("list all public stacks", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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

		user, err := currentUser(ctx, r, userService)
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

func handleGetStack(
	logger *slog.Logger,
	service *application.StackService,
	userService *userApp.UserService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.PathValue("uuid")
		if id == "" {
			http.Error(w, "missing stack uuid", http.StatusBadRequest)
			return
		}

		stack, err := service.GetByUUID(ctx, id)
		if err != nil {
			if errors.Is(err, domain.ErrStackNotFound) {
				http.Error(w, domain.ErrStackNotFound.Error(), http.StatusNotFound)
				return
			}

			logger.Error("get stack", "uuid", id, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !stack.IsPublic {
			user, err := currentUser(ctx, r, userService)
			if err != nil {
				if errors.Is(err, userDomain.ErrInvalidAuthToken) {
					http.Error(w, userDomain.ErrInvalidAuthToken.Error(), http.StatusUnauthorized)
					return
				}

				logger.Error("read current user", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if stack.Owner.ExternalID != user.ExternalID {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
		}

		resp := NewStackResponse(stack)
		if err := server.Encode(w, r, http.StatusOK, resp); err != nil {
			logger.Error("encode stack response", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func handleDeleteStack(
	logger *slog.Logger,
	service *application.StackService,
	userService *userApp.UserService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, err := currentUser(ctx, r, userService)
		if err != nil {
			if errors.Is(err, userDomain.ErrInvalidAuthToken) {
				http.Error(w, userDomain.ErrInvalidAuthToken.Error(), http.StatusUnauthorized)
				return
			}

			logger.Error("read current user", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		stackUUID := r.PathValue("uuid")

		stack, err := service.GetByUUID(ctx, stackUUID)
		if err != nil {
			logger.Error("get stack", "uuid", stackUUID, "error", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if stack.Owner.ExternalID != user.ExternalID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		if err := service.Delete(ctx, stackUUID); err != nil {
			logger.Error("delete stack", "uuid", stackUUID, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

func handleListPapersInStack(
	logger *slog.Logger,
	service *application.StackService,
	userService *userApp.UserService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		stackUUID := r.PathValue("uuid")

		stack, err := service.GetByUUID(ctx, stackUUID)
		if err != nil {
			logger.Error("get stack", "uuid", stackUUID, "error", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if !stack.IsPublic {
			user, err := currentUser(ctx, r, userService)
			if err != nil {
				if errors.Is(err, userDomain.ErrInvalidAuthToken) {
					http.Error(w, userDomain.ErrInvalidAuthToken.Error(), http.StatusUnauthorized)
					return
				}

				logger.Error("read current user", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if stack.Owner.ExternalID != user.ExternalID {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
		}

		resp := NewPaperResponses(stack.Papers)
		if err := server.Encode(w, r, http.StatusOK, resp); err != nil {
			logger.Error("encode paper response", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func handleAddPaperInStack(
	logger *slog.Logger,
	service *application.StackService,
	userService *userApp.UserService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, err := currentUser(ctx, r, userService)
		if err != nil {
			if errors.Is(err, userDomain.ErrInvalidAuthToken) {
				http.Error(w, userDomain.ErrInvalidAuthToken.Error(), http.StatusUnauthorized)
				return
			}

			logger.Error("read current user", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		stackUUID := r.PathValue("uuid")
		if stackUUID == "" {
			http.Error(w, "missing stack uuid", http.StatusBadRequest)
			return
		}

		paperUUID := r.PathValue("paperUuid")
		if paperUUID == "" {
			http.Error(w, "missing paper uuid", http.StatusBadRequest)
			return
		}

		stack, err := service.GetByUUID(ctx, stackUUID)
		if err != nil {
			logger.Error("get stack", "uuid", stackUUID, "error", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if stack.Owner.ExternalID != user.ExternalID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		if err := service.AddPaper(ctx, stackUUID, paperUUID); err != nil {
			logger.Error("add paper to stack", "stackUUID", stackUUID, "paperUUID", paperUUID, "error", err)
			if errors.Is(err, paperDomain.ErrPaperNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

func handleDeletePaperInStack(
	logger *slog.Logger,
	service *application.StackService,
	userService *userApp.UserService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, err := currentUser(ctx, r, userService)
		if err != nil {
			if errors.Is(err, userDomain.ErrInvalidAuthToken) {
				http.Error(w, userDomain.ErrInvalidAuthToken.Error(), http.StatusUnauthorized)
				return
			}

			logger.Error("read current user", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		stackUUID := r.PathValue("uuid")
		if stackUUID == "" {
			http.Error(w, "missing stack uuid", http.StatusBadRequest)
			return
		}

		paperUUID := r.PathValue("paperUuid")
		if paperUUID == "" {
			http.Error(w, "missing paper uuid", http.StatusBadRequest)
			return
		}

		stack, err := service.GetByUUID(ctx, stackUUID)
		if err != nil {
			logger.Error("get stack", "uuid", stackUUID, "error", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if stack.Owner.ExternalID != user.ExternalID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		if err := service.RemovePaper(ctx, stackUUID, paperUUID); err != nil {
			logger.Error("remove paper from stack", "stackUUID", stackUUID, "paperUUID", paperUUID, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

func currentUser(
	ctx context.Context,
	r *http.Request,
	userService *userApp.UserService,
) (userDomain.User, error) {
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		return userDomain.User{}, userDomain.ErrInvalidAuthToken
	}

	return userService.ResolveByAuthToken(ctx, token)
}

func bearerToken(header string) (string, bool) {
	token, ok := strings.CutPrefix(strings.TrimSpace(header), "Bearer ")
	if !ok {
		return "", false
	}

	token = strings.TrimSpace(token)
	return token, token != ""
}
