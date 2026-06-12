package middleware

import (
	"log/slog"
	"net/http"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
)

func NewDefault(
	logger *slog.Logger,
	sessionService commonauth.SessionService,
) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return requestLogging(logger, commonauth.SessionMiddleware(sessionService)(h))
	}
}
