package bibliography

import (
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

// CSLItem is a CSL-JSON input item.
type CSLItem struct {
	ID                  string    `json:"id"`
	Type                string    `json:"type"`
	Title               string    `json:"title,omitempty"`
	TitleShort          string    `json:"title-short,omitempty"`
	Author              []CSLName `json:"author,omitempty"`
	Issued              *CSLDate  `json:"issued,omitempty"`
	Abstract            string    `json:"abstract,omitempty"`
	Keyword             string    `json:"keyword,omitempty"`
	Status              string    `json:"status,omitempty"`
	DOI                 string    `json:"DOI,omitempty"`
	ContainerTitle      string    `json:"container-title,omitempty"`
	ContainerTitleShort string    `json:"container-title-short,omitempty"`
	CollectionTitle     string    `json:"collection-title,omitempty"`
	EventTitle          string    `json:"event-title,omitempty"`
	EventPlace          string    `json:"event-place,omitempty"`
	Publisher           string    `json:"publisher,omitempty"`
	Page                string    `json:"page,omitempty"`
	Volume              string    `json:"volume,omitempty"`
	Issue               string    `json:"issue,omitempty"`
	ISBN                string    `json:"ISBN,omitempty"`
	ISSN                string    `json:"ISSN,omitempty"`
	URL                 string    `json:"URL,omitempty"`
}

// CSLName is a CSL personal name.
type CSLName struct {
	Family string `json:"family,omitempty"`
	Given  string `json:"given,omitempty"`
}

// CSLDate is a CSL date with one date-parts value.
type CSLDate struct {
	DateParts [][]int `json:"date-parts"`
}

// CSLItemFromPaper converts a Paper to a CSL-JSON input item.
func CSLItemFromPaper(paper domain.Paper) CSLItem {
	metadata := paper.Metadata
	item := CSLItem{
		ID:                  paper.UUID,
		Type:                cslType(paper.Type),
		Title:               paper.Title,
		TitleShort:          paper.TitleShort,
		Author:              cslAuthors(paper.Authors),
		Issued:              cslDate(paper.PublicationDate),
		Abstract:            paper.Abstract,
		Keyword:             strings.Join(nonEmpty(paper.Keywords), ", "),
		Status:              paper.PublicationStatus,
		DOI:                 paper.DOI,
		ContainerTitle:      cslContainerTitle(paper),
		ContainerTitleShort: metadata.JournalAbbrev,
		CollectionTitle:     metadata.SeriesTitle,
		EventTitle:          metadata.EventTitle,
		EventPlace:          metadata.EventLocation,
		Publisher:           cslPublisher(paper),
		Page:                metadata.Pages,
		Volume:              metadata.Volume,
		Issue:               metadata.Issue,
		ISBN:                firstNonEmpty(metadata.ISBN),
		ISSN:                firstNonEmpty(metadata.ISSN),
		URL:                 firstHTTPURL(metadata.References),
	}

	return item
}

func cslType(publicationType domain.PublicationType) string {
	switch publicationType {
	case domain.PublicationTypeJournalArticle:
		return "article-journal"
	case domain.PublicationTypeConferenceArticle:
		return "paper-conference"
	case domain.PublicationTypeBook:
		return "book"
	case domain.PublicationTypeBookChapter:
		return "chapter"
	case domain.PublicationTypeThesis:
		return "thesis"
	case domain.PublicationTypeReport:
		return "report"
	case domain.PublicationTypeDataset:
		return "dataset"
	case domain.PublicationTypeWebPage:
		return "webpage"
	default:
		return "article"
	}
}

func cslAuthors(authors []domain.Author) []CSLName {
	result := make([]CSLName, 0, len(authors))
	for _, author := range authors {
		name := CSLName{
			Family: author.NameLast,
			Given:  strings.Join(nonEmpty([]string{author.NameFirst, author.NameMiddle}), " "),
		}
		if name.Family != "" || name.Given != "" {
			result = append(result, name)
		}
	}

	return result
}

func cslDate(date domain.Date) *CSLDate {
	if date.IsZero() {
		return nil
	}

	parts := []int{date.Year}
	if date.Month != 0 {
		parts = append(parts, date.Month)
	}
	if date.Day != 0 {
		parts = append(parts, date.Day)
	}

	return &CSLDate{DateParts: [][]int{parts}}
}

func cslContainerTitle(paper domain.Paper) string {
	switch paper.Type {
	case domain.PublicationTypeJournalArticle:
		return paper.Metadata.JournalTitle
	case domain.PublicationTypeConferenceArticle, domain.PublicationTypeBookChapter:
		return paper.Metadata.BookTitle
	default:
		return ""
	}
}

func cslPublisher(paper domain.Paper) string {
	if paper.Type == domain.PublicationTypeThesis || paper.Type == domain.PublicationTypeReport {
		return firstNonEmpty([]string{paper.Metadata.Institution, paper.Metadata.Publisher})
	}

	return paper.Metadata.Publisher
}
