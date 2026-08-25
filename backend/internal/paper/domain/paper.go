package domain

import (
	"strings"
	"uuid"
)

// Paper represents a scientific publication with bibliographic metadata,
// authors, and associated documents such as PDFs.
type Paper struct {
	// UUID (Version 4) is uniquely identifies a paper across the whole application
	UUID string

	// DOI is the Digital Object Identifier that uniquely identifies
	// the publication (e.g. "10.1145/1234567.1234568").
	DOI string

	// Title is the full title of the publication.
	Title string

	// TitleShort is an abbreviated version of the title used for citations.
	TitleShort string

	// Authors lists all authors who contributed to the publication.
	Authors []Author

	// PublicationDate is the date the work was published or made public.
	// It supports year, year-month, and full-date precision.
	PublicationDate Date

	// PublicationStatus describes the publication state
	// (e.g., "published", "preprint", or "retracted").
	PublicationStatus string

	// PublicationStatusTimestamp records when the publication status was changed
	// or generated, expressed as an ISO 8601 timestamp.
	PublicationStatusTimestamp string

	// Abstract contains the summary of the publication.
	Abstract string

	// Keywords contains search keywords associated with the paper.
	Keywords []string

	// Type classifies the bibliographic kind of the Paper.
	Type PublicationType

	// PDFs contains URIs pointing to PDF versions of the paper.
	PDFs []string

	// Metadata contains detailed bibliographic metadata for the publication.
	Metadata Metadata
}

func (p Paper) Normalize() Paper {
	p.DOI = strings.TrimSpace(p.DOI)
	p.Title = strings.TrimSpace(p.Title)
	p.TitleShort = strings.TrimSpace(p.TitleShort)
	p.PublicationStatus = strings.TrimSpace(p.PublicationStatus)
	p.PublicationStatusTimestamp = strings.TrimSpace(p.PublicationStatusTimestamp)
	p.Abstract = strings.TrimSpace(p.Abstract)
	p.Type = PublicationType(strings.TrimSpace(string(p.Type)))
	p.Metadata = p.Metadata.Normalize()

	for i := range p.Authors {
		p.Authors[i] = p.Authors[i].Normalize()
	}

	for i := range p.PDFs {
		p.PDFs[i] = strings.TrimSpace(p.PDFs[i])
	}

	for i := range p.Keywords {
		p.Keywords[i] = strings.TrimSpace(p.Keywords[i])
	}

	return p
}

func (p Paper) Validate() error {
	_, err := uuid.Parse(p.UUID)
	if err != nil {
		return ErrInvalidPaper
	}

	if strings.TrimSpace(p.DOI) == "" {
		return ErrInvalidPaper
	}

	if strings.TrimSpace(p.Title) == "" {
		return ErrInvalidPaper
	}

	if !p.PublicationDate.IsValid() {
		return ErrInvalidPaper
	}

	if !p.Type.IsValid() {
		return ErrInvalidPaper
	}

	return nil
}
