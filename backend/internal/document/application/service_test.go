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
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

func TestUploadSuccess(t *testing.T) {
	repo := memory.NewRepository()
	storage := memory.NewStorage()
	service := NewDocumentService(repo, storage)

	ctx := context.Background()
	user := userDomain.User{ExternalID: "test-user-123"}
	docMetadata := domain.Document{
		PaperUUID: "paper-uuid-xyz",
		FileName:  "  test_document.pdf  ",
		Size:      15,
	}

	pdfContent := []byte("%PDF-1.4\ncontent")
	r := bytes.NewReader(pdfContent)

	doc, err := service.Upload(ctx, docMetadata, user, r)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if doc.UUID == "" {
		t.Error("expected doc.UUID to be generated, got empty string")
	}
	if doc.UploaderUUID != "test-user-123" {
		t.Errorf("expected UploaderUUID to be 'test-user-123', got: %s", doc.UploaderUUID)
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
	if !strings.HasPrefix(doc.StorageURI, "mem://documents/") {
		t.Errorf("expected StorageURI to start with 'mem://documents/', got: %s", doc.StorageURI)
	}

	savedDoc, err := repo.Save(ctx, doc)
	if err != nil {
		t.Errorf("expected document to be queryable in repository: %v", err)
	}
	if savedDoc.UUID != doc.UUID {
		t.Errorf("expected saved document UUID to be %s, got %s", doc.UUID, savedDoc.UUID)
	}
}

func TestUploadExceedsDeclaredSizeLimit(t *testing.T) {
	repo := memory.NewRepository()
	storage := memory.NewStorage()
	service := NewDocumentService(repo, storage)

	ctx := context.Background()
	user := userDomain.User{ExternalID: "test-user-123"}
	docMetadata := domain.Document{
		PaperUUID: "paper-uuid-xyz",
		FileName:  "test.pdf",
		Size:      11 * 1024 * 1024,
	}

	pdfContent := []byte("%PDF-1.4\ncontent")
	r := bytes.NewReader(pdfContent)

	_, err := service.Upload(ctx, docMetadata, user, r)
	if !errors.Is(err, domain.ErrFileSizeExceeded) {
		t.Errorf("expected ErrFileSizeExceeded, got: %v", err)
	}
}

func TestUploadInvalidMagicBytes(t *testing.T) {
	repo := memory.NewRepository()
	storage := memory.NewStorage()
	service := NewDocumentService(repo, storage)

	ctx := context.Background()
	user := userDomain.User{ExternalID: "test-user-123"}
	docMetadata := domain.Document{
		PaperUUID: "paper-uuid-xyz",
		FileName:  "test.pdf",
		Size:      100,
	}

	invalidContent := []byte("NOTAPDF-1.4\ncontent")
	r := bytes.NewReader(invalidContent)

	_, err := service.Upload(ctx, docMetadata, user, r)
	if !errors.Is(err, domain.ErrInvalidFileType) {
		t.Errorf("expected ErrInvalidFileType, got: %v", err)
	}
}

func TestUploadActualStreamExceedsLimit(t *testing.T) {
	repo := memory.NewRepository()
	storage := memory.NewStorage()
	service := NewDocumentService(repo, storage)

	ctx := context.Background()
	user := userDomain.User{ExternalID: "test-user-123"}
	docMetadata := domain.Document{
		PaperUUID: "paper-uuid-xyz",
		FileName:  "test.pdf",
		Size:      100,
	}

	header := []byte("%PDF-1.4\n")
	infiniteReader := io.MultiReader(
		bytes.NewReader(header),
		io.LimitReader(infiniteZeroReader{}, 11*1024*1024),
	)

	_, err := service.Upload(ctx, docMetadata, user, infiniteReader)
	if !errors.Is(err, domain.ErrFileSizeExceeded) {
		t.Errorf("expected ErrFileSizeExceeded, got: %v", err)
	}
}

type infiniteZeroReader struct{}

func (infiniteZeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}
