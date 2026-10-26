package application

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/pdfextractor/domain"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

// ExtractMetadata parses the initial pages of a PDF to extract bibliographic metadata.
func (s *Service) ExtractMetadata(ctx context.Context, r io.Reader, opts domain.Options) (domain.ExtractedMetadata, error) {
	if opts.HeadPages <= 0 {
		opts.HeadPages = 2 // default to first 2 pages for metadata
	}

	rawText, isScanned, err := s.runPdfToText(ctx, r, opts)
	if err != nil {
		return domain.ExtractedMetadata{}, err
	}

	meta := parseMetadataHeuristics(rawText)
	meta.IsScanned = isScanned
	meta.DetermineMissingFields()

	return meta, nil
}

// ExtractText extracts textual content from the PDF, suited for downstream full-text indexing or vector embeddings.
func (s *Service) ExtractText(ctx context.Context, r io.Reader, opts domain.Options) (domain.ExtractedText, error) {
	rawText, isScanned, err := s.runPdfToText(ctx, r, opts)
	if err != nil {
		return domain.ExtractedText{}, err
	}

	return domain.ExtractedText{
		FullText:  rawText,
		IsScanned: isScanned,
	}, nil
}

func (s *Service) runPdfToText(ctx context.Context, r io.Reader, opts domain.Options) (string, bool, error) {
	if _, err := exec.LookPath("pdftotext"); err != nil {
		return "", false, domain.ErrPDFToTextNotFound
	}
	if r == nil {
		return "", false, domain.ErrEmptyPDF
	}

	// Read first byte to ensure reader is not empty without loading entire stream into memory
	firstByte := make([]byte, 1)
	n, err := r.Read(firstByte)
	if err != nil || n == 0 {
		return "", false, domain.ErrEmptyPDF
	}
	stream := io.MultiReader(bytes.NewReader(firstByte[:n]), r)

	args := []string{}
	if !opts.ExtractAllPages {
		if opts.HeadPages > 0 {
			args = append(args, "-l", strconv.Itoa(opts.HeadPages))
		}
	}
	// "-" as input reads from stdin, second "-" outputs to stdout
	args = append(args, "-layout", "-", "-")

	cmdCtx := ctx
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		cmdCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(cmdCtx, "pdftotext", args...)
	cmd.Stdin = stream
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(cmdCtx.Err(), context.DeadlineExceeded) {
			return "", false, fmt.Errorf("%w: timeout exceeded", domain.ErrExtractionFailed)
		}
		return "", false, fmt.Errorf("%w: %v, stderr: %s", domain.ErrExtractionFailed, err, stderr.String())
	}

	text := stdout.String()
	isScanned := len(strings.TrimSpace(text)) < 100

	return text, isScanned, nil
}

func parseMetadataHeuristics(rawText string) domain.ExtractedMetadata {
	res := domain.ExtractedMetadata{}

	// DOI regex matching standard DOI patterns (e.g. 10.1145/1234567.1234568)
	doiRegex := regexp.MustCompile(`10\.\d{4,9}/[-._;()/:A-Za-z0-9]+`)
	if match := doiRegex.FindString(rawText); match != "" {
		res.DOI = strings.TrimRight(match, ".,;")
	}

	// Abstract regex matching "Abstract" block followed by intro or blank line
	absRegex := regexp.MustCompile(`(?is)(?:abstract|a b s t r a c t)[:\s]+(.*?)(?:\n\s*\n|introduction|1\. introduction)`)
	if match := absRegex.FindStringSubmatch(rawText); len(match) > 1 {
		res.Abstract = strings.TrimSpace(match[1])
	}

	// Keywords regex
	keywordsRegex := regexp.MustCompile(`(?i)(?:keywords|index terms|key words)[:\s—-]+([^\n\r]+)`)
	if match := keywordsRegex.FindStringSubmatch(rawText); len(match) > 1 {
		rawKeywords := match[1]
		parts := strings.Split(rawKeywords, ",")
		if len(parts) == 1 {
			parts = strings.Split(rawKeywords, ";")
		}
		for _, part := range parts {
			cleaned := strings.TrimSpace(part)
			cleaned = strings.Trim(cleaned, ".,;")
			if cleaned != "" {
				res.Keywords = append(res.Keywords, cleaned)
			}
		}
	}

	return res
}

