package http

import "github.com/paperstacks.io/paperstacks/internal/doi/domain"

type MetadataResponse struct {
	DOI       string   `json:"doi"`
	Title     string   `json:"title"`
	Publisher string   `json:"publisher"`
	Type      string   `json:"type"`
	Authors   []string `json:"authors"`
	Published string   `json:"published"`
	URL       string   `json:"url"`
}

func NewMetadataResponse(metadata *domain.Metadata) MetadataResponse {
	if metadata == nil {
		return MetadataResponse{}
	}

	return MetadataResponse{
		DOI:       metadata.DOI,
		Title:     metadata.Title,
		Publisher: metadata.Publisher,
		Type:      metadata.Type,
		Authors:   metadata.Authors,
		Published: metadata.Published,
		URL:       metadata.URL,
	}
}
