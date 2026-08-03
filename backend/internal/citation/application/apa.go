package application

import (
	"strings"

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
