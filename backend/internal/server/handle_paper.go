package server

import (
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/paper"
)

func handlePaper(logger *slog.Logger, service *paper.Service) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
	)
}
