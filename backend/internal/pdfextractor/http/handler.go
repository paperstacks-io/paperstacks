package http

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	commonServer "github.com/paperstacks.io/paperstacks/internal/common/server"
	"github.com/paperstacks.io/paperstacks/internal/pdfextractor/domain"
)

func handleExtractMetadata(
	logger *slog.Logger,
	extractor domain.Extractor,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		file, fileHeader, err := r.FormFile(FormFieldFile)
		if err != nil {
			if errors.Is(err, http.ErrMissingFile) {
				http.Error(w, "file is required", http.StatusBadRequest)
				return
			}
			http.Error(w, "invalid multipart form data", http.StatusBadRequest)
			return
		}
		defer file.Close()

		opts := domain.Options{
			HeadPages: 2,
			Timeout:   30 * time.Second,
		}

		if pagesParam := r.FormValue(FormFieldHeadPages); pagesParam != "" {
			if p, err := strconv.Atoi(pagesParam); err == nil && p > 0 {
				opts.HeadPages = p
			}
		}

		startTime := time.Now()
		meta, err := extractor.ExtractMetadata(ctx, file, opts)
		if err != nil {
			logger.Error("extract metadata failed", "file", fileHeader.Filename, "error", err)
			http.Error(w, "failed to extract metadata", http.StatusInternalServerError)
			return
		}

		logger.Info("metadata extraction completed",
			"file", fileHeader.Filename,
			"doi", meta.DOI,
			"is_scanned", meta.IsScanned,
			"missing_fields", meta.MissingFields,
			"duration_ms", time.Since(startTime).Milliseconds(),
		)

		resp := NewMetadataResponse(meta)
		if err := commonServer.Encode(w, r, http.StatusOK, resp); err != nil {
			logger.Error("encode metadata response", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func handleExtractText(
	logger *slog.Logger,
	extractor domain.Extractor,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		file, fileHeader, err := r.FormFile(FormFieldFile)
		if err != nil {
			if errors.Is(err, http.ErrMissingFile) {
				http.Error(w, "file is required", http.StatusBadRequest)
				return
			}
			http.Error(w, "invalid multipart form data", http.StatusBadRequest)
			return
		}
		defer file.Close()

		opts := domain.Options{
			ExtractAllPages: true,
			Timeout:         60 * time.Second,
		}

		if headParam := r.FormValue(FormFieldHeadPages); headParam != "" {
			if p, err := strconv.Atoi(headParam); err == nil && p > 0 {
				opts.HeadPages = p
				opts.ExtractAllPages = false
			}
		}

		startTime := time.Now()
		extractedText, err := extractor.ExtractText(ctx, file, opts)
		if err != nil {
			logger.Error("extract text failed", "file", fileHeader.Filename, "error", err)
			http.Error(w, "failed to extract text", http.StatusInternalServerError)
			return
		}

		logger.Info("text extraction completed",
			"file", fileHeader.Filename,
			"char_count", len(extractedText.FullText),
			"is_scanned", extractedText.IsScanned,
			"duration_ms", time.Since(startTime).Milliseconds(),
		)

		resp := NewTextResponse(extractedText)
		if err := commonServer.Encode(w, r, http.StatusOK, resp); err != nil {
			logger.Error("encode text response", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func handleHealthz() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK\n"))
	})
}
