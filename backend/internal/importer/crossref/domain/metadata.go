// Package domain defines the domain model the crawler imports Crossref
// works into. It mirrors the fields present in the Crossref public data
// export (https://www.crossref.org/documentation/metadata-plus/) closely
// enough to hold every record losslessly, while staying independent of the
// raw JSON shape so parsing concerns stay out of it.
package domain

import "time"

// Paper is a single Crossref "work" record — a journal article, book,
// book chapter, or similar publication.
type Paper struct {
	// DOI is the Digital Object Identifier that uniquely identifies the
	// work (e.g. "10.1006/jmcc.2000.1342"). Every record has one; it is
	// the natural primary key for the import.
	DOI string

	// Type is the Crossref work type (e.g. "journal-article",
	// "book-chapter", "book").
	Type string

	// Source identifies the metadata source as reported by Crossref
	// (typically "Crossref").
	Source string

	Title               string
	Subtitle            string
	ContainerTitle      string // Journal, book, or proceedings title.
	ShortContainerTitle string
	Abstract            string // May contain embedded JATS XML markup.

	Publisher         string
	PublisherLocation string
	Member            string // Crossref member ID of the depositing publisher.
	Prefix            string // DOI prefix (e.g. "10.1006").

	Language string
	URL      string // Canonical resolver URL, usually "https://doi.org/<DOI>".

	// ResourceURL is the primary full-text landing page URL, if reported.
	ResourceURL string

	Volume           string
	Issue            string
	Page             string
	SpecialNumbering string

	ISSNs []ISSN
	ISBNs []ISBN

	Authors    []Contributor
	Editors    []Contributor
	Funders    []Funder
	References []Reference
	Licenses   []License

	// AlternativeIDs lists other identifiers publishers use for the same
	// work (e.g. a publisher's internal article ID).
	AlternativeIDs []string

	Issued          PartialDate // Earliest known publication date.
	Published       PartialDate
	PublishedPrint  PartialDate
	PublishedOnline PartialDate

	Created   time.Time // When the DOI was first registered with Crossref.
	Deposited time.Time // When this metadata version was deposited.
	Indexed   time.Time // When Crossref last indexed this record.

	// ReferenceCount and ReferencesCount both report the number of
	// entries in References; Crossref exposes both names in its schema,
	// so both are kept for fidelity with the source data.
	ReferenceCount      int
	ReferencesCount     int
	IsReferencedByCount int // Number of works citing this one, per Crossref.
}

// ISSN is an International Standard Serial Number bound to a specific
// medium (e.g. "print" or "electronic").
type ISSN struct {
	Value string
	Type  string
}

// ISBN is an International Standard Book Number bound to a specific
// medium (e.g. "print" or "electronic").
type ISBN struct {
	Value string
	Type  string
}

// Contributor represents a person credited on a work, either as an
// author or an editor.
type Contributor struct {
	GivenName    string
	FamilyName   string
	Sequence     string // "first" or "additional".
	ORCID        string
	Affiliations []string
}

// FullName returns the contributor's name assembled from its parts.
func (c Contributor) FullName() string {
	if c.GivenName == "" {
		return c.FamilyName
	}
	if c.FamilyName == "" {
		return c.GivenName
	}
	return c.GivenName + " " + c.FamilyName
}

// Funder represents a funding body credited for supporting the work.
type Funder struct {
	Name   string
	DOI    string
	Awards []string // Grant or award numbers attributed to this funder.
}

// Reference is a bibliographic reference cited by the work. DOI, and
// often several other fields, may be empty for unstructured or
// pre-DOI-era citations.
type Reference struct {
	Key          string
	DOI          string
	ArticleTitle string
	Author       string
	JournalTitle string
	Volume       string
	FirstPage    string
	Year         string
	Unstructured string // Free-text citation when structured fields are absent.
}

// License describes a usage license applicable to the work.
type License struct {
	URL            string
	ContentVersion string
	DelayInDays    int
	Start          PartialDate
}

// PartialDate represents a Crossref "date-parts" style date, which may be
// known only to year, year+month, or year+month+day precision. A
// PartialDate with Year == 0 carries no date information.
type PartialDate struct {
	Year  int
	Month int // 0 if unknown.
	Day   int // 0 if unknown.
}

// IsZero reports whether the date carries no information.
func (d PartialDate) IsZero() bool {
	return d.Year == 0
}

// Time converts the date to a UTC time.Time, defaulting unknown month or
// day components to 1. It returns the zero time.Time if IsZero is true.
func (d PartialDate) Time() time.Time {
	if d.IsZero() {
		return time.Time{}
	}
	month := d.Month
	if month == 0 {
		month = 1
	}
	day := d.Day
	if day == 0 {
		day = 1
	}
	return time.Date(d.Year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
