package server

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/paper"
)

func handleReadPapers(logger *slog.Logger, service *paper.Service) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				w.Header().Set("Allow", http.MethodGet)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			papers, err := service.ReadAll()
			if err != nil {
				if errors.Is(err, paper.ErrPaperNotFound) {
					logger.Error("read papers", "error", "papers "+err.Error())
					http.Error(w, "papers "+err.Error(), http.StatusNotFound)
					return
				}

				logger.Error("read papers", "error", "papers "+err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := encode(w, r, http.StatusOK, papersToResponse(papers)); err != nil {
				logger.Error("encode paper response", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}

func handleReadPaper(logger *slog.Logger, service *paper.Service) http.Handler {
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

			p, err := service.Read(id)
			if err != nil {
				if errors.Is(err, paper.ErrPaperNotFound) {
					logger.Error("read paper", "doi", id, "error", "paper "+err.Error())
					http.Error(w, "paper "+err.Error(), http.StatusNotFound)
					return
				}

				logger.Error("read paper", "doi", id, "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			response := paperToResponse(p)
			if err := encode(w, r, http.StatusOK, response); err != nil {
				logger.Error("encode paper response", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}

func handleDeletePaper(logger *slog.Logger, service *paper.Service) http.Handler {
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

			if err := service.Delete(id); err != nil {
				if errors.Is(err, paper.ErrPaperNotFound) {
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

func handleCreatePaper(logger *slog.Logger, service *paper.Service) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.Header().Set("Allow", http.MethodPost)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			req, err := decode[CreatePaperRequest](r)
			if err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}

			p := req.ToDomain()
			if err := service.Create(p); err != nil {
				if errors.Is(err, paper.ErrPaperAlreadyExists) {
					logger.Error("create paper", "doi", p.DOI, "error", err)
					http.Error(w, err.Error(), http.StatusConflict)
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

func handleUpdatePaper(logger *slog.Logger, service *paper.Service) http.Handler {
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

			req, err := decode[UpdatePaperRequest](r)
			if err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}

			p := req.ToDomain()
			if err := service.Update(id, p); err != nil {
				if errors.Is(err, paper.ErrPaperNotFound) {
					logger.Error("update paper", "id", id, "error", "paper "+err.Error())
					http.Error(w, "paper not found", http.StatusNotFound)
					return
				}

				logger.Error("update paper", "id", id, "error", err)
				http.Error(w, "failed to update paper", http.StatusInternalServerError)
				return
			}
		},
	)
}
