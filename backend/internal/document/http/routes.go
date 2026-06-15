package http

import (
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/time/rate"
	"github.com/paperstacks.io/paperstacks/internal/document/application"
	paperApp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	"github.com/paperstacks.io/paperstacks/internal/server/middleware"
	userApp "github.com/paperstacks.io/paperstacks/internal/user/application"
	"github.com/paperstacks.io/paperstacks/internal/web/auth"
)

func UploadDocumentRoute(
	mux *http.ServeMux,
	logger *slog.Logger,
	documentService *application.DocumentService,
	userService *userApp.UserService,
	paperService *paperApp.PaperService,
	sessionService auth.SessionService,
) {
	defaultMiddle := middleware.NewDefault(logger)

	// Rate limiter: 5 requests per minute (1 request every 12 seconds) with a burst of 3
	ipLimiter := middleware.NewIPRateLimiter(rate.Every(12*time.Second), 3)
	rateLimitMiddle := middleware.RateLimit(ipLimiter)

	mux.Handle("POST /document",
		auth.SessionMiddleware(sessionService)(
			rateLimitMiddle(
				defaultMiddle(handleUploadDocument(
					logger,
					documentService,
					userService,
					paperService,
				)),
			),
		),
	)
}
