package domain

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	ErrFileSizeExceeded = errors.New("file size exceeds maximum limit")
	ErrInvalidFileType  = errors.New("invalid file type: only valid PDFs allowed")
)

type Repository interface {
	Save(ctx context.Context, doc Document) (Document, error)
}

type PDFConfiguration struct {
	MaxFileSize     int64
	Signature       string
	ContentType     string
	SniffBufferSize int
}

type UploadOptions struct {
	PaperUUID string
	FileName  string
	UserID    string
	File      io.Reader
	Size      int64
	Config    PDFConfiguration
}

func ValidatePDF(r io.Reader, cfg PDFConfiguration) (io.Reader, int64, error) {
	if cfg.MaxFileSize <= 0 ||
		len(cfg.Signature) == 0 ||
		cfg.SniffBufferSize < len(cfg.Signature) {
		return nil, 0, fmt.Errorf("invalid PDF configuration")
	}

	buf := make([]byte, cfg.SniffBufferSize)

	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, 0, fmt.Errorf("failed to read file header: %w", err)
	}

	if n < len(cfg.Signature) ||
		string(buf[:len(cfg.Signature)]) != cfg.Signature ||
		http.DetectContentType(buf[:n]) != cfg.ContentType {
		return nil, 0, ErrInvalidFileType
	}

	fullReader := io.MultiReader(bytes.NewReader(buf[:n]), r)

	pdfData, err := io.ReadAll(io.LimitReader(fullReader, cfg.MaxFileSize+1))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read PDF: %w", err)
	}

	if int64(len(pdfData)) > cfg.MaxFileSize {
		return nil, 0, ErrFileSizeExceeded
	}

	return bytes.NewReader(pdfData), int64(len(pdfData)), nil
}
