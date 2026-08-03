package application

import (
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

func FormatIEEE(paper domain.Paper) string {
	authors := formatIEEEAuthors(paper.Authors)

	if authors == "" {
		return joinCitationParts("", "", "", "")
	}

	return joinCitationParts(authors, "", "", "", "")
}

func formatIEEEAuthors(authors []domain.Author) string {
	formatted := make([]string, 0, len(authors))

	for _, author := range authors {
		if name := formatIEEEAuthor(author); name != "" {
			formatted = append(formatted, name)
		}
	}

	switch len(formatted) {
	case 0:
		return ""
	case 1:
		return formatted[0]
	case 2:
		return strings.Join(formatted, " and ")
	}

	if len(formatted) > 6 {
		return formatted[0] + " et al."
	}

	return strings.Join(formatted[:len(formatted)-1], ", ") +
		", and " +
		formatted[len(formatted)-1]
}

func formatIEEEAuthor(author domain.Author) string {
	if author.NameLast == "" {
		return ""
	}

	initials := formatInitials(
		author.NameFirst,
		author.NameMiddle,
	)

	if initials == "" {
		return author.NameLast
	}

	return initials + " " + author.NameLast
}
