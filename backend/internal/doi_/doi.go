package doi_

import "regexp"

// Metadata holds the most commonly used values returned by the
// CrossRef/Datacite APIs and other DOI lookup services.
//
// Fields are intentionally exported since the data is usually marshaled
// directly into JSON for HTTP responses.
type Metadata struct {
	DOI       string   `json:"doi_"`
	Title     string   `json:"title"`
	Publisher string   `json:"publisher"`
	Type      string   `json:"type"`
	Authors   []string `json:"authors"`
	Published string   `json:"published"`
	URL       string   `json:"url"`
}

// isValidDOIPattern is a precompiled regular expression that matches a
// syntactically valid DOI. The pattern is derived from the DOI Handbook
// and commonly used validation regexes published by CrossRef/Datacite.
//
// RegEx is taken from: https://www.crossref.org/blog/dois-and-matching-regular-expressions/
//
// When porting regexes from other languages note that Go's regexp package
// does not recognize leading/trailing `/` delimiters or the `i` flag. We
// express the case‑insensitive modifier with a `(?i)` prefix and omit the
// slashes.
var isValidDOIPattern = regexp.MustCompile(`(?i)^10\.\d{4,9}/[-._;()/:A-Z0-9]+$`)

// IsValid reports whether the provided string appears to be a valid DOI.
//
// The function performs only a syntactic check; it does **not** query any
// external service. If you need to ensure the DOI actually exists, you
// should perform a lookup against an appropriate resolver.
func IsValid(doi string) bool {
	return isValidDOIPattern.MatchString(doi)
}
