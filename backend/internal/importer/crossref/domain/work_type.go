package domain

// WorkType is a Crossref work type ID as used in the "type" field and the
// "type" API filter (see https://api.crossref.org/types).
type WorkType string

const (
	WorkTypeJournalArticle     WorkType = "journal-article"
	WorkTypeProceedingsArticle WorkType = "proceedings-article"
	WorkTypeBook               WorkType = "book"
	WorkTypeDissertation       WorkType = "dissertation"
)

// ImportedWorkTypes is the whitelist of Crossref work types Paperstacks
// imports; works of any other type are skipped.
var ImportedWorkTypes = []WorkType{
	WorkTypeJournalArticle,
	WorkTypeProceedingsArticle,
	WorkTypeBook,
	WorkTypeDissertation,
}
