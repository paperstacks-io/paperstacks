package domain

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
	Publisher string `json:"publisher"`

	// PublishedIn specifies the container in which the work
	// appears, such as a journal, book, or conference proceedings.
	PublishedIn string `json:"published-in"`

	// Pages indicates the page number or page range of the work
	// within the publication (e.g. "11" or "42-53").
	Pages string `json:"pages"`

	// Volume is the volume number of the journal or proceedings.
	Volume string `json:"volume"`

	// Issue is the issue number within the volume.
	Issue string `json:"issue"`

	// ISBN lists International Standard Book Numbers associated
	// with the publication, typically used for books or proceedings.
	ISBN []string `json:"ISBN"`

	// References contains external references related to the work,
	// typically URLs or URIs pointing to additional resources.
	References []string `json:"references"`

	// License specifies the license under which the work is
	// distributed, expressed as an SPDX license identifier.
	License string `json:"license"`

	// Copyright contains the copyright statement associated
	// with the publication.
	Copyright string `json:"copyright"`

	// Funding describes funding sources or grants that supported
	// the work.
	Funding string `json:"funding"`

	// DataSource indicates the origin of the metadata, such as
	// crossref.org, manual user input, etc.
	DataSource string `json:"data-source"`

	// DataSourceTimestamp records when the metadata was retrieved
	// or generated, expressed as an ISO 8601 timestamp.
	DataSourceTimestamp string `json:"data-source-timestamp"`
}
