package application

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/common/objectstorage"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

const xmlContentType = "application/xml"

// XMLExtractor extracts structured XML from a PDF stream.
type XMLExtractor interface {
	ExtractFulltext(ctx context.Context, pdf io.Reader) ([]byte, error)
}

type keyWalker interface {
	Walk(ctx context.Context, prefix string, fn func(key string) error) error
}

func (s *PaperService) SetXMLExtractor(extractor XMLExtractor) {
	s.xmlExtractor = extractor
}

// XMLExtractionEnabled reports whether an XMLExtractor is configured.
func (s *PaperService) XMLExtractionEnabled() bool {
	return s.xmlExtractor != nil
}

// ExtractPDFXML verifies that storageKey exists in object storage, extracts
// structured XML from pdf via the configured XMLExtractor, and stores the
// resulting .xml in object storage without re-downloading the PDF.
func (s *PaperService) ExtractPDFXML(ctx context.Context, storageKey string, pdf io.Reader) (string, error) {
	if s.xmlExtractor == nil {
		return "", nil
	}
	storageKey = strings.TrimSpace(storageKey)
	if storageKey == "" {
		return "", domain.ErrInvalidPaper
	}

	exists, err := s.storage.Exists(ctx, storageKey)
	if err != nil {
		return "", fmt.Errorf("failed to check PDF existence: %w", err)
	}
	if !exists {
		return "", objectstorage.ErrObjectNotFound
	}

	return s.extractAndPutXML(ctx, storageKey, pdf)
}

// ExtractStoredPDFXML opens storageKey from object storage, extracts its XML
// via the configured XMLExtractor, and stores the resulting .xml in object storage.
func (s *PaperService) ExtractStoredPDFXML(ctx context.Context, storageKey string) (string, error) {
	if s.xmlExtractor == nil {
		return "", nil
	}
	storageKey = strings.TrimSpace(storageKey)
	if storageKey == "" {
		return "", domain.ErrInvalidPaper
	}

	obj, err := s.storage.Get(ctx, storageKey)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF for XML extraction: %w", err)
	}
	defer obj.Body.Close()

	return s.extractAndPutXML(ctx, storageKey, obj.Body)
}

func (s *PaperService) extractAndPutXML(ctx context.Context, storageKey string, pdf io.Reader) (string, error) {
	xmlData, err := s.xmlExtractor.ExtractFulltext(ctx, pdf)
	if err != nil {
		return "", fmt.Errorf("failed to extract XML from PDF: %w", err)
	}

	xmlKey := xmlStorageKey(storageKey)
	_, err = s.storage.Put(ctx, objectstorage.PutObjectInput{
		Key:         xmlKey,
		Body:        bytes.NewReader(xmlData),
		Size:        int64(len(xmlData)),
		ContentType: xmlContentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to store extracted XML: %w", err)
	}

	return xmlKey, nil
}

// SyncPDFXMLs walks the configured object storage prefix in lexicographical order,
// identifies every .pdf object that does not have a matching .xml object,
// and extracts and stores the XML for each missing PDF.
// A single bad PDF does not abort the run; only context cancellation does.
func (s *PaperService) SyncPDFXMLs(ctx context.Context, prefix string, log *slog.Logger) error {
	if s.xmlExtractor == nil {
		return nil
	}
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	walker, ok := s.storage.(keyWalker)
	if !ok {
		return fmt.Errorf("object storage does not support walking keys")
	}

	log.Info("pdfextractor starting", slog.String("prefix", prefix))
	start := time.Now()

	var (
		processed int
		skipped   int
		errs      int
		pending   []string
	)

	processOne := func(pdfKey string) error {
		if err := ctx.Err(); err != nil {
			return err
		}

		xmlKey, err := s.ExtractStoredPDFXML(ctx, pdfKey)
		if err != nil {
			errs++
			log.Warn("failed to extract pdf",
				slog.String("pdf", pdfKey),
				slog.String("error", err.Error()),
			)
			return ctx.Err()
		}

		processed++
		log.Info("extracted xml",
			slog.String("pdf", pdfKey),
			slog.String("xml", xmlKey),
		)
		return nil
	}

	walkErr := walker.Walk(ctx, prefix, func(key string) error {
		if err := ctx.Err(); err != nil {
			return err
		}

		// Resolve any pending PDFs whose expected XML key is <= current key.
		remaining := pending[:0]
		for _, pdfKey := range pending {
			expectedXML := xmlStorageKey(pdfKey)
			switch {
			case key == expectedXML:
				skipped++
				log.Debug("skipping pdf with existing xml",
					slog.String("pdf", pdfKey),
					slog.String("xml", expectedXML),
				)
			case key > expectedXML:
				if err := processOne(pdfKey); err != nil {
					return err
				}
			default:
				remaining = append(remaining, pdfKey)
			}
		}
		pending = remaining

		if isPDFKey(key) {
			pending = append(pending, key)
		}

		return nil
	})
	if walkErr != nil {
		return walkErr
	}

	// Process any remaining PDFs at the end of the bucket listing.
	for _, pdfKey := range pending {
		if err := processOne(pdfKey); err != nil {
			return err
		}
	}

	log.Info("pdfextractor finished",
		slog.Int("processed", processed),
		slog.Int("skipped", skipped),
		slog.Int("errors", errs),
		slog.Duration("elapsed", time.Since(start)),
	)

	return nil
}

func xmlStorageKey(pdfKey string) string {
	pdfKey = strings.TrimSpace(pdfKey)
	if isPDFKey(pdfKey) {
		return pdfKey[:len(pdfKey)-4] + ".xml"
	}
	return pdfKey + ".xml"
}

func isPDFKey(key string) bool {
	return len(key) > 4 && strings.EqualFold(key[len(key)-4:], ".pdf")
}
