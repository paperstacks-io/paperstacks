package http

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	"github.com/paperstacks.io/paperstacks/internal/paper/application"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

func handleSearchPapers(logger *slog.Logger, service *application.PaperService) http.Handler {
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

			logger.Error("read papers", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := NewPaperResponses(result.Items)
		setSearchPaginationHeaders(w, result)

		if err := server.Encode(w, r, http.StatusOK, resp); err != nil {
			logger.Error("encode paper response", "error", err)
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

func handleGetPaperByUUID(logger *slog.Logger, service *application.PaperService) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("uuid")
			if id == "" {
				http.Error(w, "missing paper uuid", http.StatusBadRequest)
				return
			}

			p, err := service.GetByUUID(r.Context(), id)
			if err != nil {
				if errors.Is(err, domain.ErrPaperNotFound) {
					logger.Error("read paper", "uuid", id, "error", err.Error())
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}

				logger.Error("read paper", "uuid", id, "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			resp := NewPaperResponse(p)
			if err := server.Encode(w, r, http.StatusOK, resp); err != nil {
				logger.Error("encode paper resp", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}

func handleGetPaperByDOI(logger *slog.Logger, service *application.PaperService) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("doi")
			if id == "" {
				http.Error(w, "missing paper id", http.StatusBadRequest)
				return
			}

			p, err := service.GetByDOI(r.Context(), id)
			if err != nil {

				if errors.Is(err, domain.ErrPaperNotFound) {
					logger.Error("read paper", "doi", id, "error", err.Error())
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}

				logger.Error("read paper", "doi", id, "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			resp := NewPaperResponse(p)
			if err := server.Encode(w, r, http.StatusOK, resp); err != nil {
				logger.Error("encode paper resp", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}

func handleDeletePaper(logger *slog.Logger, service *application.PaperService) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("uuid")
			if id == "" {
				http.Error(w, "missing paper uuid", http.StatusBadRequest)
				return
			}

			if err := service.Delete(r.Context(), id); err != nil {
				logger.Error("delete paper", "uuid", id, "error", err)
				if errors.Is(err, domain.ErrPaperNotFound) {
					http.Error(w, "paper not found with UUID: "+id, http.StatusNotFound)
					return
				}

				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		},
	)
}

func handleSavePaper(logger *slog.Logger, service *application.PaperService) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			req, err := server.Decode[PaperRequest](r)
			if err != nil {
				http.Error(w, "invalid req body", http.StatusBadRequest)
				return
			}

			p := req.toDomain()
			created, err := service.Create(r.Context(), p)
			if err != nil {
				if errors.Is(err, domain.ErrPaperAlreadyExists) {
					logger.Error("create paper", "doi", p.DOI, "error", err)
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}

				if errors.Is(err, domain.ErrInvalidPaper) {
					http.Error(w, "invalid paper", http.StatusBadRequest)
					return
				}

				logger.Error("create paper", "doi", p.DOI, "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			resp := NewPaperResponse(created)

			loc := server.Location(r) + "/" + created.UUID
			w.Header().Set("location", loc)

			if err := server.Encode(w, r, http.StatusCreated, resp); err != nil {
				logger.Error("encode paper resp", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}

func handleUpdatePaper(logger *slog.Logger, service *application.PaperService) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("uuid")
			if id == "" {
				http.Error(w, "missing paper uuid", http.StatusBadRequest)
				return
			}

			req, err := server.Decode[PaperRequest](r)
			if err != nil {
				http.Error(w, "unable to parse request body as PaperRequest", http.StatusBadRequest)
				return
			}

			p := req.toDomain()
			if err := service.Update(r.Context(), id, p); err != nil {
				logger.Error("update paper", "uuid", id, "error", err.Error())
				if errors.Is(err, domain.ErrPaperNotFound) {
					http.Error(w, "paper not found", http.StatusNotFound)
					return
				}

				if errors.Is(err, domain.ErrInvalidPaper) || errors.Is(err, domain.ErrUUIDMismatch) {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				http.Error(w, "failed to update paper", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		},
	)
}

func normalizeQueryParam(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
