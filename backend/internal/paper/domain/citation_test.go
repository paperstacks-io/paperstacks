package domain

import (
	"testing"
)

func TestPaperAPACitationWithoutAuthors(t *testing.T) {
	t.Parallel()

	paper := Paper{
		DOI:             "10.1000/example",
		Title:           "A paper without authors",
		PublicationYear: "2024",
		Metadata: Metadata{
			PublishedIn: "Journal of Software Testing",
			Volume:      "10",
			Issue:       "2",
			Pages:       "100-110",
		},
	}

	want := "A paper without authors. (2024). Journal of Software Testing, 10(2), 100–110. https://doi.org/10.1000/example"

	got := paper.APACitation()

	if got != want {
		t.Fatalf("APACitation() = %q, want %q", got, want)
	}
}

func TestPaperAPACitationWithMultipleAuthors(t *testing.T) {
	t.Parallel()

	paper := Paper{
		DOI:             "10.1016/j.infsof.2023.107299",
		Title:           "Code Review Guidelines for GUI-based Testing Artifacts",
		PublicationYear: "2023",
		Authors: []Author{
			{
				NameFirst: "Alexander",
				NameLast:  "Bauer",
			},
			{
				NameFirst: "Riccardo",
				NameLast:  "Coppola",
			},
			{
				NameFirst: "Emil",
				NameLast:  "Alégroth",
			},
			{
				NameFirst: "Tony",
				NameLast:  "Gorschek",
			},
		},
		Metadata: Metadata{
			PublishedIn: "Information and Software Technology",
			Volume:      "163",
			Pages:       "107299",
		},
	}

	want := "Bauer, A., Coppola, R., Alégroth, E., & Gorschek, T. (2023). Code Review Guidelines for GUI-based Testing Artifacts. Information and Software Technology, 163, 107299. https://doi.org/10.1016/j.infsof.2023.107299"

	got := paper.APACitation()

	if got != want {
		t.Fatalf("APACitation() = %q, want %q", got, want)
	}
}

func TestPaperAPACitationWithTwoAuthors(t *testing.T) {
	t.Parallel()

	paper := Paper{
		DOI:             "10.1111/j.2517-6161.1995.tb02031.x",
		Title:           "Controlling the false discovery Rate: A practical and powerful approach to multiple testing",
		PublicationYear: "1995",
		Authors: []Author{
			{
				NameFirst: "Yoav",
				NameLast:  "Benjamini",
			},
			{
				NameFirst: "Yosef",
				NameLast:  "Hochberg",
			},
		},
		Metadata: Metadata{
			PublishedIn: "Journal of the Royal Statistical Society Series B (Statistical Methodology)",
			Volume:      "57",
			Issue:       "1",
			Pages:       "289-300",
		},
	}

	want := "Benjamini, Y., & Hochberg, Y. (1995). Controlling the false discovery Rate: A practical and powerful approach to multiple testing. Journal of the Royal Statistical Society Series B (Statistical Methodology), 57(1), 289–300. https://doi.org/10.1111/j.2517-6161.1995.tb02031.x"

	got := paper.APACitation()

	if got != want {
		t.Fatalf("APACitation() = %q, want %q", got, want)
	}
}

func TestFormatAPAAuthorsWithMoreThanTwentyAuthors(t *testing.T) {
	t.Parallel()

	authors := []Author{
		{NameFirst: "Alice", NameLast: "Anderson"},
		{NameFirst: "Benjamin", NameLast: "Brown"},
		{NameFirst: "Clara", NameLast: "Clark"},
		{NameFirst: "Daniel", NameLast: "Davis"},
		{NameFirst: "Eva", NameLast: "Evans"},
		{NameFirst: "Felix", NameLast: "Fischer"},
		{NameFirst: "Grace", NameLast: "Green"},
		{NameFirst: "Henry", NameLast: "Harris"},
		{NameFirst: "Isabella", NameLast: "Irving"},
		{NameFirst: "Jack", NameLast: "Johnson"},
		{NameFirst: "Klara", NameLast: "Klein"},
		{NameFirst: "Liam", NameLast: "Lewis"},
		{NameFirst: "Maria", NameLast: "Miller"},
		{NameFirst: "Noah", NameLast: "Nelson"},
		{NameFirst: "Olivia", NameLast: "Owens"},
		{NameFirst: "Paul", NameLast: "Parker"},
		{NameFirst: "Quinn", NameLast: "Quinn"},
		{NameFirst: "Rosa", NameLast: "Roberts"},
		{NameFirst: "Samuel", NameLast: "Smith"},
		{NameFirst: "Theresa", NameLast: "Taylor"},
		{NameFirst: "Victor", NameLast: "Williams"},
	}

	want := "Anderson, A., Brown, B., Clark, C., Davis, D., Evans, E., Fischer, F., Green, G., Harris, H., Irving, I., Johnson, J., Klein, K., Lewis, L., Miller, M., Nelson, N., Owens, O., Parker, P., Quinn, Q., Roberts, R., Smith, S., … Williams, V."

	got := formatAPAAuthors(authors)

	if got != want {
		t.Fatalf("formatAPAAuthors() = %q, want %q", got, want)
	}
}

