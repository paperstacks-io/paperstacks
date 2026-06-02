package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	paperApp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
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
	stackService *application.StackService,
	userService *userApp.UserService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		stackUUID := r.PathValue("uuid")

		stack, err := stackService.GetByUUID(ctx, stackUUID)
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
	stackService *application.StackService,
	userService *userApp.UserService,
	paperService *paperApp.PaperService,
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

		stack, err := stackService.GetByUUID(ctx, stackUUID)
		if err != nil {
			logger.Error("get stack", "uuid", stackUUID, "error", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if stack.Owner.ExternalID != user.ExternalID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		req, err := server.Decode[PaperRequest](r)
		if err != nil {
			logger.Error("decode paper request", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		paper := req.toDomain()
		p, err := paperService.GetByDOI(ctx, paper.DOI)
		if err != nil {
			logger.Error("get paper", "doi", paper.DOI, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		stack.Papers = append(stack.Papers, p)

		updated, err := stackService.Update(ctx, stack)
		if err != nil {
			logger.Error("update stack", "uuid", stackUUID, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := NewStackResponse(updated)

		if err := server.Encode(w, r, http.StatusOK, resp); err != nil {
			logger.Error("encode stack response", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func handleDeletePaperInStack(
	logger *slog.Logger,
	stackService *application.StackService,
	userService *userApp.UserService,
	paperService *paperApp.PaperService,
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

		stack, err := stackService.GetByUUID(ctx, stackUUID)
		if err != nil {
			logger.Error("get stack", "uuid", stackUUID, "error", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if stack.Owner.ExternalID != user.ExternalID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		paper, err := paperService.GetByUUID(ctx, paperUUID)
		if err != nil {
			logger.Error("get paper", "uuid", paperUUID, "error", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		papers := make([]paperDomain.Paper, 0, len(stack.Papers))

		found := false
		for _, p := range stack.Papers {
			if p.UUID == paper.UUID {
				found = true
				continue
			}

			papers = append(papers, p)
		}

		if !found {
			http.Error(w, "paper not found in stack", http.StatusNotFound)
			return
		}

		stack.Papers = papers

		updated, err := stackService.Update(ctx, stack)
		if err != nil {
			logger.Error("update stack", "uuid", stackUUID, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := NewStackResponse(updated)

		if err := server.Encode(w, r, http.StatusOK, resp); err != nil {
			logger.Error("encode stack response", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
