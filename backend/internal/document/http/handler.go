package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	"github.com/paperstacks.io/paperstacks/internal/document/application"
	"github.com/paperstacks.io/paperstacks/internal/document/domain"
	paperService "github.com/paperstacks.io/paperstacks/internal/paper/application"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	userService "github.com/paperstacks.io/paperstacks/internal/user/application"
)

func handleUploadDocument(
	logger *slog.Logger,
	service *application.DocumentService,
	userService *userService.UserService,
	paperService *paperService.PaperService,
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			session, ok := commonauth.SessionFromContext(ctx)
			if !ok || session == nil || !session.IsValid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := userService.GetByExternalID(ctx, session.UserID)
			if err != nil {
				logger.Error("get user by external id", "user_id", session.UserID, "error", err)
				http.Error(w, "failed to get user", http.StatusInternalServerError)
				return
			}

			file, fileHeader, err := r.FormFile("file")
			if err != nil {
				if errors.Is(err, http.ErrMissingFile) {
					http.Error(w, "file is required", http.StatusBadRequest)
					return
				}
				http.Error(w, "invalid multipart form data", http.StatusBadRequest)
				return
			}
			defer file.Close()

			req := DocumentRequest{
				PaperUUID: r.FormValue("paper_uuid"),
				FileName:  fileHeader.Filename,
			}

			if err := req.ValidateUploadRequest(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if _, err := paperService.GetByUUID(ctx, req.PaperUUID); err != nil {
				if errors.Is(err, paperDomain.ErrPaperNotFound) {
					http.Error(w, "paper not found", http.StatusBadRequest)
					return
				}
				logger.Error("validate paper exists", "paper_uuid", req.PaperUUID, "error", err)
				http.Error(w, "failed to validate paper", http.StatusInternalServerError)
				return
			}

			d := req.toDomain()
			uploadedDocument, err := service.Upload(ctx, d, user, file)
			if err != nil {
				if errors.Is(err, domain.ErrFileSizeExceeded) || errors.Is(err, domain.ErrInvalidFileType) {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				logger.Error("upload document", "user_id", user.ExternalID, "error", err)
				http.Error(w, "failed to upload document", http.StatusInternalServerError)
				return
			}

			resp := NewDocumentResponse(uploadedDocument)

			if err := server.Encode(w, r, http.StatusCreated, resp); err != nil {
				logger.Error("encode document", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		},
	)
}
