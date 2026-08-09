package apa

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

	want := "A paper without authors. (2024). " +
		"Journal of Software Testing, 10(2), 100–110. " +
		"https://doi.org/10.1000/example"

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

	want := "Bauer, A., Coppola, R., Alégroth, E., & Gorschek, T. " +
		"(2023). " +
		"Code Review Guidelines for GUI-based Testing Artifacts. " +
		"Information and Software Technology, 163, 107299. " +
		"https://doi.org/10.1016/j.infsof.2023.107299"

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

	want := "Petersen, K., & Gerken, J. M. " +
		"(2024). " +
		"On the Road to Interactive LLM-based Systematic Mapping Studies. " +
		"Information and Software Technology, 107611. " +
		"https://doi.org/10.1016/j.infsof.2024.107611"

	got := NewFormatter().Format(paper)

	if got != want {
		t.Fatalf(
			"Formatter.Format() = %q, want %q",
			got,
			want,
		)
	}
}

func TestFormatterWithMoreThanTwentyAuthors(t *testing.T) {
	t.Parallel()

	repository := memory.NewRepository()

	paper, err := repository.GetByDOI(
		context.Background(),
		"10.48550/ARXIV.2010.03525",
	)
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}

	want := "Ralph, P., bin Ali, N., Baltes, S., Bianculli, D., " +
		"Diaz, J., Dittrich, Y., Ernst, N., Felderer, M., Feldt, R., " +
		"Filieri, A., de França, B. B. N., Furia, C. A., Gay, G., " +
		"Gold, N., Graziotin, D., He, P., Hoda, R., Juristo, N., " +
		"Kitchenham, B., . . . Vegas, S."

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

	want := "Finsterwalder, M. (n.d.). " +
		"Automating Acceptance Tests for GUI Applications " +
		"in an Extreme Programming Environment. " +
		"4. https://doi.org/10.1000/empty"

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

	want := "Bürkner, P.-C. (2017). " +
		"Brms: An R Package for Bayesian Multilevel Models Using Stan. " +
		"Journal of Statistical Software, 80(1), 1–28. " +
		"https://doi.org/10.18637/jss.v080.i01"

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

	want := "Miller, A. (2024). A paper without a DOI. " +
		"Journal of Software Testing, 10(2), 100–110."

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

	want := "Alyahya, S. (2020). " +
		"Crowdsourced Software Testing: A Systematic Literature Review. " +
		"Information and Software Technology, 127, 106363. " +
		"https://doi.org/10.1016/j.infsof.2020.106363"

	got := NewFormatter().Format(paper)

	if got != want {
		t.Fatalf(
			"Formatter.Format() = %q, want %q",
			got,
			want,
		)
	}
}
