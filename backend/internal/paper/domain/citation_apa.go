package domain

import (
	"strings"
	"unicode"
)

func (p Paper) APACitation() string {
	paper := p.Normalize()

	return joinCitationParts(
		formatAPAAuthors(paper.Authors),
		formatAPAYear(paper.PublicationYear),
		formatAPATitle(paper.Title),
		formatAPASource(paper),
		formatAPADOI(paper.DOI),
	)
}

func formatAPAAuthors(authors []Author) string {
	formatted := make([]string, 0, len(authors))

	for _, author := range authors {
		author = author.Normalize()

		if author.NameLast == "" {
			continue
		}

		name := author.NameLast

		if initials := formatInitials(
			author.NameFirst,
			author.NameMiddle,
		); initials != "" {
			name += ", " + initials
		}

		formatted = append(formatted, name)
	}

	switch len(formatted) {
	case 0:
		return ""
	case 1:
		return formatted[0]
	case 2:
		return strings.Join(formatted, ", & ")
	}

	if len(formatted) > 20 {
		return strings.Join(formatted[:19], ", ") +
			", ... " +
			formatted[len(formatted)-1]
	}

	return strings.Join(formatted[:len(formatted)-1], ", ") +
		", & " +
		formatted[len(formatted)-1]
}

func formatInitials(names ...string) string {
	var initials []string

	for _, name := range names {
		for part := range strings.FieldsSeq(name) {
			nameParts := strings.Split(part, "-")

			for i, namePart := range nameParts {
				for _, letter := range namePart {
					if unicode.IsLetter(letter) {
						nameParts[i] = strings.ToUpper(string(letter)) + "."
						break
					}
				}
			}

			initials = append(
				initials,
				strings.Join(nameParts, "-"),
			)
		}
	}

	return strings.Join(initials, " ")
}

func formatAPAYear(year string) string {
	if year = strings.TrimSpace(year); year == "" {
		year = "n.d."
	}

	return "(" + year + ")."
}

func formatAPATitle(title string) string {
	title = strings.TrimSpace(title)

	if title == "" || strings.ContainsAny(title[len(title)-1:], ".?!") {
		return title
	}

	return title + "."
}

func formatAPASource(paper Paper) string {
	metadata := paper.Metadata

	publishedIn := strings.TrimSpace(metadata.PublishedIn)
	volume := strings.TrimSpace(metadata.Volume)
	issue := strings.TrimSpace(metadata.Issue)
	pages := strings.ReplaceAll(
		strings.TrimSpace(metadata.Pages),
		"-",
		"–",
	)

	if volume != "" && issue != "" {
		volume += "(" + issue + ")"
	}

	source := joinNonEmpty(", ", publishedIn, volume, pages)

	if source == "" {
		return ""
	}

	return source + "."
}

func formatAPADOI(doi string) string {
	doi = strings.TrimSpace(doi)

	if doi == "" {
		return ""
	}

	return "https://doi.org/" + doi
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
