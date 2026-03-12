package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	"github.com/paperstacks.io/paperstacks/internal/paper"
)

func HandleReadPapers(logger *slog.Logger, paperRepo paper.Repository) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				w.Header().Set("Allow", http.MethodGet)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			papers, err := paperRepo.ReadAll(r.Context())
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

			resp := papersToResponse(papers)
			if err := server.Encode(w, r, http.StatusOK, resp); err != nil {
				logger.Error("encode paper response", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}

func HandleReadPaper(logger *slog.Logger, paperRepo paper.Repository) http.Handler {
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

			p, err := paperRepo.Read(r.Context(), id)
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
			if err := server.Encode(w, r, http.StatusOK, response); err != nil {
				logger.Error("encode paper response", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}

func HandleDeletePaper(logger *slog.Logger, paperRepo paper.Repository) http.Handler {
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

			if err := paperRepo.Delete(r.Context(), id); err != nil {
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

func HandleCreatePaper(logger *slog.Logger, paperRepo paper.Repository) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.Header().Set("Allow", http.MethodPost)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			req, err := server.Decode[CreatePaperRequest](r)
			if err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}

			p := req.ToDomain()
			if err := paperRepo.Create(r.Context(), p); err != nil {
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

func HandleUpdatePaper(logger *slog.Logger, paperRepo paper.Repository) http.Handler {
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

			req, err := server.Decode[UpdatePaperRequest](r)
			if err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}

			p := req.ToDomain()
			if err := paperRepo.Update(r.Context(), id, p); err != nil {
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
