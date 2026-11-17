// Package pdfextractor provides a PDF-to-TEI-XML extraction interface, a
// GROBID client implementation, and a parser that maps TEI-XML to the paper
// domain model.
package pdfextractor

import (
	"context"
	"errors"
	"io"
)

var (
	ErrNilPDF          = errors.New("pdf reader is nil")
	ErrInvalidPDF      = errors.New("invalid or unreadable PDF: missing %PDF- header")
	ErrEmptyExtraction = errors.New("grobid returned empty XML content")
)

// Extractor extracts structured XML from a PDF stream.
type Extractor interface {
	ExtractFulltext(ctx context.Context, pdf io.Reader) ([]byte, error)
}
