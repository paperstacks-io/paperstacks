package http

import "github.com/paperstacks.io/paperstacks/internal/paper/domain"

type PaperRequest struct {
	UUID                       string          `json:"uuid"`
	DOI                        string          `json:"DOI"`
	Title                      string          `json:"title"`
	TitleShort                 string          `json:"title-short"`
	Authors                    []AuthorRequest `json:"authors"`
	PublicationYear            string          `json:"publication-year"`
	PublicationStatus          string          `json:"publication-status"`
	PublicationStatusTimestamp string          `json:"publication-status-timestamp"`
	Abstract                   string          `json:"abstract"`
	Keywords                   []string        `json:"keywords"`
	Type                       string          `json:"type"`
	PDFs                       []string        `json:"pdfs"`
	Metadata                   MetadataRequest `json:"metadata"`
}

type PaperResponse struct {
	UUID                       string           `json:"uuid"`
	DOI                        string           `json:"DOI"`
	Title                      string           `json:"title"`
	TitleShort                 string           `json:"title-short"`
	Authors                    []AuthorResponse `json:"authors"`
	PublicationYear            string           `json:"publication-year"`
	PublicationStatus          string           `json:"publication-status"`
	PublicationStatusTimestamp string           `json:"publication-status-timestamp"`
	Abstract                   string           `json:"abstract"`
	Keywords                   []string         `json:"keywords"`
	Type                       string           `json:"type"`
	PDFs                       []string         `json:"pdfs"`
	Metadata                   MetadataResponse `json:"metadata"`
}

type AuthorRequest struct {
	NameFirst   string `json:"name-first"`
	NameMiddle  string `json:"name-middle"`
	NameLast    string `json:"name-last"`
	Affiliation string `json:"affiliation"`
	ORCID       string `json:"orcid"`
}

type AuthorResponse struct {
	NameFirst   string `json:"name-first"`
	NameMiddle  string `json:"name-middle"`
	NameLast    string `json:"name-last"`
	Affiliation string `json:"affiliation"`
	ORCID       string `json:"orcid"`
}

type MetadataRequest struct {
	Publisher           string   `json:"publisher"`
	PublishedIn         string   `json:"published-in"`
	Pages               string   `json:"pages"`
	Volume              string   `json:"volume"`
	Issue               string   `json:"issue"`
	ISBN                []string `json:"ISBN"`
	References          []string `json:"references"`
	License             string   `json:"license"`
	Copyright           string   `json:"copyright"`
	Funding             string   `json:"funding"`
	DataSource          string   `json:"data-source"`
	DataSourceTimestamp string   `json:"data-source-timestamp"`
}

type MetadataResponse struct {
	Publisher           string   `json:"publisher"`
	PublishedIn         string   `json:"published-in"`
	Pages               string   `json:"pages"`
	Volume              string   `json:"volume"`
	Issue               string   `json:"issue"`
	ISBN                []string `json:"ISBN"`
	References          []string `json:"references"`
	License             string   `json:"license"`
	Copyright           string   `json:"copyright"`
	Funding             string   `json:"funding"`
	DataSource          string   `json:"data-source"`
	DataSourceTimestamp string   `json:"data-source-timestamp"`
}

func (r PaperRequest) toDomain() domain.Paper {
	authors := make([]domain.Author, 0, len(r.Authors))
	for _, author := range r.Authors {
		authors = append(authors, author.toDomain())
	}

	return domain.Paper{
		UUID:                       r.UUID,
		DOI:                        r.DOI,
		Title:                      r.Title,
		TitleShort:                 r.TitleShort,
		Authors:                    authors,
		PublicationYear:            r.PublicationYear,
		PublicationStatus:          r.PublicationStatus,
		PublicationStatusTimestamp: r.PublicationStatusTimestamp,
		Abstract:                   r.Abstract,
		Keywords:                   r.Keywords,
		Type:                       r.Type,
		PDFs:                       r.PDFs,
		Metadata:                   r.Metadata.toDomain(),
	}
}

func NewPaperResponse(p domain.Paper) PaperResponse {
	authors := make([]AuthorResponse, 0, len(p.Authors))
	for _, author := range p.Authors {
		authors = append(authors, NewAuthorResponse(author))
	}

	return PaperResponse{
		UUID:                       p.UUID,
		DOI:                        p.DOI,
		Title:                      p.Title,
		TitleShort:                 p.TitleShort,
		Authors:                    authors,
		PublicationYear:            p.PublicationYear,
		PublicationStatus:          p.PublicationStatus,
		PublicationStatusTimestamp: p.PublicationStatusTimestamp,
		Abstract:                   p.Abstract,
		Keywords:                   p.Keywords,
		Type:                       p.Type,
		PDFs:                       p.PDFs,
		Metadata:                   NewMetadataResponse(p.Metadata),
	}
}

func NewPaperResponses(papers []domain.Paper) []PaperResponse {
	out := make([]PaperResponse, 0, len(papers))
	for _, paper := range papers {
		out = append(out, NewPaperResponse(paper))
	}

	return out
}

func (r AuthorRequest) toDomain() domain.Author {
	return domain.Author{
		NameFirst:   r.NameFirst,
		NameMiddle:  r.NameMiddle,
		NameLast:    r.NameLast,
		Affiliation: r.Affiliation,
		ORCID:       r.ORCID,
	}
}

func NewAuthorResponse(a domain.Author) AuthorResponse {
	return AuthorResponse{
		NameFirst:   a.NameFirst,
		NameMiddle:  a.NameMiddle,
		NameLast:    a.NameLast,
		Affiliation: a.Affiliation,
		ORCID:       a.ORCID,
	}
}

func (r MetadataRequest) toDomain() domain.Metadata {
	return domain.Metadata{
		Publisher:           r.Publisher,
		PublishedIn:         r.PublishedIn,
		Pages:               r.Pages,
		Volume:              r.Volume,
		Issue:               r.Issue,
		ISBN:                r.ISBN,
		References:          r.References,
		License:             r.License,
		Copyright:           r.Copyright,
		Funding:             r.Funding,
		DataSource:          r.DataSource,
		DataSourceTimestamp: r.DataSourceTimestamp,
	}
}

func NewMetadataResponse(m domain.Metadata) MetadataResponse {
	return MetadataResponse{
		Publisher:           m.Publisher,
		PublishedIn:         m.PublishedIn,
		Pages:               m.Pages,
		Volume:              m.Volume,
		Issue:               m.Issue,
		ISBN:                m.ISBN,
		References:          m.References,
		License:             m.License,
		Copyright:           m.Copyright,
		Funding:             m.Funding,
		DataSource:          m.DataSource,
		DataSourceTimestamp: m.DataSourceTimestamp,
	}
}
