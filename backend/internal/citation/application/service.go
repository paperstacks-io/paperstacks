package application

import (
	"errors"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

var ErrUnsupportedCitationStyle = errors.New("unsupported citation style")

type CitationStyle string

const (
	CitationStyleAPA  CitationStyle = "apa"
	CitationStyleIEEE CitationStyle = "ieee"
	CitationStyleACM  CitationStyle = "acm"
)

type citationFormatter func(domain.Paper) string

type CitationService map[CitationStyle]citationFormatter

func (s CitationService) Format(paper domain.Paper, style CitationStyle) (string, error) {
	formatter, ok := s[style]
	if !ok {
		return "", ErrUnsupportedCitationStyle
	}

	return formatter(paper.Normalize()), nil
}
