package application

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"uuid"

	"github.com/paperstacks.io/paperstacks/internal/common/objectstorage"
	"github.com/paperstacks.io/paperstacks/internal/document/domain"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

var (
	ErrFileSizeExceeded = errors.New("file size exceeds maximum limit")
	ErrInvalidFileType  = errors.New("invalid file type: only valid PDFs allowed")
)

type PaperGetter interface {
	GetByUUID(ctx context.Context, uuid string) (paperDomain.Paper, error)
}

type DocumentService struct {
	repo        domain.Repository
	storage     objectstorage.Store
	paperGetter PaperGetter
}

func NewDocumentService(repo domain.Repository, storage objectstorage.Store, paperGetter PaperGetter) *DocumentService {
	return &DocumentService{
		repo:        repo,
		storage:     storage,
		paperGetter: paperGetter,
	}
}

func (s *DocumentService) Upload(
	ctx context.Context,
	paperUUID string,
	fileName string,
	userID string,
	r io.Reader,
	size int64,
) (domain.Document, error) {
	if _, err := s.paperGetter.GetByUUID(ctx, paperUUID); err != nil {
		return domain.Document{}, err
	}

	const (
		maxFileSize        = 10 * 1024 * 1024
		pdfSignature       = "%PDF-"
		pdfSignatureLength = len(pdfSignature)
		documentType       = "application/pdf"
		sniffBufferSize    = 512
	)

	if size > maxFileSize {
		return domain.Document{}, ErrFileSizeExceeded
	}

	buf := make([]byte, sniffBufferSize)
	n, err := io.ReadFull(io.LimitReader(r, sniffBufferSize), buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return domain.Document{}, fmt.Errorf("failed to read file header: %w", err)
	}

	if n < pdfSignatureLength {
		return domain.Document{}, ErrInvalidFileType
	}

	if string(buf[:pdfSignatureLength]) != pdfSignature {
		return domain.Document{}, ErrInvalidFileType
	}

	detectedType := http.DetectContentType(buf[:n])
	if detectedType != documentType {
		return domain.Document{}, ErrInvalidFileType
	}

	fullReader := io.MultiReader(bytes.NewReader(buf[:n]), r)

	trimmedFileName := strings.TrimSpace(fileName)
	docUUID := uuid.New().String()
	storageKey := fmt.Sprintf("paper/%s/%s.pdf", paperUUID, docUUID)

	_, err = s.storage.Put(ctx, objectstorage.PutObjectInput{
		Key:         storageKey,
		Body:        fullReader,
		Size:        size,
		ContentType: documentType,
	})
	if err != nil {
		return domain.Document{}, fmt.Errorf("failed to store physical file: %w", err)
	}

	doc := domain.Document{
		Key:         storageKey,
		UserID:      userID,
		PaperUUID:   paperUUID,
		FileName:    trimmedFileName,
		ContentType: documentType,
		Size:        size,
	}

	savedDoc, err := s.repo.Save(ctx, doc)
	if err != nil {
		return domain.Document{}, fmt.Errorf("failed to save document metadata: %w", err)
	}

	return savedDoc, nil
}
