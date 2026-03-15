package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

func handleGetPapersByField(
	logger *slog.Logger,
	paramName string,
	redirectBase string,
	getPapers func(context.Context, string) ([]domain.Paper, error),
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.PathValue(paramName)
		if value == "" {
			http.Error(w, "missing "+paramName, http.StatusBadRequest)
			return
		}

		trimmed := strings.TrimSpace(value)
		if trimmed != value {
			target := redirectBase + "/" + url.PathEscape(trimmed)
			http.Redirect(w, r, target, http.StatusMovedPermanently)
			return
		}

		papers, err := getPapers(r.Context(), trimmed)
		if err != nil {
			if errors.Is(err, domain.ErrPaperNotFound) {
				logger.Error("read paper", paramName, trimmed, "error", err.Error())
				http.Error(w, "paper not found", http.StatusNotFound)
				return
			}

			logger.Error("read paper", paramName, trimmed, "error", err)
			http.Error(w, "failed to read paper", http.StatusInternalServerError)
			return
		}

		response := make([]PaperResponse, 0, len(papers))
		for _, p := range papers {
			response = append(response, NewPaperResponse(p))
		}

		if err := server.Encode(w, r, http.StatusOK, response); err != nil {
			logger.Error("encode papers", paramName, trimmed, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
