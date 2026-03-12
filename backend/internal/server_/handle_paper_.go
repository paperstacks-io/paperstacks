package server_

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/paper_"
)

func handleReadPapers(logger *slog.Logger, service *paper_.Service) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				w.Header().Set("Allow", http.MethodGet)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			papers, err := service.ReadAll()
			if err != nil {
				if errors.Is(err, paper_.ErrPaperNotFound) {
					logger.Error("read papers", "error", "papers "+err.Error())
					http.Error(w, "papers "+err.Error(), http.StatusNotFound)
					return
				}

				logger.Error("read papers", "error", "papers "+err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := encode(w, r, http.StatusOK, papersToResponse(papers)); err != nil {
				logger.Error("encode paper_ response", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}

func handleReadPaper(logger *slog.Logger, service *paper_.Service) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				w.Header().Set("Allow", http.MethodGet)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			// No DTO needed for a single ID request
			id := r.PathValue("paperId")
			if id == "" {
				http.Error(w, "missing paper_ id", http.StatusBadRequest)
				return
			}

			p, err := service.Read(id)
			if err != nil {
				if errors.Is(err, paper_.ErrPaperNotFound) {
					logger.Error("read paper_", "paperId", id, "error", "paper_ "+err.Error())
					http.Error(w, "paper_ "+err.Error(), http.StatusNotFound)
					return
				}

				logger.Error("read paper_", "paperId", id, "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			response := paperToResponse(p)
			if err := encode(w, r, http.StatusOK, response); err != nil {
				logger.Error("encode paper_ response", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}

func handleDeletePaper(logger *slog.Logger, service *paper_.Service) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				w.Header().Set("Allow", http.MethodDelete)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			id := r.PathValue("paperId")
			if id == "" {
				http.Error(w, "missing paper_ id", http.StatusBadRequest)
				return
			}

			if err := service.Delete(id); err != nil {
				if errors.Is(err, paper_.ErrPaperNotFound) {
					logger.Error("delete paper_", "id", id, "error", err)
					http.Error(w, "paper_ not found", http.StatusNotFound)
					return
				}

				logger.Error("delete paper_", "id", id, "error", err)
				http.Error(w, "internal server_ error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		},
	)
}

func handleCreatePaper(logger *slog.Logger, service *paper_.Service) http.Handler {
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
				if errors.Is(err, paper_.ErrPaperAlreadyExists) {
					logger.Error("create paper_", "doi_", p.DOI, "error", err)
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}

				logger.Error("create paper_", "doi_", p.DOI, "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
		},
	)
}

func handleUpdatePaper(logger *slog.Logger, service *paper_.Service) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				w.Header().Set("Allow", http.MethodPut)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			id := r.PathValue("paperId")
			if id == "" {
				http.Error(w, "missing paper_ id", http.StatusBadRequest)
				return
			}

			req, err := decode[UpdatePaperRequest](r)
			if err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}

			p := req.ToDomain()
			if err := service.Update(id, p); err != nil {
				if errors.Is(err, paper_.ErrPaperNotFound) {
					logger.Error("update paper_", "id", id, "error", "paper_ "+err.Error())
					http.Error(w, "paper_ not found", http.StatusNotFound)
					return
				}

				logger.Error("update paper_", "id", id, "error", err)
				http.Error(w, "failed to update paper_", http.StatusInternalServerError)
				return
			}
		},
	)
}
