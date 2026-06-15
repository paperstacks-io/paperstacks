package http

import (
	"errors"

	"github.com/paperstacks.io/paperstacks/internal/document/domain"
)

type DocumentRequest struct {
	UUID        string `json:"uuid"`
	PaperUUID   string `json:"paper_uuid"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	StorageURI  string `json:"storage_uri"`
}

type DocumentResponse struct {
	UUID        string `json:"uuid"`
	PaperUUID   string `json:"paper_uuid"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	StorageURI  string `json:"storage_uri"`
}

func (r DocumentRequest) toDomain() domain.Document {
	return domain.Document{
		UUID:        r.UUID,
		PaperUUID:   r.PaperUUID,
		FileName:    r.FileName,
		ContentType: r.ContentType,
		Size:        r.Size,
		StorageURI:  r.StorageURI,
	}
}

func (r DocumentRequest) ValidateUploadRequest() error {
	if r.PaperUUID == "" {
		return errors.New("paper_uuid is required")
	}
	if r.FileName == "" {
		return errors.New("file_name is required")
	}
	return nil
}

func NewDocumentResponse(d domain.Document) DocumentResponse {
	return DocumentResponse{
		UUID:        d.UUID,
		PaperUUID:   d.PaperUUID,
		FileName:    d.FileName,
		ContentType: d.ContentType,
		Size:        d.Size,
		StorageURI:  d.StorageURI,
	}
}
