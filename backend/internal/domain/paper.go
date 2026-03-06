package domain

// Paper represents a scientific publication with bibliographic metadata,
// authors, and associated documents such as PDFs.
type Paper struct {
	// ID uniquely identifies the paper.
	DOI string `json:"DOI"`

	// Title is the full title of the publication.
	Title string `json:"title"`

	// TitleShort is an abbreviated version of the title used for citations.
	TitleShort string `json:"title-short"`

	// Authors lists all authors who contributed to the publication.
	Authors []Author `json:"authors"`

	// PublicationYear is the year the work was published or made public.
	PublicationYear string `json:"publication-year"`

	// PublicationStatus describes the publication state
	// (e.g., "published", "preprint", or "retracted").
	PublicationStatus string `json:"publication-status"`

	// PublicationStatusTimestamp records when the publication status was changed
	// or generated, expressed as an ISO 8601 timestamp.
	PublicationStatusTimestamp string `json:"publication-status-timestamp"`

	// Abstract contains the summary of the publication.
	Abstract string `json:"abstract"`

	// Keywords contains search keywords associated with the paper.
	Keywords string `json:"keywords"`

	// Type specifies the publication type (e.g. "journal" or "conference").
	Type string `json:"type"`

	// PDFs contains URIs pointing to PDF versions of the paper.
	PDFs []string `json:"pdfs"`

	// Metadata contains detailed bibliographic metadata for the publication.
	Metadata Metadata `json:"metadata"`
}
