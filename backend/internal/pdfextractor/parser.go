package pdfextractor

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"

	paper "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/pdfextractor/tei"
)

// ParseTEI decodes a GROBID TEI-XML document from r and maps it to paper.Paper.
func ParseTEI(r io.Reader) (paper.Paper, error) {
	if r == nil {
		return paper.Paper{}, fmt.Errorf("parse TEI XML: nil reader")
	}

	var doc tei.Document
	if err := xml.NewDecoder(r).Decode(&doc); err != nil {
		return paper.Paper{}, fmt.Errorf("decode TEI XML: %w", err)
	}

	return PaperFromDocument(doc), nil
}

// PaperFromDocument maps a decoded tei.Document (title, authors, abstract, DOI,
// publication date, keywords, venue, and references) to the internal paper.Paper
// domain model.
func PaperFromDocument(doc tei.Document) paper.Paper {
	bibl := doc.Header.FileDesc.SourceDesc.BiblStruct

	title := firstTitle(doc.Header.FileDesc.TitleStmt.Titles)
	if title == "" {
		title = firstTitle(bibl.Analytic.Titles)
	}

	doi := findIdno(bibl.Analytic.Idnos, "DOI")
	if doi == "" {
		doi = findIdno(bibl.Monogr.Idnos, "DOI")
	}
	if doi == "" {
		doi = findIdno(bibl.Idnos, "DOI")
	}

	pubDate := parseTEIDates(doc.Header.FileDesc.PublicationStmt.Dates)
	if pubDate.Year == 0 {
		pubDate = parseTEIDates(bibl.Monogr.Imprint.Dates)
	}

	authors := mapAuthors(bibl.Analytic.Authors)
	if len(authors) == 0 {
		authors = mapAuthors(bibl.Monogr.Authors)
	}

	publisher := strings.TrimSpace(bibl.Monogr.Imprint.Publisher)
	if publisher == "" {
		publisher = strings.TrimSpace(doc.Header.FileDesc.PublicationStmt.Publisher)
	}

	journalTitle, journalAbbrev := firstJournalTitles(bibl.Monogr.Titles)
	volume, issue, pages := parseBiblScopes(bibl.Monogr.Imprint.BiblScopes)

	return paper.Paper{
		DOI:             doi,
		Title:           title,
		Authors:         authors,
		PublicationDate: pubDate,
		Abstract:        extractAbstract(doc.Header.ProfileDesc.Abstract),
		Keywords:        extractKeywords(doc.Header.ProfileDesc.TextClass.Keywords),
		Type:            paper.PublicationTypeJournalArticle,
		Metadata: paper.Metadata{
			Publisher:     publisher,
			JournalTitle:  journalTitle,
			JournalAbbrev: journalAbbrev,
			Volume:        volume,
			Issue:         issue,
			Pages:         pages,
			ISSN:          findAllIdnos(bibl.Monogr.Idnos, "ISSN"),
			ISBN:          findAllIdnos(bibl.Monogr.Idnos, "ISBN"),
			References:    extractReferences(doc.Text.Back),
			DataSource:    "grobid_extraction",
		},
	}.Normalize()
}

// MissingFields returns the names of any core metadata fields (title, authors,
// abstract, doi, publication_date, keywords) that are missing from p.
func MissingFields(p paper.Paper) []string {
	var missing []string
	if strings.TrimSpace(p.Title) == "" {
		missing = append(missing, "title")
	}
	if len(p.Authors) == 0 {
		missing = append(missing, "authors")
	}
	if strings.TrimSpace(p.Abstract) == "" {
		missing = append(missing, "abstract")
	}
	if strings.TrimSpace(p.DOI) == "" {
		missing = append(missing, "doi")
	}
	if p.PublicationDate.Year == 0 {
		missing = append(missing, "publication_date")
	}
	if len(p.Keywords) == 0 {
		missing = append(missing, "keywords")
	}
	return missing
}

func firstTitle(titles []tei.Title) string {
	for _, t := range titles {
		if strings.EqualFold(t.Type, "main") && strings.TrimSpace(t.Value) != "" {
			return strings.TrimSpace(t.Value)
		}
	}
	for _, t := range titles {
		if strings.TrimSpace(t.Value) != "" {
			return strings.TrimSpace(t.Value)
		}
	}
	return ""
}

func firstJournalTitles(titles []tei.Title) (mainTitle, abbrev string) {
	for _, t := range titles {
		val := strings.TrimSpace(t.Value)
		if val == "" {
			continue
		}
		if strings.EqualFold(t.Type, "abbrev") {
			if abbrev == "" {
				abbrev = val
			}
			continue
		}
		if mainTitle == "" {
			mainTitle = val
		}
	}
	return mainTitle, abbrev
}

func findIdno(idnos []tei.Idno, kind string) string {
	for _, id := range idnos {
		if strings.EqualFold(id.Type, kind) && strings.TrimSpace(id.Value) != "" {
			return strings.TrimSpace(id.Value)
		}
	}
	return ""
}

