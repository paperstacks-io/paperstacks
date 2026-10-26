package application

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/common/objectstorage"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
	paperMemory "github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
)

type storedPDF struct {
	body     []byte
	metadata map[string]string
}

type mockObjectStorage struct {
	objects map[string]storedPDF
	err     error
}

func (m mockObjectStorage) Put(ctx context.Context, input objectstorage.PutObjectInput) (objectstorage.ObjectInfo, error) {
	if m.err != nil {
		return objectstorage.ObjectInfo{}, m.err
	}
	data, err := io.ReadAll(input.Body)
	if err != nil {
		return objectstorage.ObjectInfo{}, err
	}
	m.objects[input.Key] = storedPDF{body: data, metadata: input.Metadata}
	return objectstorage.ObjectInfo{
		Key:         input.Key,
		Size:        input.Size,
		ContentType: input.ContentType,
		Metadata:    input.Metadata,
	}, nil
}

func (m mockObjectStorage) Get(ctx context.Context, key string) (*objectstorage.Object, error) {
	if m.err != nil {
		return nil, m.err
	}
	data, ok := m.objects[key]
	if !ok {
		return nil, objectstorage.ErrObjectNotFound
	}
	return &objectstorage.Object{Body: io.NopCloser(bytes.NewReader(data.body)), Info: objectstorage.ObjectInfo{Metadata: data.metadata}}, nil
}

func (m mockObjectStorage) Delete(ctx context.Context, key string) error {
	return nil
}

func (m mockObjectStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := m.objects[key]
	return ok, m.err
}

func setupPDFTestService(t *testing.T) (*PaperService, *paperMemory.Repository) {
	t.Helper()
	paperRepo := paperMemory.NewRepository()
	return NewPaperService(paperRepo, mockObjectStorage{objects: make(map[string]storedPDF)}), paperRepo
}

