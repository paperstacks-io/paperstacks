// Package bibliography translates Paper domain values to bibliographic formats.
package bibliography

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

var ErrInvalidBibLaTeXEntryKey = errors.New("invalid BibLaTeX entry key")

// ExportBibLaTeX exports Papers as a deterministic BibLaTeX document.
// Papers are emitted in input order. A Paper's DOI is its entry key when
// available; otherwise its UUID is used. Duplicate DOI keys use the paper UUID.
func ExportBibLaTeX(papers []domain.Paper) ([]byte, error) {
	var out strings.Builder
	keys := make(map[string]struct{}, len(papers))

	for i, paper := range papers {
		key := paper.DOI
		if key == "" {
			key = paper.UUID
		}
		if _, exists := keys[key]; exists {
			key = paper.UUID
		}

		key, err := bibLaTeXEntryKey(key)
		if err != nil {
			return nil, fmt.Errorf("paper %d: %w", i, err)
		}
		keys[key] = struct{}{}

		if i > 0 {
			out.WriteByte('\n')
		}

		writeBibLaTeXEntry(&out, paper, key)
	}

	return []byte(out.String()), nil
}

func bibLaTeXEntryKey(key string) (string, error) {
	if key == "" {
		return "", ErrInvalidBibLaTeXEntryKey
	}
	for _, r := range key {
		if unicode.IsSpace(r) || r == ',' || r == '{' || r == '}' {
			return "", fmt.Errorf("%w: %q", ErrInvalidBibLaTeXEntryKey, key)
		}
	}

	return key, nil
}

func writeBibLaTeXEntry(out *strings.Builder, paper domain.Paper, key string) {
	fields := bibLaTeXFields(paper)

	out.WriteByte('@')
	out.WriteString(bibLaTeXEntryType(paper.Type))
	out.WriteByte('{')
	out.WriteString(key)
	if len(fields) == 0 {
		out.WriteString("}\n")
		return
	}

	out.WriteString(",\n")
	for i, field := range fields {
		out.WriteString("  ")
		out.WriteString(field.name)
		out.WriteString(" = {")
		out.WriteString(escapeBibLaTeX(field.value))
		out.WriteByte('}')
		if i < len(fields)-1 {
			out.WriteByte(',')
		}
		out.WriteByte('\n')
	}

	out.WriteString("}\n")
}

type bibLaTeXField struct {
	name  string
	value string
}

func bibLaTeXFields(paper domain.Paper) []bibLaTeXField {
	metadata := paper.Metadata
	fields := make([]bibLaTeXField, 0, 21)
	fields = appendBibLaTeXField(fields, "title", paper.Title)
	fields = appendBibLaTeXField(fields, "shorttitle", paper.TitleShort)
	fields = appendBibLaTeXField(fields, "author", bibLaTeXAuthors(paper.Authors))
	fields = appendBibLaTeXField(fields, "date", paper.PublicationDate.String())
	fields = appendBibLaTeXField(fields, "doi", paper.DOI)
	fields = appendBibLaTeXField(fields, "issn", strings.Join(nonEmpty(metadata.ISSN), ", "))
	fields = appendBibLaTeXField(fields, "keywords", strings.Join(nonEmpty(paper.Keywords), ", "))
	fields = appendBibLaTeXField(fields, "journaltitle", metadata.JournalTitle)
	fields = appendBibLaTeXField(fields, "shortjournal", metadata.JournalAbbrev)
	fields = appendBibLaTeXField(fields, "booktitle", metadata.BookTitle)
	fields = appendBibLaTeXField(fields, "series", metadata.SeriesTitle)
	fields = appendBibLaTeXField(fields, "eventtitle", metadata.EventTitle)
	fields = appendBibLaTeXField(fields, "location", metadata.EventLocation)
	fields = appendBibLaTeXField(fields, "institution", metadata.Institution)
	fields = appendBibLaTeXField(fields, "publisher", metadata.Publisher)
	fields = appendBibLaTeXField(fields, "volume", metadata.Volume)
	fields = appendBibLaTeXField(fields, "number", metadata.Issue)
	fields = appendBibLaTeXField(fields, "pages", metadata.Pages)
	fields = appendBibLaTeXField(fields, "isbn", firstNonEmpty(metadata.ISBN))
	fields = appendBibLaTeXField(fields, "url", firstHTTPURL(metadata.References))
	fields = appendBibLaTeXField(fields, "abstract", paper.Abstract)

	return fields
}

func appendBibLaTeXField(fields []bibLaTeXField, name string, value string) []bibLaTeXField {
	value = strings.TrimSpace(value)
	if name == "" || value == "" {
		return fields
	}

	return append(fields, bibLaTeXField{name: name, value: value})
}

func bibLaTeXEntryType(publicationType domain.PublicationType) string {
	switch publicationType {
	case domain.PublicationTypeConferenceArticle:
		return "inproceedings"
	case domain.PublicationTypeBook:
		return "book"
	case domain.PublicationTypeBookChapter:
		return "incollection"
	case domain.PublicationTypeThesis:
		return "thesis"
	case domain.PublicationTypeReport:
		return "report"
	case domain.PublicationTypeDataset:
		return "dataset"
	case domain.PublicationTypeWebPage:
		return "online"
	case domain.PublicationTypeJournalArticle, "":
		return "article"
	default:
		return "article"
	}
}

func bibLaTeXAuthors(authors []domain.Author) string {
	names := make([]string, 0, len(authors))
	for _, author := range authors {
		last := author.NameLast
		given := strings.Join(nonEmpty([]string{author.NameFirst, author.NameMiddle}), " ")

		switch {
		case last == "":
			names = append(names, given)
		case given == "":
			names = append(names, last)
		default:
			names = append(names, last+", "+given)
		}
	}

	return strings.Join(nonEmpty(names), " and ")
}

func nonEmpty(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			result = append(result, value)
		}
	}

	return result
}

func firstNonEmpty(values []string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}

func firstHTTPURL(references []string) string {
	for _, reference := range references {
		parsed, err := url.Parse(strings.TrimSpace(reference))
		if err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" {
			return parsed.String()
		}
	}

	return ""
}

func escapeBibLaTeX(value string) string {
	var escaped strings.Builder
	escaped.Grow(len(value))

	for _, r := range value {
		switch r {
		case '\\':
			escaped.WriteString(`\textbackslash{}`)
		case '{':
			escaped.WriteString(`\{`)
		case '}':
			escaped.WriteString(`\}`)
		case '&':
			escaped.WriteString(`\&`)
		case '%':
			escaped.WriteString(`\%`)
		case '$':
			escaped.WriteString(`\$`)
		case '#':
			escaped.WriteString(`\#`)
		case '_':
			escaped.WriteString(`\_`)
		case '~':
			escaped.WriteString(`\textasciitilde{}`)
		case '^':
			escaped.WriteString(`\textasciicircum{}`)
		default:
			escaped.WriteRune(r)
		}
	}

	return escaped.String()
}
