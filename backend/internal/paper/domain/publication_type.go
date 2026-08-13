package domain

// PublicationType identifies the bibliographic kind of a Paper.
type PublicationType string

const (
	PublicationTypeJournalArticle    PublicationType = "journal-article"
	PublicationTypeConferenceArticle PublicationType = "conference-article"
	PublicationTypeBook              PublicationType = "book"
	PublicationTypeBookChapter       PublicationType = "book-chapter"
	PublicationTypeThesis            PublicationType = "thesis"
	PublicationTypeReport            PublicationType = "report"
	PublicationTypeDataset           PublicationType = "dataset"
	PublicationTypeWebPage           PublicationType = "webpage"
)

func (t PublicationType) IsValid() bool {
	switch t {
	case "",
		PublicationTypeJournalArticle,
		PublicationTypeConferenceArticle,
		PublicationTypeBook,
		PublicationTypeBookChapter,
		PublicationTypeThesis,
		PublicationTypeReport,
		PublicationTypeDataset,
		PublicationTypeWebPage:
		return true
	default:
		return false
	}
}
