package parser

// rawRecord mirrors the JSON shape of a single Crossref "work" record as
// found in the Crossref public data export. Field names follow the
// export's schema; see https://api.crossref.org/swagger-ui/index.html
// for the reference.
type rawRecord struct {
	DOI                 string           `json:"DOI"`
	Type                string           `json:"type"`
	Source              string           `json:"source"`
	Title               []string         `json:"title"`
	Subtitle            []string         `json:"subtitle"`
	ContainerTitle      []string         `json:"container-title"`
	ShortContainerTitle []string         `json:"short-container-title"`
	Abstract            string           `json:"abstract"`
	Publisher           string           `json:"publisher"`
	PublisherLocation   string           `json:"publisher-location"`
	Member              string           `json:"member"`
	Prefix              string           `json:"prefix"`
	Language            string           `json:"language"`
	URL                 string           `json:"URL"`
	Resource            rawResource      `json:"resource"`
	Volume              string           `json:"volume"`
	Issue               string           `json:"issue"`
	Page                string           `json:"page"`
	SpecialNumbering    string           `json:"special_numbering"`
	ISSN                []string         `json:"ISSN"`
	ISSNType            []rawTypedID     `json:"issn-type"`
	ISBN                []string         `json:"ISBN"`
	ISBNType            []rawTypedID     `json:"isbn-type"`
	Author              []rawContributor `json:"author"`
	Editor              []rawContributor `json:"editor"`
	Funder              []rawFunder      `json:"funder"`
	Reference           []rawReference   `json:"reference"`
	License             []rawLicense     `json:"license"`
	AlternativeID       []string         `json:"alternative-id"`
	Issued              rawDateField     `json:"issued"`
	Published           rawDateField     `json:"published"`
	PublishedPrint      rawDateField     `json:"published-print"`
	PublishedOnline     rawDateField     `json:"published-online"`
	Created             rawDateField     `json:"created"`
	Deposited           rawDateField     `json:"deposited"`
	Indexed             rawDateField     `json:"indexed"`
	ReferenceCount      int              `json:"reference-count"`
	ReferencesCount     int              `json:"references-count"`
	IsReferencedByCount int              `json:"is-referenced-by-count"`
}

type rawResource struct {
	Primary struct {
		URL string `json:"URL"`
	} `json:"primary"`
}

// rawTypedID is a value bound to a medium type, used for both
// "issn-type" and "isbn-type" entries (e.g. {"type": "print", "value": "..."}).
type rawTypedID struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type rawContributor struct {
	Given       string           `json:"given"`
	Family      string           `json:"family"`
	Sequence    string           `json:"sequence"`
	ORCID       string           `json:"ORCID"`
	Affiliation []rawAffiliation `json:"affiliation"`
}

type rawAffiliation struct {
	Name string `json:"name"`
}

type rawFunder struct {
	Name  string   `json:"name"`
	DOI   string   `json:"DOI"`
	Award []string `json:"award"`
}

type rawReference struct {
	Key          string `json:"key"`
	DOI          string `json:"DOI"`
	ArticleTitle string `json:"article-title"`
	Author       string `json:"author"`
	JournalTitle string `json:"journal-title"`
	Volume       string `json:"volume"`
	FirstPage    string `json:"first-page"`
	Year         string `json:"year"`
	Unstructured string `json:"unstructured"`
}

type rawLicense struct {
	URL            string       `json:"URL"`
	ContentVersion string       `json:"content-version"`
	DelayInDays    int          `json:"delay-in-days"`
	Start          rawDateField `json:"start"`
}

// rawDateField is Crossref's common date representation, used for
// "issued", "published", "created", "deposited", "indexed" and similar
// fields. DateParts holds [[year, month, day]] with trailing components
// omitted when unknown.
type rawDateField struct {
	DateParts [][]int `json:"date-parts"`
	DateTime  string  `json:"date-time"`
	Timestamp *int64  `json:"timestamp"`
}
