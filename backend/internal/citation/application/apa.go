package application

import (
	"strings"
	"unicode"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

func FormatAPA(paper domain.Paper) string {
	authors := formatAPAAuthors(paper.Authors)
	year := formatAPAYear(paper.PublicationYear)
	title := formatAPATitle(paper.Title)
	source := formatAPASource(paper.Metadata)
	doi := formatAPADOI(paper.DOI)

	if authors == "" {
		return joinCitationParts(title, year, source, doi)
	}

	return joinCitationParts(authors, year, title, source, doi)
}

func formatAPAAuthors(authors []domain.Author) string {
	formatted := make([]string, 0, len(authors))

	for _, author := range authors {
		if name := formatAPAAuthor(author); name != "" {
			formatted = append(formatted, name)
		}
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
			", … " +
			formatted[len(formatted)-1]
	}

	return strings.Join(formatted[:len(formatted)-1], ", ") +
		", & " +
		formatted[len(formatted)-1]
}

func formatAPAAuthor(author domain.Author) string {
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

	return author.NameLast + ", " + initials
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

func formatAPAYear(year string) string {
	if year = strings.TrimSpace(year); year == "" {
		year = "n.d."
	}

	return "(" + year + ")."
}

func formatAPATitle(title string) string {
	title = strings.TrimSpace(title)

	if title == "" {
		return ""
	}

	if strings.HasSuffix(title, ".") ||
		strings.HasSuffix(title, "?") ||
		strings.HasSuffix(title, "!") {
		return title
	}

	return title + "."
}

func formatAPASource(metadata domain.Metadata) string {
	publishedIn := strings.TrimSpace(metadata.PublishedIn)
	volume := strings.TrimSpace(metadata.Volume)
	issue := strings.TrimSpace(metadata.Issue)
	pages := strings.ReplaceAll(
		strings.TrimSpace(metadata.Pages),
		"-",
		"–",
	)

	var publicationInfo string

	switch {
	case volume != "" && issue != "":
		publicationInfo = volume + "(" + issue + ")"
	case volume != "":
		publicationInfo = volume
	case issue != "":
		publicationInfo = "(" + issue + ")"
	}

	source := joinNonEmpty(
		", ",
		publishedIn,
		publicationInfo,
		pages,
	)

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
