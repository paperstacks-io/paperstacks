package application

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/paperstacks.io/paperstacks/internal/document/domain"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

type PaperService interface {
	GetByUUID(ctx context.Context, uuid string) (paperDomain.Paper, error)
}

type DocumentService struct {
	repo         domain.Repository
	storage      domain.Storage
	paperService PaperService
}

func NewDocumentService(repo domain.Repository, storage domain.Storage, paperService PaperService) *DocumentService {
	return &DocumentService{
		repo:         repo,
		storage:      storage,
		paperService: paperService,
	}
}

type limitCountingReader struct {
	r     io.Reader
	limit int64
	read  int64
}

func (c *limitCountingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.read += int64(n)
	if c.read > c.limit {
		return n, domain.ErrFileSizeExceeded
	}
	return n, err
}

func (s *DocumentService) Upload(
	ctx context.Context,
	paperUUID string,
	fileName string,
	userID string,
	r io.Reader,
) (domain.Document, error) {
	if _, err := s.paperService.GetByUUID(ctx, paperUUID); err != nil {
		return domain.Document{}, err
	}

	const (
		maxFileSize        = 10 * 1024 * 1024
		pdfSignature       = "%PDF-"
		pdfSignatureLength = len(pdfSignature)
		documentType       = "application/pdf"
		sniffBufferSize    = 512
	)

	buf := make([]byte, sniffBufferSize)
	n, err := io.ReadFull(io.LimitReader(r, sniffBufferSize), buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return domain.Document{}, fmt.Errorf("failed to read file header: %w", err)
	}

	if n < pdfSignatureLength {
		return domain.Document{}, domain.ErrInvalidFileType
	}

	if string(buf[:pdfSignatureLength]) != pdfSignature {
		return domain.Document{}, domain.ErrInvalidFileType
	}

	detectedType := http.DetectContentType(buf[:n])
	if detectedType != documentType {
		return domain.Document{}, domain.ErrInvalidFileType
	}

	fullReader := io.MultiReader(bytes.NewReader(buf[:n]), r)
	limitReader := &limitCountingReader{
		r:     fullReader,
		limit: maxFileSize,
	}

	trimmedFileName := strings.TrimSpace(fileName)
	docUUID := uuid.NewString()
	storageKey := fmt.Sprintf("paper/%s/%s.pdf", paperUUID, docUUID)

	err = s.storage.Put(ctx, storageKey, limitReader)
	if err != nil {
		return domain.Document{}, fmt.Errorf("failed to store physical file: %w", err)
	}

	doc := domain.Document{
		Key:         storageKey,
		UserID:      userID,
		PaperUUID:   paperUUID,
		FileName:    trimmedFileName,
		ContentType: "application/pdf",
		Size:        limitReader.read,
	}

	savedDoc, err := s.repo.Save(ctx, doc)
	if err != nil {
		return domain.Document{}, fmt.Errorf("failed to save document metadata: %w", err)
	}

	return savedDoc, nil
}
