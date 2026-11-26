package pdfextractor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/common/objectstorage"
)

const xmlContentType = "application/xml"

// SyncStore is the subset of object storage operations required by Sync.
// It is satisfied by *objectstorage.S3Store.
type SyncStore interface {
	Get(ctx context.Context, key string) (*objectstorage.Object, error)
	Put(ctx context.Context, input objectstorage.PutObjectInput) (objectstorage.ObjectInfo, error)
	Walk(ctx context.Context, prefix string, fn func(key string) error) error
}

// Sync walks the configured object storage prefix in lexicographical order,
// identifies every .pdf object that does not have a matching .xml object,
// and extracts and stores the XML for each missing PDF.
// A single bad PDF does not abort the run; only context cancellation does.
func Sync(ctx context.Context, store SyncStore, extractor Extractor, prefix string, log *slog.Logger) error {
	if store == nil {
		return fmt.Errorf("sync pdf xml: object store is nil")
	}
	if extractor == nil {
		return fmt.Errorf("sync pdf xml: extractor is nil")
	}
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	log.Info("pdf-xml-sync starting", slog.String("prefix", prefix))
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

		xmlKey, err := extractAndStoreXML(ctx, store, extractor, pdfKey)
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

	walkErr := store.Walk(ctx, prefix, func(key string) error {
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

	log.Info("pdf-xml-sync finished",
		slog.Int("processed", processed),
		slog.Int("skipped", skipped),
		slog.Int("errors", errs),
		slog.Duration("elapsed", time.Since(start)),
	)

	return nil
}

func extractAndStoreXML(ctx context.Context, store SyncStore, extractor Extractor, pdfKey string) (string, error) {
	obj, err := store.Get(ctx, pdfKey)
	if err != nil {
		return "", fmt.Errorf("open PDF for XML extraction: %w", err)
	}
	defer obj.Body.Close()

	xmlData, err := extractor.ExtractFulltext(ctx, obj.Body)
	if err != nil {
		return "", fmt.Errorf("extract XML from PDF: %w", err)
	}

	xmlKey := xmlStorageKey(pdfKey)
	_, err = store.Put(ctx, objectstorage.PutObjectInput{
		Key:         xmlKey,
		Body:        bytes.NewReader(xmlData),
		Size:        int64(len(xmlData)),
		ContentType: xmlContentType,
	})
	if err != nil {
		return "", fmt.Errorf("store extracted XML: %w", err)
	}

	return xmlKey, nil
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
