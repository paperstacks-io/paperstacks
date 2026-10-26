package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/common/objectstorage"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

// UploadPDF uploads a PDF to object storage and returns the assigned key.
func (s *PaperService) UploadPDF(
	ctx context.Context,
	opts domain.UploadPDFOptions,
) (string, error) {
	paper, err := s.GetByUUID(ctx, opts.PaperUUID)
	if err != nil {
		return "", err
	}

	if opts.Size > domain.PDFMaxFileSize {
		return "", domain.ErrFileSizeExceeded
	}

	pdfReader, pdfSize, err := domain.ValidatePDF(opts.File)
	if err != nil {
		return "", err
	}

	storageKey := fmt.Sprintf("pdf/%s.pdf", opts.PaperUUID)

	_, err = s.storage.Put(ctx, objectstorage.PutObjectInput{
		Key:         storageKey,
		Body:        pdfReader,
		Size:        pdfSize,
		ContentType: domain.PDFContentType,
		Metadata:    map[string]string{"filename": opts.FileName},
	})
	if err != nil {
		return "", fmt.Errorf("failed to store physical file: %w", err)
	}

	paper.PDFStorageKey = storageKey
	err = s.Update(ctx, paper.UUID, paper)
	if err != nil {
		return "", fmt.Errorf("failed to update paper PDF key: %w", err)
	}

	return storageKey, nil
}

// OpenPDF opens a PDF from object storage. The caller must close the returned PDFFile.Body.
func (s *PaperService) OpenPDF(ctx context.Context, storageKey string) (*domain.PDFFile, error) {
	object, err := s.storage.Get(ctx, storageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}

	fileName := object.Info.Metadata["filename"]
	if fileName == "" {
		fileName = strings.TrimPrefix(storageKey, "pdf/")
	}

	return &domain.PDFFile{
		Body:        object.Body,
		FileName:    fileName,
		ContentType: object.Info.ContentType,
		Size:        object.Info.Size,
	}, nil
}

func (s *PaperService) PDFExists(ctx context.Context, storageKey string) (bool, error) {
	return s.storage.Exists(ctx, storageKey)
}
