package domain

import "testing"

func TestMetadataNormalizeTrimsISSN(t *testing.T) {
	t.Parallel()

	got := (Metadata{
		ISSN: []string{" 1234-5678 ", " 8765-4321 "},
	}).Normalize()
	want := []string{"1234-5678", "8765-4321"}

	if len(got.ISSN) != len(want) {
		t.Fatalf("normalized ISSN length = %d, want %d", len(got.ISSN), len(want))
	}
	for i := range want {
		if got.ISSN[i] != want[i] {
			t.Fatalf("normalized ISSN[%d] = %q, want %q", i, got.ISSN[i], want[i])
		}
	}
}
