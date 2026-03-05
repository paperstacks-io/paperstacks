package server

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/paper"
)

func handleReadPapers(logger *slog.Logger, service *paper.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		// No DTO needed for a single ID request
		id := r.PathValue("paperId")

		w.Header().Set("Content-Type", "application/json")

		if id != "" {
			p, err := service.Read(id)
			if err != nil {
				if errors.Is(err, paper.ErrPaperNotFound) {
					logger.Error("read paper", "paperId", id, "error", "paper "+err.Error())
					http.Error(w, "paper "+err.Error(), http.StatusNotFound)
					return
				}

				logger.Error("read paper", "paperId", id, "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			response := paperToResponse(p)
			if err := encode(w, r, http.StatusOK, response); err != nil {
				logger.Error("encode paper response", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		papers, err := service.ReadAll()
		if err != nil {
			if errors.Is(err, paper.ErrPaperNotFound) {
				logger.Error("read papers", "paperId", id, "error", "papers "+err.Error())
				http.Error(w, "papers "+err.Error(), http.StatusNotFound)
				return
			}

			logger.Error("read papers", "paperId", id, "error", "papers"+err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := encode(w, r, http.StatusOK, papersMapToResponse(papers)); err != nil {
			logger.Error("encode paper response", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func handleDeletePaper(logger *slog.Logger, service *paper.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		id := r.PathValue("paperId")
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
	})
}
