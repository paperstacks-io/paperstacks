package citation

import "errors"

var ErrUnsupportedCitationStyle = errors.New("unsupported citation style")

const (
	CitationStyleAPA  CitationStyle = "APA"
	CitationStyleIEEE CitationStyle = "IEEE"
	CitationStyleACM  CitationStyle = "ACM"
)

type CitationStyle string
