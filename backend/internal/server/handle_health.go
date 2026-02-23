package server

import (
	"log/slog"
	"net/http"
)

func handleHealth() http.Handler {
	health := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			slog.Debug("request", "method", "GET", "endpoint", "/health", "status", 200)

			err := encode(w, r, http.StatusOK, health)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}
