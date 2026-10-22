package domain

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	ErrFileSizeExceeded = errors.New("file size exceeds maximum limit")
	ErrInvalidFileType  = errors.New("invalid file type: only valid PDFs allowed")
)

const (
	MiB                = 1024 * 1024
	PDFMaxFileSize     = 10 * MiB
	PDFContentType     = "application/pdf"
	pdfSignature       = "%PDF-"
	pdfSniffBufferSize = 512
)

type UploadPDFOptions struct {
	PaperUUID string
	FileName  string
	UserID    string
	File      io.Reader
	Size      int64
}

func ValidatePDF(r io.Reader) (io.Reader, int64, error) {
	buf := make([]byte, pdfSniffBufferSize)

	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, 0, fmt.Errorf("failed to read file header: %w", err)
	}

	if n < len(pdfSignature) ||
		string(buf[:len(pdfSignature)]) != pdfSignature ||
		http.DetectContentType(buf[:n]) != PDFContentType {
		return nil, 0, ErrInvalidFileType
	}

	fullReader := io.MultiReader(bytes.NewReader(buf[:n]), r)

	pdfData, err := io.ReadAll(io.LimitReader(fullReader, PDFMaxFileSize+1))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read PDF: %w", err)
	}

	if int64(len(pdfData)) > PDFMaxFileSize {
		return nil, 0, ErrFileSizeExceeded
	}

	return bytes.NewReader(pdfData), int64(len(pdfData)), nil
}

type PDF struct {
	Key         string
	PaperUUID   string
	UserID      string
	FileName    string
	ContentType string
	Size        int64
}

func (d PDF) Validate() error {
	if d.PaperUUID == "" {
		return errors.New("paper UUID is required")
	}

	if d.FileName == "" {
		return errors.New("file name is required")
	}

	return nil
}
