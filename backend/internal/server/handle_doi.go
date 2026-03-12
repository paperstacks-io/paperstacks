package server

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	"github.com/paperstacks.io/paperstacks/internal/doi"
)

func handleDOI(logger *slog.Logger, service *doi.Service) http.Handler {
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
				case errors.Is(err, doi.ErrEmptyDOI):
					http.Error(w, err.Error(), http.StatusBadRequest)
				case errors.Is(err, doi.ErrNotFound):
					http.Error(w, err.Error(), http.StatusNotFound)
				default:
					logger.Error("resolve DOI metadata", "DOI", decodedDOI, "error", err)
					http.Error(w, "failed to resolve doi metadata", http.StatusBadGateway)
				}
				return
			}

			if err := server.Encode(w, r, http.StatusOK, metadata); err != nil {
				logger.Error("encode DOI response", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}