func TestUploadPDFSuccess(t *testing.T) {
	service, paperRepo := setupPDFTestService(t)
	const paperUUID = "a4b065f1-1b88-4f50-a7fe-1177f3489fcf"

	ctx := context.Background()

	_, err := paperRepo.Save(ctx, domain.Paper{
		UUID:  paperUUID,
		DOI:   "10.1000/test-paper",
		Title: "Test Paper",
	})
	if err != nil {
		t.Fatalf("failed to save paper: %v", err)
	}

	pdfContent := []byte("%PDF-1.4\ncontent")
	r := bytes.NewReader(pdfContent)

	opts := domain.UploadPDFOptions{
		PaperUUID: paperUUID,
		FileName:  "  test_document.pdf  ",
		UserID:    "test-user-123",
		File:      r,
		Size:      int64(len(pdfContent)),
	}

	doc, err := service.UploadPDF(ctx, opts)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if doc.Key != "pdf/"+paperUUID+".pdf" {
		t.Errorf("unexpected storage key: %s", doc.Key)
	}
	if doc.UserID != "test-user-123" {
		t.Errorf("expected UploaderUUID to be 'test-user-123', got: %s", doc.UserID)
	}
	if doc.PaperUUID != paperUUID {
		t.Errorf("expected PaperUUID to be %q, got: %s", paperUUID, doc.PaperUUID)
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
	paper, err := paperRepo.GetByUUID(ctx, paperUUID)
	if err != nil {
		t.Fatal(err)
	}
	if paper.PDFURL != doc.Key {
		t.Errorf("stored PDFURL = %q, want %q", paper.PDFURL, doc.Key)
	}

	object, err := service.DownloadPDF(ctx, doc.Key)
	if err != nil {
		t.Fatal(err)
	}
	defer object.Body.Close()
	if got := object.Info.Metadata["filename"]; got != "test_document.pdf" {
		t.Fatalf("downloaded filename = %q, want test_document.pdf", got)
	}

}

func TestUploadPDFInvalidMagicBytes(t *testing.T) {
	service, paperRepo := setupPDFTestService(t)

	ctx := context.Background()

	_, _ = paperRepo.Save(ctx, domain.Paper{UUID: "paper-uuid-xyz"})

	invalidContent := []byte("NOTAPDF-1.4\ncontent")
	r := bytes.NewReader(invalidContent)

	opts := domain.UploadPDFOptions{
		PaperUUID: "paper-uuid-xyz",
		FileName:  "test.pdf",
		UserID:    "test-user-123",
		File:      r,
		Size:      int64(len(invalidContent)),
	}

	_, err := service.UploadPDF(ctx, opts)
	if !errors.Is(err, domain.ErrInvalidFileType) {
		t.Errorf("expected ErrInvalidFileType, got: %v", err)
	}
}

func TestUploadPDFActualStreamExceedsLimit(t *testing.T) {
	service, paperRepo := setupPDFTestService(t)

	ctx := context.Background()

	_, _ = paperRepo.Save(ctx, domain.Paper{UUID: "paper-uuid-xyz"})

	const sizeExceedingLimit = 11 * 1024 * 1024
	header := []byte("%PDF-1.4\n")
	infiniteReader := io.MultiReader(
		bytes.NewReader(header),
		io.LimitReader(infiniteZeroReader{}, sizeExceedingLimit),
	)

	opts := domain.UploadPDFOptions{
		PaperUUID: "paper-uuid-xyz",
		FileName:  "test.pdf",
		UserID:    "test-user-123",
		File:      infiniteReader,
		Size:      0,
	}

	_, err := service.UploadPDF(ctx, opts)
	if !errors.Is(err, domain.ErrFileSizeExceeded) {
		t.Errorf("expected ErrFileSizeExceeded, got: %v", err)
	}
}

func TestUploadPDFPaperDoesNotExist(t *testing.T) {
	service, _ := setupPDFTestService(t)

	ctx := context.Background()

	pdfContent := []byte("%PDF-1.4\ncontent")
	r := bytes.NewReader(pdfContent)

	opts := domain.UploadPDFOptions{
		PaperUUID: "non-existent-paper-uuid",
		FileName:  "test.pdf",
		UserID:    "test-user-123",
		File:      r,
		Size:      int64(len(pdfContent)),
	}

	_, err := service.UploadPDF(ctx, opts)
	if !errors.Is(err, domain.ErrPaperNotFound) {
		t.Errorf("expected ErrPaperNotFound, got: %v", err)
	}
}

type infiniteZeroReader struct{}

func TestPDFsAreStoredByPaperUUID(t *testing.T) {
	service, repo := setupPDFTestService(t)
	ctx := context.Background()
	paperUUIDs := []string{"a4b065f1-1b88-4f50-a7fe-1177f3489fcf", "a4b065f1-1b88-4f50-a7fe-1177f3489fd0"}
	for i, paperUUID := range paperUUIDs {
		if _, err := repo.Save(ctx, domain.Paper{UUID: paperUUID, DOI: "10.1000/same-paper", Title: "Test Paper"}); err != nil {
			t.Fatal(err)
		}
		if exists, err := service.HasPDF(ctx, paperUUID); err != nil || exists {
			t.Fatalf("HasPDF before upload = %v, %v", exists, err)
		}
		content := "%PDF-1.4\n" + paperUUID
		doc, err := service.UploadPDF(ctx, domain.UploadPDFOptions{
			PaperUUID: paperUUID, UserID: []string{"user-one", "user-two"}[i],
			FileName: "main.pdf", File: bytes.NewBufferString(content),
		})
		if err != nil {
			t.Fatal(err)
		}
		if doc.Key != "pdf/"+paperUUID+".pdf" {
			t.Fatalf("unexpected storage key: %s", doc.Key)
		}
	}
	for _, paperUUID := range paperUUIDs {
		if exists, err := service.HasPDF(ctx, paperUUID); err != nil || !exists {
			t.Fatalf("HasPDF after upload = %v, %v", exists, err)
		}
		object, err := service.DownloadPDF(ctx, "pdf/"+paperUUID+".pdf")
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(object.Body)
		object.Body.Close()
		if err != nil || string(content) != "%PDF-1.4\n"+paperUUID {
			t.Fatalf("downloaded content = %q, error = %v", content, err)
		}
	}
}

func TestPDFStorageErrors(t *testing.T) {
	ctx := context.Background()
	repo := paperMemory.NewRepository()
	if _, err := repo.Save(ctx, domain.Paper{UUID: "paper-one"}); err != nil {
		t.Fatal(err)
	}
	wantErr := errors.New("storage unavailable")
	service := NewPaperService(repo, mockObjectStorage{err: wantErr})
	if _, err := service.UploadPDF(ctx, domain.UploadPDFOptions{
		PaperUUID: "paper-one", FileName: "main.pdf",
		File: bytes.NewBufferString("%PDF-1.4\ncontent"),
	}); !errors.Is(err, wantErr) {
		t.Fatalf("UploadPDF error = %v, want %v", err, wantErr)
	}
	if _, err := service.DownloadPDF(ctx, "paper-one"); !errors.Is(err, wantErr) {
		t.Fatalf("DownloadPDF error = %v, want %v", err, wantErr)
	}
	if _, err := service.HasPDF(ctx, "paper-one"); !errors.Is(err, wantErr) {
		t.Fatalf("HasPDF error = %v, want %v", err, wantErr)
	}
}

func (infiniteZeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}
