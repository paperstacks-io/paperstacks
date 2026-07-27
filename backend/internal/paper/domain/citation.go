package domain

import (
	"strings"
)

type CitationStyle string

const (
	CitationStyleAPA CitationStyle = "apa"
)

func (p Paper) APACitation() string {
	paper := p.Normalize()

	authors := formatAPAAuthors(paper.Authors)

	year := paper.PublicationYear
	if year == "" {
		year = "n.d."
	}

	var citation strings.Builder

	if authors != "" {
		citation.WriteString(authors)
		citation.WriteString(" ")
	}

	citation.WriteString("(")
	citation.WriteString(year)
	citation.WriteString("). ")

	citation.WriteString(paper.Title)
	citation.WriteString(".")

	if paper.DOI != "" {
		citation.WriteString(" https://doi.org/" + paper.DOI)
	}

	return citation.String()
}

func formatAPAAuthors(authors []Author) string {
	formattedAuthors := make([]string, 0, len(authors))

	for _, author := range authors {
		author = author.Normalize()

		if author.NameLast == "" {
			continue
		}

		initials := formatAPAInitials(
			author.NameFirst,
			author.NameMiddle,
		)

		formattedAuthor := author.NameLast

		if initials != "" {
			formattedAuthor += ", " + initials
		}

		formattedAuthors = append(
			formattedAuthors,
			formattedAuthor,
		)
	}

	switch len(formattedAuthors) {
	case 0:
		return ""

	case 1:
		return formattedAuthors[0]

	case 2:
		return formattedAuthors[0] +
			", & " +
			formattedAuthors[1]

	default:
		return strings.Join(
			formattedAuthors[:len(formattedAuthors)-1],
			", ",
		) + ", & " +
			formattedAuthors[len(formattedAuthors)-1]
	}
}

func formatAPAInitials(
	firstName string,
	middleName string,
) string {
	initials := make([]string, 0, 2)

	if initial := firstInitial(firstName); initial != "" {
		initials = append(initials, initial+".")
	}

	if initial := firstInitial(middleName); initial != "" {
		initials = append(initials, initial+".")
	}

	return strings.Join(initials, " ")
}

func firstInitial(name string) string {
	name = strings.TrimSpace(name)

	if name == "" {
		return ""
	}

	return strings.ToUpper(
		string([]rune(name)[0]),
	)
}
