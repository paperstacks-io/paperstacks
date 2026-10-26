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
	body []byte
	info objectstorage.ObjectInfo
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
	info := objectstorage.ObjectInfo{
		Key:         input.Key,
		Size:        input.Size,
		ContentType: input.ContentType,
		Metadata:    input.Metadata,
	}
	m.objects[input.Key] = storedPDF{body: data, info: info}
	return info, nil
}

func (m mockObjectStorage) Get(ctx context.Context, key string) (*objectstorage.Object, error) {
	if m.err != nil {
		return nil, m.err
	}
	data, ok := m.objects[key]
	if !ok {
		return nil, objectstorage.ErrObjectNotFound
	}
	return &objectstorage.Object{Body: io.NopCloser(bytes.NewReader(data.body)), Info: data.info}, nil
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
		File:      r,
		Size:      int64(len(pdfContent)),
	}

	storageKey, err := service.UploadPDF(ctx, opts)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if storageKey != "pdf/"+paperUUID+".pdf" {
		t.Errorf("unexpected storage key: %s", storageKey)
	}
	paper, err := paperRepo.GetByUUID(ctx, paperUUID)
	if err != nil {
		t.Fatal(err)
	}
	if paper.PDFStorageKey != storageKey {
		t.Errorf("stored PDFStorageKey = %q, want %q", paper.PDFStorageKey, storageKey)
	}

	object, err := service.OpenPDF(ctx, storageKey)
	if err != nil {
		t.Fatal(err)
	}
	defer object.Body.Close()
	if got := object.FileName; got != opts.FileName {
		t.Fatalf("stored filename = %q, want %q", got, opts.FileName)
	}
	if object.ContentType != domain.PDFContentType {
		t.Fatalf("content type = %q, want %q", object.ContentType, domain.PDFContentType)
	}
	if object.Size != int64(len(pdfContent)) {
		t.Fatalf("size = %d, want %d", object.Size, len(pdfContent))
	}
	content, err := io.ReadAll(object.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(content, pdfContent) {
		t.Fatalf("stored content = %q, want %q", content, pdfContent)
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
	for _, paperUUID := range paperUUIDs {
		if _, err := repo.Save(ctx, domain.Paper{UUID: paperUUID, DOI: "10.1000/same-paper", Title: "Test Paper"}); err != nil {
			t.Fatal(err)
		}
		if exists, err := service.PDFExists(ctx, "pdf/"+paperUUID+".pdf"); err != nil || exists {
			t.Fatalf("PDFExists before upload = %v, %v", exists, err)
		}
		content := "%PDF-1.4\n" + paperUUID
		storageKey, err := service.UploadPDF(ctx, domain.UploadPDFOptions{
			PaperUUID: paperUUID,
			FileName:  "main.pdf", File: bytes.NewBufferString(content),
		})
		if err != nil {
			t.Fatal(err)
		}
		if storageKey != "pdf/"+paperUUID+".pdf" {
			t.Fatalf("unexpected storage key: %s", storageKey)
		}
	}
	for _, paperUUID := range paperUUIDs {
		if exists, err := service.PDFExists(ctx, "pdf/"+paperUUID+".pdf"); err != nil || !exists {
			t.Fatalf("PDFExists after upload = %v, %v", exists, err)
		}
		object, err := service.OpenPDF(ctx, "pdf/"+paperUUID+".pdf")
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

func TestOpenPDFFallsBackToStorageKey(t *testing.T) {
	for _, tc := range []struct {
		name     string
		metadata map[string]string
	}{
		{name: "no metadata"},
		{name: "missing filename", metadata: map[string]string{"other": "value"}},
		{name: "empty filename", metadata: map[string]string{"filename": ""}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const key = "archive/custom.pdf"
			service := NewPaperService(nil, mockObjectStorage{objects: map[string]storedPDF{
				key: {info: objectstorage.ObjectInfo{Metadata: tc.metadata}},
			}})
			pdf, err := service.OpenPDF(context.Background(), key)
			if err != nil {
				t.Fatal(err)
			}
			defer pdf.Body.Close()
			if pdf.FileName != key {
				t.Fatalf("filename = %q, want %q", pdf.FileName, key)
			}
		})
	}
}

func TestPDFExistsUsesStorageKey(t *testing.T) {
	service := NewPaperService(nil, mockObjectStorage{objects: map[string]storedPDF{
		"archive/custom.pdf": {},
	}})
	ctx := context.Background()
	if exists, err := service.PDFExists(ctx, "archive/custom.pdf"); err != nil || !exists {
		t.Fatalf("PDFExists for existing key = %v, %v", exists, err)
	}
	if exists, err := service.PDFExists(ctx, "archive/missing.pdf"); err != nil || exists {
		t.Fatalf("PDFExists for missing key = %v, %v", exists, err)
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
	if _, err := service.OpenPDF(ctx, "paper-one"); !errors.Is(err, wantErr) {
		t.Fatalf("OpenPDF error = %v, want %v", err, wantErr)
	}
	if _, err := service.PDFExists(ctx, "pdf/paper-one.pdf"); !errors.Is(err, wantErr) {
		t.Fatalf("PDFExists error = %v, want %v", err, wantErr)
	}
}

func (infiniteZeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}
