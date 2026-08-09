package acm

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

	return shared.JoinNonEmpty(
		" ",
		authors,
		year,
		title,
		source,
		doi,
	)
}

func (f Formatter) authors(
	authors []domain.Author,
) string {
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
		result = strings.Join(formatted, " and ")

	case 3:
		result = strings.Join(
			formatted[:2],
			", ",
		) + ", and " + formatted[2]

	default:
		result = formatted[0] + " et al."
	}

	return result + "."
}

func (f Formatter) author(
	author domain.Author,
) string {
	return shared.JoinNonEmpty(
		" ",
		author.NameFirst,
		author.NameMiddle,
		author.NameLast,
	)
}

func (f Formatter) year(year string) string {
	year = strings.TrimSpace(year)

	if year == "" {
		return ""
	}

	return year + "."
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
