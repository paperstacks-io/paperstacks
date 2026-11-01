package http

import "github.com/paperstacks.io/paperstacks/internal/paper/domain"

type PaperRequest struct {
	UUID                       string                 `json:"uuid"`
	DOI                        string                 `json:"doi"`
	Title                      string                 `json:"title"`
	TitleShort                 string                 `json:"title_short"`
	Authors                    []AuthorRequest        `json:"authors"`
	PublicationDate            PublicationDate        `json:"publication_date"`
	PublicationStatus          string                 `json:"publication_status"`
	PublicationStatusTimestamp string                 `json:"publication_status_timestamp"`
	Abstract                   string                 `json:"abstract"`
	Keywords                   []string               `json:"keywords"`
	Type                       domain.PublicationType `json:"type"`
	PDFStorageKey              string                 `json:"pdf_storage_key"`
	Metadata                   MetadataRequest        `json:"metadata"`
}

type PaperResponse struct {
	UUID                       string                 `json:"uuid"`
	DOI                        string                 `json:"doi"`
	Title                      string                 `json:"title"`
	TitleShort                 string                 `json:"title_short"`
	Authors                    []AuthorResponse       `json:"authors"`
	PublicationDate            PublicationDate        `json:"publication_date"`
	PublicationStatus          string                 `json:"publication_status"`
	PublicationStatusTimestamp string                 `json:"publication_status_timestamp"`
	Abstract                   string                 `json:"abstract"`
	Keywords                   []string               `json:"keywords"`
	Type                       domain.PublicationType `json:"type"`
	PDFStorageKey              string                 `json:"pdf_storage_key"`
	Metadata                   MetadataResponse       `json:"metadata"`
}

type PublicationDate struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

type AuthorRequest struct {
	NameFirst   string `json:"name_first"`
	NameMiddle  string `json:"name_middle"`
	NameLast    string `json:"name_last"`
	Affiliation string `json:"affiliation"`
	ORCID       string `json:"orcid"`
}

type AuthorResponse struct {
	NameFirst   string `json:"name_first"`
	NameMiddle  string `json:"name_middle"`
	NameLast    string `json:"name_last"`
	Affiliation string `json:"affiliation"`
	ORCID       string `json:"orcid"`
}

type MetadataRequest struct {
	Publisher           string   `json:"publisher"`
	JournalTitle        string   `json:"journal_title"`
	JournalAbbrev       string   `json:"journal_abbrev"`
	BookTitle           string   `json:"book_title"`
	SeriesTitle         string   `json:"series_title"`
	EventTitle          string   `json:"event_title"`
	EventPlace          string   `json:"event_place"`
	Institution         string   `json:"institution"`
	Pages               string   `json:"pages"`
	Volume              string   `json:"volume"`
	Issue               string   `json:"issue"`
	ISBN                []string `json:"isbn"`
	ISSN                []string `json:"issn"`
	References          []string `json:"references"`
	License             string   `json:"license"`
	Copyright           string   `json:"copyright"`
	Funding             string   `json:"funding"`
	DataSource          string   `json:"data_source"`
	DataSourceTimestamp string   `json:"data_source_timestamp"`
}

type MetadataResponse struct {
	Publisher           string   `json:"publisher"`
	JournalTitle        string   `json:"journal_title"`
	JournalAbbrev       string   `json:"journal_abbrev"`
	BookTitle           string   `json:"book_title"`
	SeriesTitle         string   `json:"series_title"`
	EventTitle          string   `json:"event_title"`
	EventPlace          string   `json:"event_place"`
	Institution         string   `json:"institution"`
	Pages               string   `json:"pages"`
	Volume              string   `json:"volume"`
	Issue               string   `json:"issue"`
	ISBN                []string `json:"isbn"`
	ISSN                []string `json:"issn"`
	References          []string `json:"references"`
	License             string   `json:"license"`
	Copyright           string   `json:"copyright"`
	Funding             string   `json:"funding"`
	DataSource          string   `json:"data_source"`
	DataSourceTimestamp string   `json:"data_source_timestamp"`
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
		PublicationDate:            r.PublicationDate.toDomain(),
		PublicationStatus:          r.PublicationStatus,
		PublicationStatusTimestamp: r.PublicationStatusTimestamp,
		Abstract:                   r.Abstract,
		Keywords:                   r.Keywords,
		Type:                       r.Type,
		PDFStorageKey:              r.PDFStorageKey,
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
		PublicationDate:            NewPublicationDate(p.PublicationDate),
		PublicationStatus:          p.PublicationStatus,
		PublicationStatusTimestamp: p.PublicationStatusTimestamp,
		Abstract:                   p.Abstract,
		Keywords:                   p.Keywords,
		Type:                       p.Type,
		PDFStorageKey:              p.PDFStorageKey,
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

func (d PublicationDate) toDomain() domain.Date {
	return domain.Date{
		Year:  d.Year,
		Month: d.Month,
		Day:   d.Day,
	}
}

func NewPublicationDate(d domain.Date) PublicationDate {
	return PublicationDate{
		Year:  d.Year,
		Month: d.Month,
		Day:   d.Day,
	}
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
		JournalTitle:        r.JournalTitle,
		JournalAbbrev:       r.JournalAbbrev,
		BookTitle:           r.BookTitle,
		SeriesTitle:         r.SeriesTitle,
		EventTitle:          r.EventTitle,
		EventLocation:       r.EventPlace,
		Institution:         r.Institution,
		Pages:               r.Pages,
		Volume:              r.Volume,
		Issue:               r.Issue,
		ISBN:                r.ISBN,
		ISSN:                r.ISSN,
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
		JournalTitle:        m.JournalTitle,
		JournalAbbrev:       m.JournalAbbrev,
		BookTitle:           m.BookTitle,
		SeriesTitle:         m.SeriesTitle,
		EventTitle:          m.EventTitle,
		EventPlace:          m.EventLocation,
		Institution:         m.Institution,
		Pages:               m.Pages,
		Volume:              m.Volume,
		Issue:               m.Issue,
		ISBN:                m.ISBN,
		ISSN:                m.ISSN,
		References:          m.References,
		License:             m.License,
		Copyright:           m.Copyright,
		Funding:             m.Funding,
		DataSource:          m.DataSource,
		DataSourceTimestamp: m.DataSourceTimestamp,
	}
}
