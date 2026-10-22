package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/common/objectstorage"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

func (s *PaperService) UploadPDF(
	ctx context.Context,
	opts domain.UploadPDFOptions,
) (domain.PDF, error) {
	if _, err := s.GetByUUID(ctx, opts.PaperUUID); err != nil {
		return domain.PDF{}, err
	}

	if opts.Size > domain.PDFMaxFileSize {
		return domain.PDF{}, domain.ErrFileSizeExceeded
	}

	pdfReader, pdfSize, err := domain.ValidatePDF(opts.File)
	if err != nil {
		return domain.PDF{}, err
	}

	trimmedFileName := strings.TrimSpace(opts.FileName)
	storageKey := fmt.Sprintf("pdf/%s/main.pdf", opts.PaperUUID)

	_, err = s.storage.Put(ctx, objectstorage.PutObjectInput{
		Key:         storageKey,
		Body:        pdfReader,
		Size:        pdfSize,
		ContentType: domain.PDFContentType,
	})
	if err != nil {
		return domain.PDF{}, fmt.Errorf("failed to store physical file: %w", err)
	}

	return domain.PDF{
		Key:         storageKey,
		UserID:      opts.UserID,
		PaperUUID:   opts.PaperUUID,
		FileName:    trimmedFileName,
		ContentType: domain.PDFContentType,
		Size:        pdfSize,
	}, nil
}

func (s *PaperService) DownloadPDF(ctx context.Context, paperUUID string) (*objectstorage.Object, error) {
	key := fmt.Sprintf("pdf/%s/main.pdf", paperUUID)

	object, err := s.storage.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to download PDF: %w", err)
	}

	return object, nil
}

func (s *PaperService) HasPDF(ctx context.Context, paperUUID string) (bool, error) {
	key := fmt.Sprintf("pdf/%s/main.pdf", paperUUID)

	return s.storage.Exists(ctx, key)
}
