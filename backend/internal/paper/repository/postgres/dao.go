package postgres

import (
	"database/sql"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

// paperDAO represents a database row. SQL nullability stays out of the domain.
type paperDAO struct {
	UUID              string
	DOI               sql.NullString
	Title             sql.NullString
	TitleShort        sql.NullString
	Year              sql.NullInt16
	Month             sql.NullInt16
	Day               sql.NullInt16
	Type              sql.NullString
	PublicationStatus sql.NullString
	StatusTimestamp   sql.NullTime
	Abstract          sql.NullString
	Keywords          []string
	PDFURL            sql.NullString
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
	Publisher           sql.NullString
	JournalTitle        sql.NullString
	JournalAbbrev       sql.NullString
	BookTitle           sql.NullString
	SeriesTitle         sql.NullString
	EventTitle          sql.NullString
	EventLocation       sql.NullString
	Institution         sql.NullString
	Pages               sql.NullString
	Volume              sql.NullString
	Issue               sql.NullString
	ISBN                []string
	ISSN                []string
	References          []string
	License             sql.NullString
	Copyright           sql.NullString
	Funding             sql.NullString
	DataSource          sql.NullString
	DataSourceTimestamp sql.NullTime
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
	First       sql.NullString
	Middle      sql.NullString
	Last        sql.NullString
	Affiliation sql.NullString
	ORCID       sql.NullString
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
