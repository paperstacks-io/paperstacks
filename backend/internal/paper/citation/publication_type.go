package citation

import "github.com/paperstacks.io/paperstacks/internal/paper/domain"

type PublicationType string

func publicationType(t domain.PublicationType) string {
	switch t {
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
		return ""
	}
}
