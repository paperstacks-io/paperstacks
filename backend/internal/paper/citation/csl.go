package citation

import (
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

// https://github.com/citation-style-language/schema/blob/master/schemas/input/csl-data.json
type CSLItem struct {
	Type           string   `json:"type"`
	Title          string   `json:"title"`
	Author         []Author `json:"author"`
	Issued         Date     `json:"issued"`
	ContainerTitle string   `json:"container-title"`
	Volume         string   `json:"volume"`
	Issue          string   `json:"issue"`
	Page           string   `json:"page"`
	Publisher      string   `json:"publisher"`
	DOI            string   `json:"DOI"`
	ISBN           string   `json:"ISBN"`
	EventTitle     string   `json:"event-title"`
	EventPlace     string   `json:"event-place"`
}

func NewCSLItem(paper domain.Paper) CSLItem {
	paper = cslNormalize(paper)

	authors := make([]Author, 0, len(paper.Authors))

	for _, author := range paper.Authors {
		authors = append(authors, Author{
			Given:  given(author.NameFirst, author.NameMiddle),
			Family: author.NameLast,
		})
	}

	return CSLItem{
		Type:           publicationType(paper.Type),
		Title:          paper.Title,
		Author:         authors,
		Issued:         date(paper.PublicationDate),
		ContainerTitle: containerTitle(paper),
		Volume:         paper.Metadata.Volume,
		Issue:          paper.Metadata.Issue,
		Page:           paper.Metadata.Pages,
		Publisher:      paper.Metadata.Publisher,
		DOI:            paper.DOI,
		EventTitle:     paper.Metadata.EventTitle,
		EventPlace:     paper.Metadata.EventLocation,
	}
}

func containerTitle(paper domain.Paper) string {
	switch paper.Type {
	case domain.PublicationTypeJournalArticle:
		return paper.Metadata.JournalTitle
	case domain.PublicationTypeConferenceArticle,
		domain.PublicationTypeBookChapter:
		return paper.Metadata.BookTitle
	default:
		return ""
	}
}

func cslNormalize(paper domain.Paper) domain.Paper {
	paper.Metadata.Pages = strings.ReplaceAll(
		paper.Metadata.Pages,
		"--",
		"-",
	)

	return paper
}
