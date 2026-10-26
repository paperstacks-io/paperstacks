package postgres

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

// paperDAO represents a database row. SQL nullability stays out of the domain.
type paperDAO struct {
	UUID              string
	DOI               pgtype.Text
	Title             pgtype.Text
	TitleShort        pgtype.Text
	Year              pgtype.Int2
	Month             pgtype.Int2
	Day               pgtype.Int2
	Type              pgtype.Text
	PublicationStatus pgtype.Text
	StatusTimestamp   pgtype.Timestamptz
	Abstract          pgtype.Text
	Keywords          []string
	PDFURL            pgtype.Text
}

func (d paperDAO) domain() domain.Paper {
	return domain.Paper{
		UUID:                       d.UUID,
		DOI:                        d.DOI.String,
		Title:                      d.Title.String,
		TitleShort:                 d.TitleShort.String,
		PublicationDate:            domain.Date{Year: int(d.Year.Int16), Month: int(d.Month.Int16), Day: int(d.Day.Int16)},
		Type:                       domain.PublicationType(d.Type.String),
		PublicationStatus:          d.PublicationStatus.String,
		PublicationStatusTimestamp: formatTimestamp(d.StatusTimestamp),
		Abstract:                   d.Abstract.String,
		Keywords:                   d.Keywords,
		PDFURL:                     d.PDFURL.String,
	}
}

// metadataDAO represents bibliographic metadata, including its paper foreign key.
type metadataDAO struct {
	PaperUUID           string
	Publisher           pgtype.Text
	JournalTitle        pgtype.Text
	JournalAbbrev       pgtype.Text
	BookTitle           pgtype.Text
	SeriesTitle         pgtype.Text
	EventTitle          pgtype.Text
	EventLocation       pgtype.Text
	Institution         pgtype.Text
	Pages               pgtype.Text
	Volume              pgtype.Text
	Issue               pgtype.Text
	ISBN                []string
	ISSN                []string
	References          []string
	License             pgtype.Text
	Copyright           pgtype.Text
	Funding             pgtype.Text
	DataSource          pgtype.Text
	DataSourceTimestamp pgtype.Timestamptz
}

func (d metadataDAO) domain() domain.Metadata {
	return domain.Metadata{
		Publisher:           d.Publisher.String,
		JournalTitle:        d.JournalTitle.String,
		JournalAbbrev:       d.JournalAbbrev.String,
		BookTitle:           d.BookTitle.String,
		SeriesTitle:         d.SeriesTitle.String,
		EventTitle:          d.EventTitle.String,
		EventLocation:       d.EventLocation.String,
		Institution:         d.Institution.String,
		Pages:               d.Pages.String,
		Volume:              d.Volume.String,
		Issue:               d.Issue.String,
		ISBN:                d.ISBN,
		ISSN:                d.ISSN,
		References:          d.References,
		License:             d.License.String,
		Copyright:           d.Copyright.String,
		Funding:             d.Funding.String,
		DataSource:          d.DataSource.String,
		DataSourceTimestamp: formatTimestamp(d.DataSourceTimestamp),
	}
}

// authorDAO represents an author joined with its affiliation and paper link.
type authorDAO struct {
	PaperUUID   string
	First       pgtype.Text
	Middle      pgtype.Text
	Last        pgtype.Text
	Affiliation pgtype.Text
	ORCID       pgtype.Text
}

func (d authorDAO) domain() domain.Author {
	return domain.Author{
		NameFirst:   d.First.String,
		NameMiddle:  d.Middle.String,
		NameLast:    d.Last.String,
		Affiliation: d.Affiliation.String,
		ORCID:       d.ORCID.String,
	}
}
