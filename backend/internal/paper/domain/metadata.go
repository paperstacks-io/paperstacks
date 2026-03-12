package domain

import "strings"

// Metadata contains bibliographic and provenance information
// describing where and how a publication was published.
//
// It includes identifiers such as DOI and ISBN, publication
// details like volume and issue, licensing information, and
// metadata about the source from which this information was
// retrieved (e.g., CrossRef or IEEE).
type Metadata struct {
	// Publisher is the organization responsible for publishing
	// the work (e.g., "IEEE", "ACM", or "Springer").
	Publisher string

	// PublishedIn specifies the container in which the work
	// appears, such as a journal, book, or conference proceedings.
	PublishedIn string

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
	m.Publisher = strings.TrimSpace(m.Publisher)
	m.PublishedIn = strings.TrimSpace(m.PublishedIn)
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

	for i := range m.References {
		m.References[i] = strings.TrimSpace(m.References[i])
	}

	return m
}
