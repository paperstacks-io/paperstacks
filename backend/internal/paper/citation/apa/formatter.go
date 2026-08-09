package apa

import (
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/paper/citation/shared"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

type Formatter struct {
	sources shared.SourceRegistry
}

func NewFormatter() Formatter {
	return Formatter{
		sources: shared.NewSourceRegistry(
			map[string]shared.SourceFormatter{
				"article":       formatJournalSource,
				"inproceedings": formatConferenceSource,
			},
		),
	}
}

func (f Formatter) Format(paper domain.Paper) string {
	authors := f.authors(paper.Authors)
	year := f.year(paper.PublicationYear)
	title := f.title(paper.Title)
	source := f.sources.Format(
		shared.NewSource(paper),
	)
	doi := f.doi(paper.DOI)

	if authors == "" {
		return shared.JoinNonEmpty(
			" ",
			title,
			year,
			source,
			doi,
		)
	}

	return shared.JoinNonEmpty(
		" ",
		authors,
		year,
		title,
		source,
		doi,
	)
}

func (f Formatter) authors(authors []domain.Author) string {
	formatted := make([]string, 0, len(authors))

	for _, author := range authors {
		if name := f.author(author); name != "" {
			formatted = append(formatted, name)
		}
	}

	var result string

	switch len(formatted) {
	case 0:
		return ""

	case 1:
		result = formatted[0]

	case 2:
		result = strings.Join(formatted, ", & ")

	default:
		if len(formatted) > 20 {
			result = strings.Join(
				formatted[:19],
				", ",
			) + ", . . . " + formatted[len(formatted)-1]
		} else {
			result = strings.Join(
				formatted[:len(formatted)-1],
				", ",
			) + ", & " + formatted[len(formatted)-1]
		}
	}

	if strings.HasSuffix(result, ".") {
		return result
	}

	return result + "."
}

func (f Formatter) author(author domain.Author) string {
	if author.NameLast == "" {
		return ""
	}

	initials := shared.FormatInitials(
		author.NameFirst,
		author.NameMiddle,
	)

	if initials == "" {
		return author.NameLast
	}

	return author.NameLast + ", " + initials
}

func (f Formatter) year(year string) string {
	year = strings.TrimSpace(year)

	if year == "" {
		year = "n.d."
	}

	return "(" + year + ")."
}

func (f Formatter) title(title string) string {
	title = strings.TrimSpace(title)

	if title == "" {
		return ""
	}

	if strings.HasSuffix(title, "?") ||
		strings.HasSuffix(title, "!") ||
		strings.HasSuffix(title, ".") {
		return title
	}

	return title + "."
}

func (f Formatter) doi(doi string) string {
	doi = strings.TrimSpace(doi)

	if doi == "" {
		return ""
	}

	return "https://doi.org/" + doi
}
