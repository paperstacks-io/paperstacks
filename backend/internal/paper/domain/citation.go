package domain

import (
	"fmt"
)

type CitationStyle string

const (
	CitationStyleAPA  CitationStyle = "apa"
	CitationStyleIEEE CitationStyle = "ieee"
	CitationStyleACM  CitationStyle = "acm"
)

func (p Paper) Citation(style CitationStyle) (string, error) {
	switch style {
	case CitationStyleAPA:
		return p.APACitation(), nil

	/*
		case CitationStyleIEEE:
			return p.IEEECitation(), nil

		case CitationStyleACM:
			return p.ACMCitation(), nil
	*/

	default:
		return "", fmt.Errorf(
			"unsupported citation style: %q",
			style,
		)
	}
}
