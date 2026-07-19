package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	"github.com/paperstacks.io/paperstacks/internal/document/application"
	"github.com/paperstacks.io/paperstacks/internal/document/domain"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

func handleUploadDocument(
	logger *slog.Logger,
	service *application.DocumentService,
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			session, ok := commonauth.SessionFromContext(ctx)
			if !ok || session == nil || !session.IsValid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
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

			paperUUID := r.FormValue("paper_uuid")
			if paperUUID == "" {
				http.Error(w, "paper_uuid is required", http.StatusBadRequest)
				return
			}

			fileName := fileHeader.Filename
			if fileName == "" {
				http.Error(w, "file_name is required", http.StatusBadRequest)
				return
			}

			uploadedDocument, err := service.Upload(ctx, paperUUID, fileName, session.UserID, file)
			if err != nil {
				if errors.Is(err, domain.ErrFileSizeExceeded) || errors.Is(err, domain.ErrInvalidFileType) || errors.Is(err, paperDomain.ErrPaperNotFound) {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				logger.Error("upload document", "user_id", session.UserID, "error", err)
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
