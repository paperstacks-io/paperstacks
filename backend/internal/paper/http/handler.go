package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	"github.com/paperstacks.io/paperstacks/internal/paper/application"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

func handleListPapers(logger *slog.Logger, service application.Service) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				w.Header().Set("Allow", http.MethodGet)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			papers, err := service.List(r.Context())
			if err != nil {
				if errors.Is(err, domain.ErrPaperNotFound) {
					logger.Error("read papers", "error", "papers "+err.Error())
					http.Error(w, "papers "+err.Error(), http.StatusNotFound)
					return
				}

				logger.Error("read papers", "error", "papers "+err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			resp := newPaperResponses(papers)
			if err := server.Encode(w, r, http.StatusOK, resp); err != nil {
				logger.Error("encode paper response", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}

func handleGetPaperByDOI(logger *slog.Logger, service application.Service) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				w.Header().Set("Allow", http.MethodGet)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			// No DTO needed for a single ID request
			id := r.PathValue("doi")
			if id == "" {
				http.Error(w, "missing paper id", http.StatusBadRequest)
				return
			}

			p, err := service.GetByDOI(r.Context(), id)
			if err != nil {

				if errors.Is(err, domain.ErrPaperNotFound) {
					logger.Error("read paper", "doi", id, "error", "paper "+err.Error())
					http.Error(w, "paper "+err.Error(), http.StatusNotFound)
					return
				}

				logger.Error("read paper", "doi", id, "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			response := newPaperResponse(p)
			if err := server.Encode(w, r, http.StatusOK, response); err != nil {
				logger.Error("encode paper response", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}

func handleDeletePaper(logger *slog.Logger, service application.Service) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				w.Header().Set("Allow", http.MethodDelete)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

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

func handleCreatePaper(logger *slog.Logger, service application.Service) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.Header().Set("Allow", http.MethodPost)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			req, err := server.Decode[paperRequest](r)
			if err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
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

func HandleUpdatePaper(logger *slog.Logger, service application.Service) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				w.Header().Set("Allow", http.MethodPut)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			id := r.PathValue("doi")
			if id == "" {
				http.Error(w, "missing paper doi", http.StatusBadRequest)
				return
			}

			req, err := server.Decode[paperRequest](r)
			if err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
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
