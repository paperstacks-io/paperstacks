package http

import (
	"log/slog"
	"net/http"
	"time"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	"github.com/paperstacks.io/paperstacks/internal/common/server/middleware"
	"github.com/paperstacks.io/paperstacks/internal/document/application"
	paperApp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	userApp "github.com/paperstacks.io/paperstacks/internal/user/application"
	"golang.org/x/time/rate"
)

func UploadDocumentRoute(
	mux *http.ServeMux,
	logger *slog.Logger,
	documentService *application.DocumentService,
	userService *userApp.UserService,
	paperService *paperApp.PaperService,
	sessionService commonauth.SessionService,
) {
	defaultMiddle := middleware.NewDefault(logger, sessionService)
	requireAuthMiddle := commonauth.RequireAuthAPIMiddleware()

	// Rate limiter: 5 requests per minute (1 request every 12 seconds) with a burst of 3
	ipLimiter := middleware.NewIPRateLimiter(rate.Every(12*time.Second), 3)
	rateLimitMiddle := middleware.RateLimit(ipLimiter)

	mux.Handle("POST /document", rateLimitMiddle(defaultMiddle(requireAuthMiddle(handleUploadDocument(
		logger,
		documentService,
		userService,
		paperService,
	)))))
}
