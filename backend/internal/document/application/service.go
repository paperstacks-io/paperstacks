package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/common/objectstorage"
	"github.com/paperstacks.io/paperstacks/internal/document/domain"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

type PaperGetter interface {
	GetByUUID(ctx context.Context, uuid string) (paperDomain.Paper, error)
}

type DocumentService struct {
	repo        domain.Repository
	storage     objectstorage.Store
	paperGetter PaperGetter
}

func NewDocumentService(
	repo domain.Repository,
	storage objectstorage.Store,
	paperGetter PaperGetter,
) *DocumentService {
	return &DocumentService{
		repo:        repo,
		storage:     storage,
		paperGetter: paperGetter,
	}
}

func (s *DocumentService) Upload(
	ctx context.Context,
	opts domain.UploadOptions,
) (domain.Document, error) {
	if _, err := s.paperGetter.GetByUUID(ctx, opts.PaperUUID); err != nil {
		return domain.Document{}, err
	}

	if opts.Size > opts.Config.MaxFileSize {
		return domain.Document{}, domain.ErrFileSizeExceeded
	}

	pdfReader, pdfSize, err := domain.ValidatePDF(opts.File, opts.Config)
	if err != nil {
		return domain.Document{}, err
	}

	trimmedFileName := strings.TrimSpace(opts.FileName)
	storageKey := fmt.Sprintf("pdf/%s/main.pdf", opts.PaperUUID)

	_, err = s.storage.Put(ctx, objectstorage.PutObjectInput{
		Key:         storageKey,
		Body:        pdfReader,
		Size:        pdfSize,
		ContentType: opts.Config.ContentType,
	})
	if err != nil {
		return domain.Document{}, fmt.Errorf("failed to store physical file: %w", err)
	}

	doc := domain.Document{
		Key:         storageKey,
		UserID:      opts.UserID,
		PaperUUID:   opts.PaperUUID,
		FileName:    trimmedFileName,
		ContentType: opts.Config.ContentType,
		Size:        pdfSize,
	}

	savedDoc, err := s.repo.Save(ctx, doc)
	if err != nil {
		return domain.Document{}, fmt.Errorf("failed to save document metadata: %w", err)
	}

	return savedDoc, nil
}
