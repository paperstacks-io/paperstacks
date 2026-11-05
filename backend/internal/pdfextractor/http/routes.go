package http

import (
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/pdfextractor/domain"
)

func AddExtractorRoute(
	mux *http.ServeMux,
	logger *slog.Logger,
	extractor domain.Extractor,
) {
	mux.Handle(http.MethodPost+" /extract/metadata", handleExtractMetadata(logger, extractor))
	mux.Handle(http.MethodPost+" /extract/text", handleExtractText(logger, extractor))
}

func AddHealthzRoute(rootMux *http.ServeMux) {
	rootMux.Handle(http.MethodGet+" /healthz", handleHealthz())
}
