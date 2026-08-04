package application

import (
	"errors"
	"strings"
	"unicode"

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

func formatInitials(names ...string) string {
	var initials []string

	for _, name := range names {
		for part := range strings.FieldsSeq(name) {
			if initial := formatNameInitial(part); initial != "" {
				initials = append(initials, initial)
			}
		}
	}

	return strings.Join(initials, " ")
}

func formatNameInitial(name string) string {
	parts := strings.Split(name, "-")
	initials := make([]string, 0, len(parts))

	for _, part := range parts {
		if initial := firstLetterInitial(part); initial != "" {
			initials = append(initials, initial)
		}
	}

	return strings.Join(initials, "-")
}

func firstLetterInitial(name string) string {
	for _, letter := range name {
		if unicode.IsLetter(letter) {
			return strings.ToUpper(string(letter)) + "."
		}
	}

	return ""
}

func joinCitationParts(parts ...string) string {
	return joinNonEmpty(" ", parts...)
}

func joinNonEmpty(separator string, parts ...string) string {
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, part)
		}
	}

	return strings.Join(result, separator)
}
