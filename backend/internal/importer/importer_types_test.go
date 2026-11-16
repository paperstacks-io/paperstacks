package importer

import (
	"testing"

	crossref "github.com/paperstacks.io/paperstacks/internal/importer/crossref/domain"
	paper "github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

func TestToPaperMapsCrossrefTypes(t *testing.T) {
	tests := []struct {
		crossrefType string
		want         paper.PublicationType
	}{
		{"journal-article", paper.PublicationTypeJournalArticle},
		{"proceedings-article", paper.PublicationTypeConferenceArticle},
		{"book", paper.PublicationTypeBook},
		{"dissertation", paper.PublicationTypeThesis},
		{"posted-content", ""},
	}
	for _, tt := range tests {
		t.Run(tt.crossrefType, func(t *testing.T) {
			got := toPaper(&crossref.Paper{Type: tt.crossrefType}).Type
			if got != tt.want {
				t.Errorf("toPaper(type %q).Type = %q, want %q", tt.crossrefType, got, tt.want)
			}
		})
	}
}
