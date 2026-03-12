package middleware_

import (
	"log/slog"
	"net/http"
)

func NewDefault(
	logger *slog.Logger,
) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return requestLogging(logger, h)
	}
}
