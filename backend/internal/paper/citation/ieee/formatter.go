package ieee

import (
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/paper/citation/shared"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

const maxAuthors = 6

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
	title := f.title(paper.Title)
	source := f.sources.Format(
		shared.NewSource(paper),
	)
	year := f.year(paper.PublicationYear)
	doi := f.doi(paper.DOI)

	var publication string

	//TODO: Find a better way to handle this
	switch paper.Type {
	case "inproceedings":
		publication = shared.JoinNonEmpty(
			", ",
			source,
			year,
			formatPages(paper.Metadata.Pages),
			doi,
		)
	default:
		publication = shared.JoinNonEmpty(
			", ",
			source,
			year,
			doi,
		)
	}

	citation := shared.JoinNonEmpty(
		" ",
		authors,
		title,
		publication,
	)

	citation = strings.TrimSpace(citation)

	if citation == "" {
		return ""
	}

	citation = strings.TrimSuffix(citation, ",")
	citation = strings.TrimSuffix(citation, ".")

	return citation + "."
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
		result = strings.Join(formatted, " and ")
	default:
		if len(formatted) > maxAuthors {
			result = formatted[0] + " et al."
		} else {
			result = strings.Join(
				formatted[:len(formatted)-1],
				", ",
			) + ", and " + formatted[len(formatted)-1]
		}
	}

	return result + ","
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

	return initials + " " + author.NameLast
}

func (f Formatter) year(year string) string {
	return strings.TrimSpace(year)
}

func (f Formatter) title(title string) string {
	title = strings.TrimSpace(title)

	if title == "" {
		return ""
	}

	if strings.HasSuffix(title, "?") ||
		strings.HasSuffix(title, "!") {
		return "“" + title + "”"
	}

	title = strings.TrimSuffix(title, ".")

	return "“" + title + ",”"
}

func (f Formatter) doi(doi string) string {
	doi = strings.TrimSpace(doi)

	if doi == "" {
		return ""
	}

	return "doi: " + doi
}
