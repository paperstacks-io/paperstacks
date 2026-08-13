package domain

import "testing"

func TestMetadataNormalizeTrimsContainerFields(t *testing.T) {
	t.Parallel()

	got := (Metadata{
		JournalTitle:  " Journal Title ",
		JournalAbbrev: " J. Title ",
		BookTitle:     " Book Title ",
		SeriesTitle:   " Series Title ",
		EventTitle:    " Event Title ",
		EventLocation: " Event Place ",
		Institution:   " Example University ",
		ISSN:          []string{" 1234-5678 ", " 8765-4321 "},
	}).Normalize()

	if got.JournalTitle != "Journal Title" || got.JournalAbbrev != "J. Title" || got.BookTitle != "Book Title" || got.SeriesTitle != "Series Title" || got.EventTitle != "Event Title" || got.EventLocation != "Event Place" || got.Institution != "Example University" {
		t.Fatalf("normalized container metadata = %#v", got)
	}

	wantISSN := []string{"1234-5678", "8765-4321"}
	if len(got.ISSN) != len(wantISSN) {
		t.Fatalf("normalized ISSN length = %d, want %d", len(got.ISSN), len(wantISSN))
	}
	for i := range wantISSN {
		if got.ISSN[i] != wantISSN[i] {
			t.Fatalf("normalized ISSN[%d] = %q, want %q", i, got.ISSN[i], wantISSN[i])
		}
	}
}