func TestPaperAPACitationWithoutPublicationYear(t *testing.T) {
	t.Parallel()

	paper := Paper{
		DOI:   "10.1000/example",
		Title: "An example paper without a publication year",
		Authors: []Author{
			{
				NameFirst: "Alice",
				NameLast:  "Miller",
			},
		},
		Metadata: Metadata{
			PublishedIn: "Example Journal",
			Volume:      "12",
			Issue:       "3",
			Pages:       "10-20",
		},
	}

	want := "Miller, A. (n.d.). An example paper without a publication year. Example Journal, 12(3), 10–20. https://doi.org/10.1000/example"

	got := paper.APACitation()

	if got != want {
		t.Fatalf("APACitation() = %q, want %q", got, want)
	}
}

func TestPaperAPACitationWithHyphenatedFirstName(t *testing.T) {
	t.Parallel()

	paper := Paper{
		DOI:             "10.1000/example",
		Title:           "Testing hyphenated author names",
		PublicationYear: "2024",
		Authors: []Author{
			{
				NameFirst: "Jean-Pierre",
				NameLast:  "Dupont",
			},
		},
		Metadata: Metadata{
			PublishedIn: "Journal of Software Testing",
			Volume:      "10",
			Issue:       "2",
			Pages:       "100-110",
		},
	}

	want := "Dupont, J.-P. (2024). Testing hyphenated author names. Journal of Software Testing, 10(2), 100–110. https://doi.org/10.1000/example"

	got := paper.APACitation()

	if got != want {
		t.Fatalf("APACitation() = %q, want %q", got, want)
	}
}

func TestPaperAPACitationWithoutDOI(t *testing.T) {
	t.Parallel()

	paper := Paper{
		Title:           "A paper without a DOI",
		PublicationYear: "2024",
		Authors: []Author{
			{
				NameFirst: "Alice",
				NameLast:  "Miller",
			},
		},
		Metadata: Metadata{
			PublishedIn: "Journal of Software Testing",
			Volume:      "10",
			Issue:       "2",
			Pages:       "100-110",
		},
	}

	want := "Miller, A. (2024). A paper without a DOI. Journal of Software Testing, 10(2), 100–110."

	got := paper.APACitation()

	if got != want {
		t.Fatalf("APACitation() = %q, want %q", got, want)
	}
}

func TestPaperAPACitationWithoutIssue(t *testing.T) {
	t.Parallel()

	paper := Paper{
		DOI:             "10.1000/example",
		Title:           "A paper without an issue",
		PublicationYear: "2024",
		Authors: []Author{
			{
				NameFirst: "Alice",
				NameLast:  "Miller",
			},
		},
		Metadata: Metadata{
			PublishedIn: "Journal of Software Testing",
			Volume:      "10",
			Pages:       "100-110",
		},
	}

	want := "Miller, A. (2024). A paper without an issue. Journal of Software Testing, 10, 100–110. https://doi.org/10.1000/example"

	got := paper.APACitation()

	if got != want {
		t.Fatalf("APACitation() = %q, want %q", got, want)
	}
}

func TestFormatAPASourceWithVolumeAndIssue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		metadata Metadata
		want     string
	}{
		{
			name: "volume and issue",
			metadata: Metadata{
				PublishedIn: "Journal of Software Testing",
				Volume:      "10",
				Issue:       "2",
				Pages:       "100-110",
			},
			want: "Journal of Software Testing, 10(2), 100–110.",
		},
		{
			name: "volume without issue",
			metadata: Metadata{
				PublishedIn: "Journal of Software Testing",
				Volume:      "10",
				Pages:       "100-110",
			},
			want: "Journal of Software Testing, 10, 100–110.",
		},
		{
			name: "issue without volume",
			metadata: Metadata{
				PublishedIn: "Journal of Software Testing",
				Issue:       "2",
				Pages:       "100-110",
			},
			want: "Journal of Software Testing, (2), 100–110.",
		},
		{
			name: "without volume and issue",
			metadata: Metadata{
				PublishedIn: "Journal of Software Testing",
				Pages:       "100-110",
			},
			want: "Journal of Software Testing, 100–110.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatAPASource(tt.metadata)

			if got != tt.want {
				t.Fatalf(
					"formatAPASource() = %q, want %q",
					got,
					tt.want,
				)
			}
		})
	}
}
