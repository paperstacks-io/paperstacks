package server_

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/paperstacks.io/paperstacks/internal/doi_"
)

func handleDOI(logger *slog.Logger, service *doi_.Service) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			rawDOI := r.PathValue("doi_")
			decodedDOI, err := url.PathUnescape(rawDOI)
			if err != nil {
				http.Error(w, "invalid DOI", http.StatusBadRequest)
				return
			}

			metadata, err := service.ResolveMetadata(r.Context(), decodedDOI)
			if err != nil {
				switch {
				case errors.Is(err, doi_.ErrEmptyDOI):
					http.Error(w, err.Error(), http.StatusBadRequest)
				case errors.Is(err, doi_.ErrNotFound):
					http.Error(w, err.Error(), http.StatusNotFound)
				default:
					logger.Error("resolve DOI metadata", "DOI", decodedDOI, "error", err)
					http.Error(w, "failed to resolve doi_ metadata", http.StatusBadGateway)
				}
				return
			}

			if err := encode(w, r, http.StatusOK, metadata); err != nil {
				logger.Error("encode DOI response", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}
