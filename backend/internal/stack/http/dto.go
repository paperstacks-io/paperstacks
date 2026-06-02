package http

import (
	"time"

	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

type StackRequest struct {
	UUID      string         `json:"uuid"`
	Name      string         `json:"name"`
	Owner     UserRequest    `json:"owner"`
	Papers    []PaperRequest `json:"papers"`
	IsPublic  bool           `json:"is_public"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type StackResponse struct {
	UUID      string          `json:"uuid"`
	Name      string          `json:"name"`
	Owner     UserResponse    `json:"owner"`
	Papers    []PaperResponse `json:"papers"`
	IsPublic  bool            `json:"is_public"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type UserRequest struct {
	ExternalID string    `json:"external_id"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type UserResponse struct {
	ExternalID string    `json:"external_id"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type PaperRequest struct {
	UUID                       string          `json:"uuid"`
	DOI                        string          `json:"doi"`
	Title                      string          `json:"title"`
	TitleShort                 string          `json:"title_short"`
	Authors                    []AuthorRequest `json:"authors"`
	PublicationYear            string          `json:"publication_year"`
	PublicationStatus          string          `json:"publication_status"`
	PublicationStatusTimestamp string          `json:"publication_status_timestamp"`
	Abstract                   string          `json:"abstract"`
	Keywords                   []string        `json:"keywords"`
	Type                       string          `json:"type"`
	PDFs                       []string        `json:"pdfs"`
	Metadata                   MetadataRequest `json:"metadata"`
}

type PaperResponse struct {
	UUID                       string           `json:"uuid"`
	DOI                        string           `json:"doi"`
	Title                      string           `json:"title"`
	TitleShort                 string           `json:"title_short"`
	Authors                    []AuthorResponse `json:"authors"`
	PublicationYear            string           `json:"publication_year"`
	PublicationStatus          string           `json:"publication_status"`
	PublicationStatusTimestamp string           `json:"publication_status_timestamp"`
	Abstract                   string           `json:"abstract"`
	Keywords                   []string         `json:"keywords"`
	Type                       string           `json:"type"`
	PDFs                       []string         `json:"pdfs"`
	Metadata                   MetadataResponse `json:"metadata"`
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
	PublishedIn         string   `json:"published_in"`
	Pages               string   `json:"pages"`
	Volume              string   `json:"volume"`
	Issue               string   `json:"issue"`
	ISBN                []string `json:"isbn"`
	References          []string `json:"references"`
	License             string   `json:"license"`
	Copyright           string   `json:"copyright"`
	Funding             string   `json:"funding"`
	DataSource          string   `json:"data_source"`
	DataSourceTimestamp string   `json:"data_source_timestamp"`
}

type MetadataResponse struct {
	Publisher           string   `json:"publisher"`
	PublishedIn         string   `json:"published_in"`
	Pages               string   `json:"pages"`
	Volume              string   `json:"volume"`
	Issue               string   `json:"issue"`
	ISBN                []string `json:"isbn"`
	References          []string `json:"references"`
	License             string   `json:"license"`
	Copyright           string   `json:"copyright"`
	Funding             string   `json:"funding"`
	DataSource          string   `json:"data_source"`
	DataSourceTimestamp string   `json:"data_source_timestamp"`
}

func (s StackRequest) toDomain() domain.Stack {
	papers := make([]paperDomain.Paper, len(s.Papers))
	for i, paper := range s.Papers {
		papers[i] = paper.toDomain()
	}

	return domain.Stack{
		UUID:      s.UUID,
		Name:      s.Name,
		Owner:     s.Owner.toDomain(),
		Papers:    papers,
		IsPublic:  s.IsPublic,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func NewStackResponse(s domain.Stack) StackResponse {
	papers := make([]PaperResponse, len(s.Papers))
	for i, paper := range s.Papers {
		papers[i] = NewPaperResponse(paper)
	}

	return StackResponse{
		UUID:      s.UUID,
		Name:      s.Name,
		Owner:     NewUserResponse(s.Owner),
		Papers:    papers,
		IsPublic:  s.IsPublic,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func NewStackResponses(stacks []domain.Stack) []StackResponse {
	resp := make([]StackResponse, len(stacks))

	for i, stack := range stacks {
		resp[i] = NewStackResponse(stack)
	}

	return resp
}

func (u UserRequest) toDomain() userDomain.User {
	return userDomain.User{
		ExternalID: u.ExternalID,
		Email:      u.Email,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}

func NewUserResponse(u userDomain.User) UserResponse {
	return UserResponse{
		ExternalID: u.ExternalID,
		Email:      u.Email,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}

func (p PaperRequest) toDomain() paperDomain.Paper {
	authors := make([]paperDomain.Author, len(p.Authors))

	for i, author := range p.Authors {
		authors[i] = author.toDomain()
	}

	return paperDomain.Paper{
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
		Metadata:                   p.Metadata.toDomain(),
	}
}

func NewPaperResponse(p paperDomain.Paper) PaperResponse {
	return PaperResponse{
		UUID:                       p.UUID,
		DOI:                        p.DOI,
		Title:                      p.Title,
		TitleShort:                 p.TitleShort,
		Authors:                    NewAuthorResponses(p.Authors),
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

func NewPaperResponses(papers []paperDomain.Paper) []PaperResponse {
	resp := make([]PaperResponse, len(papers))

	for i, paper := range papers {
		resp[i] = NewPaperResponse(paper)
	}

	return resp
}

func (a AuthorRequest) toDomain() paperDomain.Author {
	return paperDomain.Author{
		NameFirst:   a.NameFirst,
		NameMiddle:  a.NameMiddle,
		NameLast:    a.NameLast,
		Affiliation: a.Affiliation,
		ORCID:       a.ORCID,
	}
}

func NewAuthorResponses(authors []paperDomain.Author) []AuthorResponse {
	resp := make([]AuthorResponse, len(authors))

	for i, author := range authors {
		resp[i] = AuthorResponse{
			NameFirst:   author.NameFirst,
			NameMiddle:  author.NameMiddle,
			NameLast:    author.NameLast,
			Affiliation: author.Affiliation,
			ORCID:       author.ORCID,
		}
	}

	return resp
}

func (m MetadataRequest) toDomain() paperDomain.Metadata {
	return paperDomain.Metadata{
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

func NewMetadataResponse(metadata paperDomain.Metadata) MetadataResponse {
	return MetadataResponse{
		Publisher:           metadata.Publisher,
		PublishedIn:         metadata.PublishedIn,
		Pages:               metadata.Pages,
		Volume:              metadata.Volume,
		Issue:               metadata.Issue,
		ISBN:                metadata.ISBN,
		References:          metadata.References,
		License:             metadata.License,
		Copyright:           metadata.Copyright,
		Funding:             metadata.Funding,
		DataSource:          metadata.DataSource,
		DataSourceTimestamp: metadata.DataSourceTimestamp,
	}
}
