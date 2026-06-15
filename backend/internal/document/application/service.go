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
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

type DocumentService struct {
	repo    domain.Repository
	storage domain.Storage
}

func NewDocumentService(repo domain.Repository, storage domain.Storage) *DocumentService {
	return &DocumentService{
		repo:    repo,
		storage: storage,
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
	document domain.Document,
	user userDomain.User,
	r io.Reader,
) (domain.Document, error) {
	const maxFileSize = 10 * 1024 * 1024
	if document.Size > maxFileSize {
		return domain.Document{}, domain.ErrFileSizeExceeded
	}

	buf := make([]byte, 512)
	n, err := io.ReadFull(io.LimitReader(r, 512), buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return domain.Document{}, fmt.Errorf("failed to read file header: %w", err)
	}

	if n < 5 {
		return domain.Document{}, domain.ErrInvalidFileType
	}

	if string(buf[:5]) != "%PDF-" {
		return domain.Document{}, domain.ErrInvalidFileType
	}

	detectedType := http.DetectContentType(buf[:n])
	if detectedType != "application/pdf" {
		return domain.Document{}, domain.ErrInvalidFileType
	}

	fullReader := io.MultiReader(bytes.NewReader(buf[:n]), r)
	limitReader := &limitCountingReader{
		r:     fullReader,
		limit: maxFileSize,
	}

	docUUID := uuid.NewString()
	storageKey := docUUID + ".pdf"

	storageURI, err := s.storage.Put(ctx, storageKey, limitReader)
	if err != nil {
		return domain.Document{}, fmt.Errorf("failed to store physical file: %w", err)
	}

	doc := domain.Document{
		UUID:         docUUID,
		UploaderUUID: user.ExternalID,
		PaperUUID:    document.PaperUUID,
		FileName:     strings.TrimSpace(document.FileName),
		ContentType:  "application/pdf",
		Size:         limitReader.read,
		StorageURI:   storageURI,
	}

	savedDoc, err := s.repo.Save(ctx, doc)
	if err != nil {
		return domain.Document{}, fmt.Errorf("failed to save document metadata: %w", err)
	}

	return savedDoc, nil
}
