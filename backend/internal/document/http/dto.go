package http

import (
	"github.com/paperstacks.io/paperstacks/internal/document/domain"
)

type DocumentResponse struct {
	Key         string `json:"key"`
	PaperUUID   string `json:"paper_uuid"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

func NewDocumentResponse(d domain.Document) DocumentResponse {
	return DocumentResponse{
		Key:         d.Key,
		PaperUUID:   d.PaperUUID,
		FileName:    d.FileName,
		ContentType: d.ContentType,
		Size:        d.Size,
	}
}
