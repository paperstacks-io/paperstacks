package server

import "github.com/paperstacks.io/paperstacks/internal/domain"

type createPaperRequest struct {
	ID                         string          `json:"id"`
	Title                      string          `json:"title"`
	TitleShort                 string          `json:"title-short"`
	Authors                    []domain.Author `json:"authors"`
	PublicationYear            string          `json:"publication-year"`
	PublicationStatus          string          `json:"publication-status"`
	PublicationStatusTimestamp string          `json:"publication-status-timestamp"`
	Abstract                   string          `json:"abstract"`
	Keywords                   string          `json:"keywords"`
	Type                       string          `json:"type"`
	PDFs                       []string        `json:"pdfs"`
	Metadata                   domain.Metadata `json:"metadata"`
}

type updatePaperRequest struct {
	Title                      string          `json:"title"`
	TitleShort                 string          `json:"title-short"`
	Authors                    []domain.Author `json:"authors"`
	PublicationYear            string          `json:"publication-year"`
	PublicationStatus          string          `json:"publication-status"`
	PublicationStatusTimestamp string          `json:"publication-status-timestamp"`
	Abstract                   string          `json:"abstract"`
	Keywords                   string          `json:"keywords"`
	Type                       string          `json:"type"`
	PDFs                       []string        `json:"pdfs"`
	Metadata                   domain.Metadata `json:"metadata"`
}

type paperResponse struct {
	ID                         string          `json:"id"`
	Title                      string          `json:"title"`
	TitleShort                 string          `json:"title-short"`
	Authors                    []domain.Author `json:"authors"`
	PublicationYear            string          `json:"publication-year"`
	PublicationStatus          string          `json:"publication-status"`
	PublicationStatusTimestamp string          `json:"publication-status-timestamp"`
	Abstract                   string          `json:"abstract"`
	Keywords                   string          `json:"keywords"`
	Type                       string          `json:"type"`
	PDFs                       []string        `json:"pdfs"`
	Metadata                   domain.Metadata `json:"metadata"`
}

func paperToResponse(p domain.Paper) paperResponse {
	return paperResponse{
		ID:                         p.ID,
		Title:                      p.Title,
		TitleShort:                 p.TitleShort,
		Authors:                    p.Authors,
		PublicationYear:            p.PublicationYear,
		PublicationStatus:          p.PublicationStatus,
		PublicationStatusTimestamp: p.PublicationStatusTimestamp,
		Abstract:                   p.Abstract,
		Keywords:                   p.Keywords,
		Type:                       p.Type,
		PDFs:                       p.PDFs,
		Metadata:                   p.Metadata,
	}
}

func papersMapToResponse(m map[string]domain.Paper) map[string]paperResponse {
	out := make(map[string]paperResponse, len(m))
	for id, p := range m {
		out[id] = paperToResponse(p)
	}
	return out
}