func findAllIdnos(idnos []tei.Idno, kind string) []string {
	var out []string
	for _, id := range idnos {
		if strings.EqualFold(id.Type, kind) {
			if v := strings.TrimSpace(id.Value); v != "" {
				out = append(out, v)
			}
		}
	}
	return out
}

func mapAuthors(raw []tei.Author) []paper.Author {
	var authors []paper.Author
	for _, a := range raw {
		var first, middle string
		for _, fn := range a.PersName.Forenames {
			val := strings.TrimSpace(fn.Value)
			if val == "" {
				continue
			}
			switch strings.ToLower(fn.Type) {
			case "first":
				if first == "" {
					first = val
				}
			case "middle":
				if middle == "" {
					middle = val
				} else {
					middle += " " + val
				}
			default:
				if first == "" {
					first = val
				} else if middle == "" {
					middle = val
				}
			}
		}
		last := strings.TrimSpace(a.PersName.Surname)
		if first == "" && last == "" {
			continue
		}

		var affParts []string
		for _, aff := range a.Affiliations {
			var orgs []string
			for _, org := range aff.OrgNames {
				if o := strings.TrimSpace(org); o != "" {
					orgs = append(orgs, o)
				}
			}
			if len(orgs) > 0 {
				affParts = append(affParts, strings.Join(orgs, ", "))
			}
		}

		authors = append(authors, paper.Author{
			NameFirst:   first,
			NameMiddle:  middle,
			NameLast:    last,
			Affiliation: strings.Join(affParts, "; "),
			ORCID:       findIdno(a.Idnos, "ORCID"),
		})
	}
	return authors
}

func parseTEIDates(dates []tei.Date) paper.Date {
	for _, d := range dates {
		if strings.EqualFold(d.Type, "published") || d.Type == "" {
			if parsed := parseISODate(d.When); parsed.Year != 0 {
				return parsed
			}
		}
	}
	for _, d := range dates {
		if parsed := parseISODate(d.When); parsed.Year != 0 {
			return parsed
		}
	}
	return paper.Date{}
}

func parseISODate(s string) paper.Date {
	s = strings.TrimSpace(s)
	if s == "" {
		return paper.Date{}
	}
	parts := strings.Split(s, "-")
	year, err := strconv.Atoi(parts[0])
	if err != nil || year < 1 {
		return paper.Date{}
	}
	var month, day int
	if len(parts) >= 2 {
		if m, err := strconv.Atoi(parts[1]); err == nil && m >= 1 && m <= 12 {
			month = m
		}
	}
	if len(parts) >= 3 && month != 0 {
		if d, err := strconv.Atoi(parts[2]); err == nil && d >= 1 && d <= 31 {
			day = d
		}
	}
	return paper.Date{Year: year, Month: month, Day: day}
}

func parseBiblScopes(scopes []tei.BiblScope) (volume, issue, pages string) {
	for _, s := range scopes {
		val := strings.TrimSpace(s.Value)
		switch strings.ToLower(s.Unit) {
		case "volume":
			if volume == "" {
				volume = val
			}
		case "issue":
			if issue == "" {
				issue = val
			}
		case "page":
			if pages == "" {
				from := strings.TrimSpace(s.From)
				to := strings.TrimSpace(s.To)
				switch {
				case from != "" && to != "" && from != to:
					pages = from + "-" + to
				case from != "":
					pages = from
				default:
					pages = val
				}
			}
		}
	}
	return volume, issue, pages
}

func extractAbstract(abs tei.Abstract) string {
	var paragraphs []string
	for _, p := range abs.Paragraphs {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			paragraphs = append(paragraphs, trimmed)
		}
	}
	for _, div := range abs.Divs {
		for _, p := range div.Paragraphs {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				paragraphs = append(paragraphs, trimmed)
			}
		}
	}
	return strings.Join(paragraphs, "\n\n")
}

func extractKeywords(kw tei.Keywords) []string {
	var out []string
	for _, t := range kw.Terms {
		if trimmed := strings.TrimSpace(t); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 && strings.TrimSpace(kw.Raw) != "" {
		for _, part := range strings.FieldsFunc(kw.Raw, func(r rune) bool {
			return r == ',' || r == ';' || r == '\n'
		}) {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				out = append(out, trimmed)
			}
		}
	}
	return out
}

func extractReferences(back tei.Back) []string {
	items := back.ListBibl
	for _, div := range back.Divs {
		items = append(items, div.ListBibl...)
	}

	var refs []string
	for _, item := range items {
		doi := findIdno(item.Analytic.Idnos, "DOI")
		if doi == "" {
			doi = findIdno(item.Monogr.Idnos, "DOI")
		}
		if doi == "" {
			doi = findIdno(item.Idnos, "DOI")
		}
		if doi != "" {
			refs = append(refs, doi)
			continue
		}
		title := firstTitle(item.Analytic.Titles)
		if title == "" {
			title = firstTitle(item.Monogr.Titles)
		}
		if title != "" {
			refs = append(refs, title)
		}
	}
	return refs
}
