package application

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/document/domain"
	"github.com/paperstacks.io/paperstacks/internal/document/repository/memory"
	paperApp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	paperMemory "github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
)

func setupTestService(t *testing.T) (*DocumentService, *memory.Repository, *memory.Storage, *paperMemory.Repository) {
	repo := memory.NewRepository()
	storage := memory.NewStorage()
	paperRepo := paperMemory.NewRepository()
	paperService := paperApp.NewPaperService(paperRepo)
	service := NewDocumentService(repo, storage, paperService)
	return service, repo, storage, paperRepo
}

func TestUploadSuccess(t *testing.T) {
	service, repo, _, paperRepo := setupTestService(t)

	ctx := context.Background()

	_, err := paperRepo.Save(ctx, paperDomain.Paper{
		UUID: "paper-uuid-xyz",
	})
	if err != nil {
		t.Fatalf("failed to save paper: %v", err)
	}

	pdfContent := []byte("%PDF-1.4\ncontent")
	r := bytes.NewReader(pdfContent)

	doc, err := service.Upload(ctx, "paper-uuid-xyz", "  test_document.pdf  ", "test-user-123", r)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !strings.HasPrefix(doc.Key, "paper/paper-uuid-xyz/") || !strings.HasSuffix(doc.Key, ".pdf") {
		t.Errorf("expected doc.Key to have format 'paper/paper-uuid-xyz/{uuid}.pdf', got: %s", doc.Key)
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
	service, _, _, paperRepo := setupTestService(t)

	ctx := context.Background()

	_, _ = paperRepo.Save(ctx, paperDomain.Paper{UUID: "paper-uuid-xyz"})

	invalidContent := []byte("NOTAPDF-1.4\ncontent")
	r := bytes.NewReader(invalidContent)

	_, err := service.Upload(ctx, "paper-uuid-xyz", "test.pdf", "test-user-123", r)
	if !errors.Is(err, domain.ErrInvalidFileType) {
		t.Errorf("expected ErrInvalidFileType, got: %v", err)
	}
}

func TestUploadActualStreamExceedsLimit(t *testing.T) {
	service, _, _, paperRepo := setupTestService(t)

	ctx := context.Background()

	_, _ = paperRepo.Save(ctx, paperDomain.Paper{UUID: "paper-uuid-xyz"})

	const sizeExceedingLimit = 11 * 1024 * 1024
	header := []byte("%PDF-1.4\n")
	infiniteReader := io.MultiReader(
		bytes.NewReader(header),
		io.LimitReader(infiniteZeroReader{}, sizeExceedingLimit),
	)

	_, err := service.Upload(ctx, "paper-uuid-xyz", "test.pdf", "test-user-123", infiniteReader)
	if !errors.Is(err, domain.ErrFileSizeExceeded) {
		t.Errorf("expected ErrFileSizeExceeded, got: %v", err)
	}
}

func TestUploadPaperDoesNotExist(t *testing.T) {
	service, _, _, _ := setupTestService(t)

	ctx := context.Background()

	pdfContent := []byte("%PDF-1.4\ncontent")
	r := bytes.NewReader(pdfContent)

	_, err := service.Upload(ctx, "non-existent-paper-uuid", "test.pdf", "test-user-123", r)
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
