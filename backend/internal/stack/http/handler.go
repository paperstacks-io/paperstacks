package http

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/application"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

func handleCreateStack(
	logger *slog.Logger,
	service *application.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		session, ok := commonauth.SessionFromContext(ctx)
		if !ok || session == nil || !session.IsValid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		req, err := server.Decode[StackRequest](r)
		if err != nil {
			logger.Error("decode stack request", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		stack := req.toDomain()
		user := userDomain.NewUser(session.UserID, session.Email)
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
			session, ok := commonauth.SessionFromContext(ctx)
			if !ok || session == nil || !session.IsValid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if stack.Owner.ExternalID != session.UserID {
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
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		session, ok := commonauth.SessionFromContext(ctx)
		if !ok || session == nil || !session.IsValid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		stackUUID := r.PathValue("uuid")

		stack, err := service.GetByUUID(ctx, stackUUID)
		if err != nil {
			logger.Error("get stack", "uuid", stackUUID, "error", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if stack.Owner.ExternalID != session.UserID {
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
			session, ok := commonauth.SessionFromContext(ctx)
			if !ok || session == nil || !session.IsValid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if stack.Owner.ExternalID != session.UserID {
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
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		session, ok := commonauth.SessionFromContext(ctx)
		if !ok || session == nil || !session.IsValid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
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

		if stack.Owner.ExternalID != session.UserID {
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
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		session, ok := commonauth.SessionFromContext(ctx)
		if !ok || session == nil || !session.IsValid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
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

		if stack.Owner.ExternalID != session.UserID {
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

func handleSearchStacks(
	logger *slog.Logger,
	service *application.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := normalizeQueryParam(r.URL.Query().Get("q"))
		sortByRaw := normalizeQueryParam(r.URL.Query().Get("sortBy"))
		page, _ := strconv.Atoi(normalizeQueryParam(r.URL.Query().Get("page")))
		pageSize, _ := strconv.Atoi(normalizeQueryParam(r.URL.Query().Get("pageSize")))

		sortBy, desc := strings.CutPrefix(sortByRaw, "-")
		sortBy, _ = strings.CutPrefix(sortBy, "+")

		result, err := service.Search(r.Context(), domain.SearchOptions{
			Query:    query,
			SortBy:   sortBy,
			Desc:     desc,
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			if errors.Is(err, domain.ErrInvalidSearch) {
				http.Error(w, "invalid search options", http.StatusBadRequest)
				return
			}

			logger.Error("read stacks", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := NewStackResponses(result.Items)
		setSearchPaginationHeaders(w, result)

		if err := server.Encode(w, r, http.StatusOK, resp); err != nil {
			logger.Error("encode stack response", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func setSearchPaginationHeaders(w http.ResponseWriter, result domain.SearchResult) {
	w.Header().Set("X-Page", strconv.Itoa(result.Page))
	w.Header().Set("X-Page-Size", strconv.Itoa(result.PageSize))

	w.Header().Set("X-Total-Count", strconv.Itoa(result.Total))
	w.Header().Set("X-Has-Next", strconv.FormatBool(result.HasNext))
	w.Header().Set("Access-Control-Expose-Headers", "X-Page, X-Page-Size, X-Total-Count, X-Has-Next")
}

func normalizeQueryParam(param string) string {
	return strings.ToLower(strings.TrimSpace(param))
}
