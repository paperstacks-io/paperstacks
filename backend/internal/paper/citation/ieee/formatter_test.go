package ieee

import (
	"context"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
)

func TestFormatterWithoutAuthors(t *testing.T) {
	t.Parallel()

	paper := domain.Paper{
		Type:            "article",
		DOI:             "10.1000/example",
		Title:           "A paper without authors",
		PublicationYear: "2024",
		Metadata: domain.Metadata{
			PublishedIn: "Journal of Software Testing",
			Volume:      "10",
			Issue:       "2",
			Pages:       "100-110",
		},
	}

	want := "“A paper without authors,” " +
		"Journal of Software Testing, vol. 10, no. 2, " +
		"pp. 100–110, 2024, doi: 10.1000/example."

	got := NewFormatter().Format(paper)

	if got != want {
		t.Fatalf(
			"Formatter.Format() = %q, want %q",
			got,
			want,
		)
	}
}

func TestFormatterWithMultipleAuthors(t *testing.T) {
	t.Parallel()

	repository := memory.NewRepository()

	paper, err := repository.GetByDOI(
		context.Background(),
		"10.1016/j.infsof.2023.107299",
	)
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}

	want := "A. Bauer, R. Coppola, E. Alégroth, and T. Gorschek, " +
		"“Code Review Guidelines for GUI-based Testing Artifacts,” " +
		"Information and Software Technology, vol. 163, p. 107299, " +
		"2023, doi: 10.1016/j.infsof.2023.107299."

	got := NewFormatter().Format(paper)

	if got != want {
		t.Fatalf(
			"Formatter.Format() = %q, want %q",
			got,
			want,
		)
	}
}

func TestFormatterWithTwoAuthors(t *testing.T) {
	t.Parallel()

	repository := memory.NewRepository()

	paper, err := repository.GetByDOI(
		context.Background(),
		"10.1016/j.infsof.2024.107611",
	)
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}

	want := "K. Petersen and J. M. Gerken, " +
		"“On the Road to Interactive LLM-based Systematic Mapping Studies,” " +
		"Information and Software Technology, p. 107611, " +
		"2024, doi: 10.1016/j.infsof.2024.107611."

	got := NewFormatter().Format(paper)

	if got != want {
		t.Fatalf(
			"Formatter.Format() = %q, want %q",
			got,
			want,
		)
	}
}

func TestFormatterAuthorsWithMoreThanSixAuthors(t *testing.T) {
	t.Parallel()

	repository := memory.NewRepository()

	paper, err := repository.GetByDOI(
		context.Background(),
		"10.48550/ARXIV.2010.03525",
	)
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}

	want := "P. Ralph et al.,"

	formatter := NewFormatter()
	got := formatter.authors(paper.Authors)

	if got != want {
		t.Fatalf(
			"Formatter.authors() = %q, want %q",
			got,
			want,
		)
	}
}

func TestFormatterWithoutPublicationYear(t *testing.T) {
	t.Parallel()

	repository := memory.NewRepository()

	paper, err := repository.GetByUUID(
		context.Background(),
		"5966a651-6ed5-5bbc-bfe0-c76331d373a6",
	)
	if err != nil {
		t.Fatalf("GetByUUID() error = %v", err)
	}

	want := "M. Finsterwalder, " +
		"“Automating Acceptance Tests for GUI Applications in an Extreme Programming Environment,” " +
		"p. 4, doi: 10.1000/empty."

	got := NewFormatter().Format(paper)

	if got != want {
		t.Fatalf(
			"Formatter.Format() = %q, want %q",
			got,
			want,
		)
	}
}

func TestFormatterWithHyphenatedFirstName(t *testing.T) {
	t.Parallel()

	repository := memory.NewRepository()

	paper, err := repository.GetByDOI(
		context.Background(),
		"10.18637/jss.v080.i01",
	)
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}

	want := "P.-C. Bürkner, " +
		"“Brms: An R Package for Bayesian Multilevel Models Using Stan,” " +
		"Journal of Statistical Software, vol. 80, no. 1, pp. 1–28, " +
		"2017, doi: 10.18637/jss.v080.i01."

	got := NewFormatter().Format(paper)

	if got != want {
		t.Fatalf(
			"Formatter.Format() = %q, want %q",
			got,
			want,
		)
	}
}

func TestFormatterWithoutDOI(t *testing.T) {
	t.Parallel()

	paper := domain.Paper{
		Type:            "article",
		Title:           "A paper without a DOI",
		PublicationYear: "2024",
		Authors: []domain.Author{
			{
				NameFirst: "Alice",
				NameLast:  "Miller",
			},
		},
		Metadata: domain.Metadata{
			PublishedIn: "Journal of Software Testing",
			Volume:      "10",
			Issue:       "2",
			Pages:       "100-110",
		},
	}

	want := "A. Miller, “A paper without a DOI,” " +
		"Journal of Software Testing, vol. 10, no. 2, " +
		"pp. 100–110, 2024."

	got := NewFormatter().Format(paper)

	if got != want {
		t.Fatalf(
			"Formatter.Format() = %q, want %q",
			got,
			want,
		)
	}
}

func TestFormatterWithoutIssue(t *testing.T) {
	t.Parallel()

	repository := memory.NewRepository()

	paper, err := repository.GetByDOI(
		context.Background(),
		"10.1016/j.infsof.2020.106363",
	)
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}

	want := "S. Alyahya, " +
		"“Crowdsourced Software Testing: A Systematic Literature Review,” " +
		"Information and Software Technology, vol. 127, p. 106363, " +
		"2020, doi: 10.1016/j.infsof.2020.106363."

	got := NewFormatter().Format(paper)

	if got != want {
		t.Fatalf(
			"Formatter.Format() = %q, want %q",
			got,
			want,
		)
	}
}
