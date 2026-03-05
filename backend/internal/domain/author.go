package domain

// Author represents a person who contributed to a publication.
//
// It contains the author's personal name components, institutional
// affiliation, and an optional ORCID identifier used to uniquely
// identify researchers.
type Author struct {
	// NameFirst is the author's given (first) name.
	NameFirst string `json:"name-first"`

	// NameMiddle is the author's middle name or middle initials.
	// It may be empty if the author does not use a middle name.
	NameMiddle string `json:"name-middle"`

	// NameLast is the author's family or last name.
	NameLast string `json:"name-last"`

	// Affiliation specifies the institution or organization
	// the author was associated with at the time of publication.
	Affiliation string `json:"affiliation"`

	// ORCID is the author's ORCID identifier, a persistent
	// digital identifier for researchers (e.g. "0000-0002-1825-0097").
	// It may be empty if the author has no ORCID.
	ORCID string `json:"orcid"`
}

// FullName returns the author's full name assembled from
// the available name components.
func (a Author) FullName() string {
	if a.NameMiddle != "" {
		return a.NameFirst + " " + a.NameMiddle + " " + a.NameLast
	}
	return a.NameFirst + " " + a.NameLast
}
