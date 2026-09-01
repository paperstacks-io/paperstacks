package domain

import "testing"

func TestPublicationTypeIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		type_ PublicationType
		want  bool
	}{
		{name: "unspecified", type_: "", want: true},
		{name: "journal article", type_: PublicationTypeJournalArticle, want: true},
		{name: "conference article", type_: PublicationTypeConferenceArticle, want: true},
		{name: "book", type_: PublicationTypeBook, want: true},
		{name: "book chapter", type_: PublicationTypeBookChapter, want: true},
		{name: "thesis", type_: PublicationTypeThesis, want: true},
		{name: "report", type_: PublicationTypeReport, want: true},
		{name: "preprint status", type_: "preprint", want: false},
		{name: "dataset", type_: PublicationTypeDataset, want: true},
		{name: "web page", type_: PublicationTypeWebPage, want: true},
		{name: "unknown", type_: "article", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.type_.IsValid(); got != tt.want {
				t.Fatalf("PublicationType.IsValid() = %t, want %t", got, tt.want)
			}
		})
	}
}
