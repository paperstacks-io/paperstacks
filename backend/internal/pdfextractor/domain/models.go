package domain

import (
	"context"
	"errors"
	"io"
	"time"
)

var (
	ErrPDFToTextNotFound = errors.New("pdftotext command not found (requires poppler-utils)")
	ErrExtractionFailed  = errors.New("failed to extract content from PDF")
	ErrEmptyPDF          = errors.New("empty PDF stream provided")
)

// Options configures how the PDF extraction is performed.
type Options struct {
	EnableOCR       bool          // Whether to use OCR if the document is scanned
	OCRLanguage     string        // Language for OCR (e.g., "eng")
	HeadPages       int           // Number of initial pages to parse (for metadata: title, authors, DOI, abstract)
	TailPages       int           // Number of trailing pages to parse (for references, acknowledgements)
	ExtractAllPages bool          // When true, extracts the complete document regardless of HeadPages/TailPages
	Timeout         time.Duration // Maximum execution time allowed for extraction
}

// ExtractedMetadata represents the bibliographic metadata parsed from the PDF.
type ExtractedMetadata struct {
	Title           string     `json:"title"`
	Authors         []string   `json:"authors"`
	Abstract        string     `json:"abstract"`
	DOI             string     `json:"doi"`
	PublicationDate *time.Time `json:"publication_date,omitempty"`
	Keywords        []string   `json:"keywords"`
	IsScanned       bool       `json:"is_scanned"`
	MissingFields   []string   `json:"missing_fields"`
}

// DetermineMissingFields identifies which core fields could not be extracted.
func (m *ExtractedMetadata) DetermineMissingFields() {
	var missing []string
	if m.Title == "" {
		missing = append(missing, "title")
	}
	if len(m.Authors) == 0 {
		missing = append(missing, "authors")
	}
	if m.Abstract == "" {
		missing = append(missing, "abstract")
	}
	if m.DOI == "" {
		missing = append(missing, "doi")
	}
	if m.PublicationDate == nil {
		missing = append(missing, "publication_date")
	}
	if len(m.Keywords) == 0 {
		missing = append(missing, "keywords")
	}
	m.MissingFields = missing
}

// ExtractedText represents the raw text content extracted from the PDF, suited for indexing and embeddings.
type ExtractedText struct {
	FullText   string `json:"full_text"`
	TotalPages int    `json:"total_pages"`
	IsScanned  bool   `json:"is_scanned"`
}

// Extractor defines the interface for PDF text and metadata extraction.
type Extractor interface {
	ExtractMetadata(ctx context.Context, r io.Reader, opts Options) (ExtractedMetadata, error)
	ExtractText(ctx context.Context, r io.Reader, opts Options) (ExtractedText, error)
}

