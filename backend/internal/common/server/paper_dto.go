package server

import (
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

type CreatePaperRequest struct {
	DOI                        string          `json:"DOI"`
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

type UpdatePaperRequest struct {
	DOI                        string          `json:"DOI"`
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

type PaperResponse struct {
	DOI                        string          `json:"DOI"`
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

func (r CreatePaperRequest) ToDomain() domain.Paper {
	return domain.Paper{
		DOI:                        r.DOI,
		Title:                      r.Title,
		TitleShort:                 r.TitleShort,
		Authors:                    r.Authors,
		PublicationYear:            r.PublicationYear,
		PublicationStatus:          r.PublicationStatus,
		PublicationStatusTimestamp: r.PublicationStatusTimestamp,
		Abstract:                   r.Abstract,
		Keywords:                   r.Keywords,
		Type:                       r.Type,
		PDFs:                       r.PDFs,
		Metadata:                   r.Metadata,
	}
}

func (r UpdatePaperRequest) ToDomain() domain.Paper {
	return domain.Paper{
		DOI:                        r.DOI,
		Title:                      r.Title,
		TitleShort:                 r.TitleShort,
		Authors:                    r.Authors,
		PublicationYear:            r.PublicationYear,
		PublicationStatus:          r.PublicationStatus,
		PublicationStatusTimestamp: r.PublicationStatusTimestamp,
		Abstract:                   r.Abstract,
		Keywords:                   r.Keywords,
		Type:                       r.Type,
		PDFs:                       r.PDFs,
		Metadata:                   r.Metadata,
	}
}

func paperToResponse(p domain.Paper) PaperResponse {
	return PaperResponse{
		DOI:                        p.DOI,
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

func PapersToResponse(papers []domain.Paper) []PaperResponse {
	out := make([]PaperResponse, 0, len(papers))

	for _, p := range papers {
		out = append(out, paperToResponse(p))
	}

	return out
}
