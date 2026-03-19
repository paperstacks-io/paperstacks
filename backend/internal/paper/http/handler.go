package http

import (
	"cmp"
	"errors"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	"github.com/paperstacks.io/paperstacks/internal/paper/application"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

func handleListPapers(logger *slog.Logger, service *application.PaperService) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			papers, err := service.List(r.Context())
			if err != nil {
				if errors.Is(err, domain.ErrPaperNotFound) {
					logger.Error("read papers", "error", err.Error())
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}

				logger.Error("read papers", "error", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			resp := NewPaperResponses(papers)
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
			id := r.PathValue("doi")
			if id == "" {
				http.Error(w, "missing paper id", http.StatusBadRequest)
				return
			}

			if err := service.Delete(r.Context(), id); err != nil {
				if errors.Is(err, domain.ErrPaperNotFound) {
					logger.Error("delete paper", "id", id, "error", err)
					http.Error(w, "paper not found", http.StatusNotFound)
					return
				}

				logger.Error("delete paper", "id", id, "error", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
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
			if err := service.Create(r.Context(), p); err != nil {
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

			w.WriteHeader(http.StatusCreated)
		},
	)
}

func handleUpdatePaper(logger *slog.Logger, service *application.PaperService) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("doi")
			if id == "" {
				http.Error(w, "missing paper doi", http.StatusBadRequest)
				return
			}

			req, err := server.Decode[PaperRequest](r)
			if err != nil {
				http.Error(w, "invalid req body", http.StatusBadRequest)
				return
			}

			p := req.toDomain()
			if err := service.Update(r.Context(), id, p); err != nil {
				if errors.Is(err, domain.ErrPaperNotFound) {
					logger.Error("update paper", "id", id, "error", "paper "+err.Error())
					http.Error(w, "paper not found", http.StatusNotFound)
					return
				}

				if errors.Is(err, domain.ErrInvalidPaper) || errors.Is(err, domain.ErrDOIMismatch) {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				logger.Error("update paper", "id", id, "error", err)
				http.Error(w, "failed to update paper", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		},
	)
}

func handleSearchPapers(logger *slog.Logger, service *application.PaperService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		title := strings.TrimSpace(r.URL.Query().Get("title"))
		keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
		sortBy := r.URL.Query().Get("sortBy")

		var desc bool // desc indicates descending sort order (default: false)
		if sortBy != "" {
			sortBy, _ = strings.CutPrefix(sortBy, " ")
			sortBy, desc = strings.CutPrefix(sortBy, "-")

			sortBy = strings.ToLower(sortBy)
			if sortBy != "title" && sortBy != "year" {
				http.Error(w, "invalid sort field", http.StatusBadRequest)
				return
			}
		}

		papers, err := service.Search(r.Context(), title, keyword)
		if err != nil {
			logger.Error("read papers", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if sortBy != "" && len(papers) > 1 {
			sort.Slice(papers, func(i, j int) bool {
				switch sortBy {
				case "title":
					if desc {
						return compare(papers[i].Title, papers[j].Title, true)
					}
					return compare(papers[i].Title, papers[j].Title, false)

				case "year":
					if desc {
						return compare(papers[i].PublicationYear, papers[j].PublicationYear, true)
					}
					return compare(papers[i].PublicationYear, papers[j].PublicationYear, false)
					
				default:
					return false
				}
			})
		}

		resp := NewPaperResponses(papers)

		if err := server.Encode(w, r, http.StatusOK, resp); err != nil {
			logger.Error("encode paper response", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func compare[T cmp.Ordered](a, b T, desc bool) bool {
	if desc {
		return a > b
	}

	return a < b
}
