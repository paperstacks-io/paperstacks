package middleware

import (
	"log/slog"
	"net/http"
)

func requestLogging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Incomming request", slog.String("method", r.Method), slog.String("path", r.URL.Path))
		next.ServeHTTP(w, r)
	})
}
