package domain

import "strings"

// Author represents a person who contributed to a publication.
//
// It contains the author's personal name components, institutional
// affiliation, and an optional ORCID identifier used to uniquely
// identify researchers.
type Author struct {
	// NameFirst is the author's given (first) name.
	NameFirst string

	// NameMiddle is the author's middle name or middle initials.
	// It may be empty if the author does not use a middle name.
	NameMiddle string

	// NameLast is the author's family or last name.
	NameLast string

	// Affiliation specifies the institution or organization
	// the author was associated with at the time of publication.
	Affiliation string

	// ORCID is the author's ORCID identifier, a persistent
	// digital identifier for researchers (e.g. "0000-0002-1825-0097").
	// It may be empty if the author has no ORCID.
	ORCID string
}

// FullName returns the author's name assembled from the available name
// components. Middle names are reduced to their first initial.
func (a Author) FullName() string {
	if a.NameMiddle != "" {
		return a.NameFirst + " " + a.NameMiddle[:1] + ". " + a.NameLast
	}

	return a.NameFirst + " " + a.NameLast
}

func (a Author) Normalize() Author {
	a.NameFirst = strings.TrimSpace(a.NameFirst)
	a.NameMiddle = strings.TrimSpace(a.NameMiddle)
	a.NameLast = strings.TrimSpace(a.NameLast)
	a.Affiliation = strings.TrimSpace(a.Affiliation)
	a.ORCID = strings.TrimSpace(a.ORCID)

	return a
}
