package http

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	"github.com/paperstacks.io/paperstacks/internal/doi/application"
)

func handleDOI(logger *slog.Logger, service *application.DOIService) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			rawDOI := r.PathValue("doi")
			decodedDOI, err := url.PathUnescape(rawDOI)
			if err != nil {
				http.Error(w, "invalid DOI", http.StatusBadRequest)
				return
			}

			metadata, err := service.ResolveMetadata(r.Context(), decodedDOI)
			if err != nil {
				switch {
				case errors.Is(err, application.ErrEmptyDOI):
					http.Error(w, err.Error(), http.StatusBadRequest)
				case errors.Is(err, application.ErrNotFound):
					http.Error(w, err.Error(), http.StatusNotFound)
				default:
					logger.Error("resolve DOI metadata", "DOI", decodedDOI, "error", err)
					http.Error(w, "failed to resolve doi metadata", http.StatusBadGateway)
				}
				return
			}

			response := NewMetadataResponse(metadata)
			if err := server.Encode(w, r, http.StatusOK, response); err != nil {
				logger.Error("encode DOI response", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}
