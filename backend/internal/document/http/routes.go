package http

import (
	"log/slog"
	"net/http"
	"time"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	"github.com/paperstacks.io/paperstacks/internal/common/server/middleware"
	"github.com/paperstacks.io/paperstacks/internal/document/application"
	"golang.org/x/time/rate"
)

func UploadDocumentRoute(
	mux *http.ServeMux,
	logger *slog.Logger,
	documentService *application.DocumentService,
	sessionService commonauth.SessionService,
) {
	defaultMiddle := middleware.NewDefault(logger, sessionService)

	const (
		ipRateLimitInterval = 12 * time.Second
		ipRateLimitBurst    = 3
	)
	// Rate limiter: 5 requests per minute (1 request every 12 seconds) with a burst of 3
	ipLimiter := middleware.NewIPRateLimiter(rate.Every(ipRateLimitInterval), ipRateLimitBurst)
	rateLimitMiddle := middleware.RateLimit(ipLimiter)

	mux.Handle("POST /document",
		commonauth.SessionMiddleware(sessionService)(
			rateLimitMiddle(
				defaultMiddle(handleUploadDocument(
					logger,
					documentService,
				)),
			),
		),
	)
}
