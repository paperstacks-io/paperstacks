package bibliography

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

func TestCSLItemFromPaper(t *testing.T) {
	t.Parallel()

	paper := domain.Paper{
		UUID:              "64ecb626-96d0-4b5b-ba50-02ebd11506cf",
		DOI:               "10.1000/csl",
		Title:             "Testing CSL",
		TitleShort:        "CSL",
		Authors:           []domain.Author{{NameFirst: "Jane", NameMiddle: "Q.", NameLast: "Doe", Affiliation: "University", ORCID: "0000-0000-0000-0001"}},
		PublicationDate:   domain.Date{Year: 2024, Month: 3, Day: 14},
		PublicationStatus: "published",
		Abstract:          "Abstract",
		Keywords:          []string{"citation", "metadata"},
		Type:              domain.PublicationTypeJournalArticle,
		Metadata: domain.Metadata{
			Publisher:     "Publisher",
			JournalTitle:  "Journal",
			JournalAbbrev: "J.",
			SeriesTitle:   "Series",
			EventTitle:    "Conference",
			EventLocation: "Berlin",
			Pages:         "10-20",
			Volume:        "4",
			Issue:         "2",
			ISBN:          []string{"", "978-1-23456-789-0"},
			ISSN:          []string{"", "1234-5678"},
			References:    []string{"not a URL", "https://example.org/paper"},
		},
	}

	got := CSLItemFromPaper(paper)
	want := CSLItem{
		ID:                  paper.UUID,
		Type:                "article-journal",
		Title:               "Testing CSL",
		TitleShort:          "CSL",
		Author:              []CSLName{{Family: "Doe", Given: "Jane Q."}},
		Issued:              &CSLDate{DateParts: [][]int{{2024, 3, 14}}},
		Abstract:            "Abstract",
		Keyword:             "citation, metadata",
		Status:              "published",
		DOI:                 "10.1000/csl",
		ContainerTitle:      "Journal",
		ContainerTitleShort: "J.",
		CollectionTitle:     "Series",
		EventTitle:          "Conference",
		EventPlace:          "Berlin",
		Publisher:           "Publisher",
		Page:                "10-20",
		Volume:              "4",
		Issue:               "2",
		ISBN:                "978-1-23456-789-0",
		ISSN:                "1234-5678",
		URL:                 "https://example.org/paper",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CSLItemFromPaper() = %#v, want %#v", got, want)
	}

	document, err := json.Marshal([]CSLItem{got})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	wantJSON := `[{"id":"64ecb626-96d0-4b5b-ba50-02ebd11506cf","type":"article-journal","title":"Testing CSL","title-short":"CSL","author":[{"family":"Doe","given":"Jane Q."}],"issued":{"date-parts":[[2024,3,14]]},"abstract":"Abstract","keyword":"citation, metadata","status":"published","DOI":"10.1000/csl","container-title":"Journal","container-title-short":"J.","collection-title":"Series","event-title":"Conference","event-place":"Berlin","publisher":"Publisher","page":"10-20","volume":"4","issue":"2","ISBN":"978-1-23456-789-0","ISSN":"1234-5678","URL":"https://example.org/paper"}]`
	if string(document) != wantJSON {
		t.Fatalf("json.Marshal() = %s, want %s", document, wantJSON)
	}
}

func TestCSLItemFromPaperMapsPublicationType(t *testing.T) {
	t.Parallel()

	metadata := domain.Metadata{
		JournalTitle: "Journal",
		BookTitle:    "Proceedings",
		Publisher:    "Publisher",
		Institution:  "Institution",
	}
	tests := []struct {
		name          string
		publication   domain.PublicationType
		wantType      string
		wantContainer string
		wantPublisher string
	}{
		{name: "journal article", publication: domain.PublicationTypeJournalArticle, wantType: "article-journal", wantContainer: "Journal", wantPublisher: "Publisher"},
		{name: "conference article", publication: domain.PublicationTypeConferenceArticle, wantType: "paper-conference", wantContainer: "Proceedings", wantPublisher: "Publisher"},
		{name: "book", publication: domain.PublicationTypeBook, wantType: "book", wantPublisher: "Publisher"},
		{name: "book chapter", publication: domain.PublicationTypeBookChapter, wantType: "chapter", wantContainer: "Proceedings", wantPublisher: "Publisher"},
		{name: "thesis", publication: domain.PublicationTypeThesis, wantType: "thesis", wantPublisher: "Institution"},
		{name: "report", publication: domain.PublicationTypeReport, wantType: "report", wantPublisher: "Institution"},
		{name: "dataset", publication: domain.PublicationTypeDataset, wantType: "dataset", wantPublisher: "Publisher"},
		{name: "webpage", publication: domain.PublicationTypeWebPage, wantType: "webpage", wantPublisher: "Publisher"},
		{name: "default", wantType: "article", wantPublisher: "Publisher"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := CSLItemFromPaper(domain.Paper{Type: tt.publication, Metadata: metadata})
			if got.Type != tt.wantType || got.ContainerTitle != tt.wantContainer || got.Publisher != tt.wantPublisher {
				t.Fatalf("CSLItemFromPaper() type/container/publisher = %q/%q/%q, want %q/%q/%q", got.Type, got.ContainerTitle, got.Publisher, tt.wantType, tt.wantContainer, tt.wantPublisher)
			}
		})
	}
}

func TestCSLItemFromPaperPreservesDatePrecision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		date domain.Date
		want *CSLDate
	}{
		{name: "unknown"},
		{name: "year", date: domain.Date{Year: 2024}, want: &CSLDate{DateParts: [][]int{{2024}}}},
		{name: "month", date: domain.Date{Year: 2024, Month: 3}, want: &CSLDate{DateParts: [][]int{{2024, 3}}}},
		{name: "day", date: domain.Date{Year: 2024, Month: 3, Day: 14}, want: &CSLDate{DateParts: [][]int{{2024, 3, 14}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := CSLItemFromPaper(domain.Paper{PublicationDate: tt.date})
			if !reflect.DeepEqual(got.Issued, tt.want) {
				t.Fatalf("CSLItemFromPaper().Issued = %#v, want %#v", got.Issued, tt.want)
			}
		})
	}
}

func TestCSLItemFromPaperOmitsEmptyAuthors(t *testing.T) {
	t.Parallel()

	got := CSLItemFromPaper(domain.Paper{Authors: []domain.Author{
		{NameFirst: "Ada"},
		{NameLast: "Lovelace"},
		{},
	}})
	want := []CSLName{{Given: "Ada"}, {Family: "Lovelace"}}
	if !reflect.DeepEqual(got.Author, want) {
		t.Fatalf("CSLItemFromPaper().Author = %#v, want %#v", got.Author, want)
	}
}
