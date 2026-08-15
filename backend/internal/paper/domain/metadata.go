package domain

import "strings"

// Metadata contains bibliographic and provenance information
// describing where and how a publication was published.
//
// It includes identifiers such as DOI, ISBN, and ISSN, publication
// details like volume and issue, licensing information, and provenance
// metadata about the source from which this information was
// retrieved (e.g., CrossRef or IEEE).
type Metadata struct {
	// Publisher is the organization responsible for publishing
	// the work (e.g., "IEEE", "ACM", or "Springer").
	Publisher string

	// JournalTitle is the title of the journal in which the work appears.
	JournalTitle string

	// JournalAbbrev is the abbreviated title of the journal.
	JournalAbbrev string

	// BookTitle is the title of the book or proceedings containing the work.
	BookTitle string

	// SeriesTitle is the title of the series containing the work.
	SeriesTitle string

	// EventTitle is the title of the conference or other event associated with the work.
	EventTitle string

	// EventLocation is the location of the conference or other event associated with the work.
	EventLocation string

	// Institution is the institution associated with a thesis or report.
	Institution string

	// Pages indicates the page number or page range of the work
	// within the publication (e.g. "11" or "42-53").
	Pages string

	// Volume is the volume number of the journal or proceedings.
	Volume string

	// Issue is the issue number within the volume.
	Issue string

	// ISBN lists International Standard Book Numbers associated
	// with the publication, typically used for books or proceedings.
	ISBN []string

	// ISSN lists International Standard Serial Numbers associated with
	// journal and other serial publications.
	ISSN []string

	// References contains external references related to the work,
	// typically URLs or URIs pointing to additional resources.
	References []string

	// License specifies the license under which the work is
	// distributed, expressed as an SPDX license identifier.
	License string

	// Copyright contains the copyright statement associated
	// with the publication.
	Copyright string

	// Funding describes funding sources or grants that supported
	// the work.
	Funding string

	// DataSource indicates the origin of the metadata, such as
	// crossref.org, manual user input, etc.
	DataSource string

	// DataSourceTimestamp records when the metadata was retrieved
	// or generated, expressed as an ISO 8601 timestamp.
	DataSourceTimestamp string
}

func (m Metadata) Normalize() Metadata {
	m.JournalTitle = strings.TrimSpace(m.JournalTitle)
	m.JournalAbbrev = strings.TrimSpace(m.JournalAbbrev)
	m.BookTitle = strings.TrimSpace(m.BookTitle)
	m.SeriesTitle = strings.TrimSpace(m.SeriesTitle)
	m.EventTitle = strings.TrimSpace(m.EventTitle)
	m.EventLocation = strings.TrimSpace(m.EventLocation)
	m.Institution = strings.TrimSpace(m.Institution)
	m.Pages = strings.TrimSpace(m.Pages)
	m.Volume = strings.TrimSpace(m.Volume)
	m.Issue = strings.TrimSpace(m.Issue)
	m.License = strings.TrimSpace(m.License)
	m.Copyright = strings.TrimSpace(m.Copyright)
	m.Funding = strings.TrimSpace(m.Funding)
	m.DataSource = strings.TrimSpace(m.DataSource)
	m.DataSourceTimestamp = strings.TrimSpace(m.DataSourceTimestamp)

	for i := range m.ISBN {
		m.ISBN[i] = strings.TrimSpace(m.ISBN[i])
	}

	for i := range m.ISSN {
		m.ISSN[i] = strings.TrimSpace(m.ISSN[i])
	}

	for i := range m.References {
		m.References[i] = strings.TrimSpace(m.References[i])
	}

	return m
}
