package application

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/common/objectstorage"
	"github.com/paperstacks.io/paperstacks/internal/document/domain"
	"github.com/paperstacks.io/paperstacks/internal/document/repository/memory"
	paperApp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	paperMemory "github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
)

type mockObjectStorage struct{}

func (m mockObjectStorage) Put(ctx context.Context, input objectstorage.PutObjectInput) (objectstorage.ObjectInfo, error) {
	return objectstorage.ObjectInfo{
		Key:         input.Key,
		Size:        input.Size,
		ContentType: input.ContentType,
	}, nil
}

func (m mockObjectStorage) Get(ctx context.Context, key string) (*objectstorage.Object, error) {
	return nil, nil
}

func (m mockObjectStorage) Delete(ctx context.Context, key string) error {
	return nil
}

func (m mockObjectStorage) Exists(ctx context.Context, key string) (bool, error) {
	return true, nil
}

func setupTestService(t *testing.T) (*DocumentService, *memory.Repository, *paperMemory.Repository) {
	repo := memory.NewRepository()
	storage := mockObjectStorage{}
	paperRepo := paperMemory.NewRepository()
	paperService := paperApp.NewPaperService(paperRepo)
	service := NewDocumentService(repo, storage, paperService)
	return service, repo, paperRepo
}

func testPDFConfig() domain.PDFConfiguration {
	return domain.PDFConfiguration{
		MaxFileSize:     10 * 1024 * 1024,
		Signature:       "%PDF-",
		ContentType:     "application/pdf",
		SniffBufferSize: 512,
	}
}

func TestUploadSuccess(t *testing.T) {
	service, repo, paperRepo := setupTestService(t)

	ctx := context.Background()

	_, err := paperRepo.Save(ctx, paperDomain.Paper{
		UUID: "paper-uuid-xyz",
	})
	if err != nil {
		t.Fatalf("failed to save paper: %v", err)
	}

	pdfContent := []byte("%PDF-1.4\ncontent")
	r := bytes.NewReader(pdfContent)

	opts := domain.UploadOptions{
		PaperUUID: "paper-uuid-xyz",
		FileName:  "  test_document.pdf  ",
		UserID:    "test-user-123",
		File:      r,
		Size:      int64(len(pdfContent)),
		Config:    testPDFConfig(),
	}

	doc, err := service.Upload(ctx, opts)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if doc.Key != "pdf/paper-uuid-xyz/main.pdf" {
		t.Errorf("expected doc.Key to be 'pdf/paper-uuid-xyz/main.pdf', got: %s", doc.Key)
	}
	if doc.UserID != "test-user-123" {
		t.Errorf("expected UploaderUUID to be 'test-user-123', got: %s", doc.UserID)
	}
	if doc.PaperUUID != "paper-uuid-xyz" {
		t.Errorf("expected PaperUUID to be 'paper-uuid-xyz', got: %s", doc.PaperUUID)
	}
	if doc.FileName != "test_document.pdf" {
		t.Errorf("expected FileName to be trimmed to 'test_document.pdf', got: %s", doc.FileName)
	}
	if doc.ContentType != "application/pdf" {
		t.Errorf("expected ContentType to be 'application/pdf', got: %s", doc.ContentType)
	}
	if doc.Size != int64(len(pdfContent)) {
		t.Errorf("expected Size to be %d, got: %d", len(pdfContent), doc.Size)
	}

	savedDoc, err := repo.Save(ctx, doc)
	if err != nil {
		t.Errorf("expected document to be queryable in repository: %v", err)
	}
	if savedDoc.Key != doc.Key {
		t.Errorf("expected saved document UUID to be %s, got %s", doc.Key, savedDoc.Key)
	}
}

func TestUploadInvalidMagicBytes(t *testing.T) {
	service, _, paperRepo := setupTestService(t)

	ctx := context.Background()

	_, _ = paperRepo.Save(ctx, paperDomain.Paper{UUID: "paper-uuid-xyz"})

	invalidContent := []byte("NOTAPDF-1.4\ncontent")
	r := bytes.NewReader(invalidContent)

	opts := domain.UploadOptions{
		PaperUUID: "paper-uuid-xyz",
		FileName:  "test.pdf",
		UserID:    "test-user-123",
		File:      r,
		Size:      int64(len(invalidContent)),
		Config:    testPDFConfig(),
	}

	_, err := service.Upload(ctx, opts)
	if !errors.Is(err, domain.ErrInvalidFileType) {
		t.Errorf("expected ErrInvalidFileType, got: %v", err)
	}
}

func TestUploadActualStreamExceedsLimit(t *testing.T) {
	service, _, paperRepo := setupTestService(t)

	ctx := context.Background()

	_, _ = paperRepo.Save(ctx, paperDomain.Paper{UUID: "paper-uuid-xyz"})

	const sizeExceedingLimit = 11 * 1024 * 1024
	header := []byte("%PDF-1.4\n")
	infiniteReader := io.MultiReader(
		bytes.NewReader(header),
		io.LimitReader(infiniteZeroReader{}, sizeExceedingLimit),
	)

	opts := domain.UploadOptions{
		PaperUUID: "paper-uuid-xyz",
		FileName:  "test.pdf",
		UserID:    "test-user-123",
		File:      infiniteReader,
		Size:      0,
		Config:    testPDFConfig(),
	}

	_, err := service.Upload(ctx, opts)
	if !errors.Is(err, domain.ErrFileSizeExceeded) {
		t.Errorf("expected ErrFileSizeExceeded, got: %v", err)
	}
}

func TestUploadPaperDoesNotExist(t *testing.T) {
	service, _, _ := setupTestService(t)

	ctx := context.Background()

	pdfContent := []byte("%PDF-1.4\ncontent")
	r := bytes.NewReader(pdfContent)

	opts := domain.UploadOptions{
		PaperUUID: "non-existent-paper-uuid",
		FileName:  "test.pdf",
		UserID:    "test-user-123",
		File:      r,
		Size:      int64(len(pdfContent)),
		Config:    testPDFConfig(),
	}

	_, err := service.Upload(ctx, opts)
	if !errors.Is(err, paperDomain.ErrPaperNotFound) {
		t.Errorf("expected ErrPaperNotFound, got: %v", err)
	}
}

type infiniteZeroReader struct{}

func (infiniteZeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}
