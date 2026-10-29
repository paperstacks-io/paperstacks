package http

import (
	"time"

	"github.com/paperstacks.io/paperstacks/internal/pdfextractor/domain"
)

const (
	FormFieldFile      = "file"
	FormFieldHeadPages = "head_pages"
)

type MetadataResponse struct {
	Title           string     `json:"title"`
	Authors         []string   `json:"authors"`
	Abstract        string     `json:"abstract"`
	DOI             string     `json:"doi"`
	PublicationDate *time.Time `json:"publication_date,omitempty"`
	Keywords        []string   `json:"keywords"`
	IsScanned       bool       `json:"is_scanned"`
	MissingFields   []string   `json:"missing_fields"`
}

func NewMetadataResponse(m domain.ExtractedMetadata) MetadataResponse {
	return MetadataResponse{
		Title:           m.Title,
		Authors:         m.Authors,
		Abstract:        m.Abstract,
		DOI:             m.DOI,
		PublicationDate: m.PublicationDate,
		Keywords:        m.Keywords,
		IsScanned:       m.IsScanned,
		MissingFields:   m.MissingFields,
	}
}

type TextResponse struct {
	FullText   string `json:"full_text"`
	TotalPages int    `json:"total_pages"`
	IsScanned  bool   `json:"is_scanned"`
}

func NewTextResponse(t domain.ExtractedText) TextResponse {
	return TextResponse{
		FullText:   t.FullText,
		TotalPages: t.TotalPages,
		IsScanned:  t.IsScanned,
	}
}
