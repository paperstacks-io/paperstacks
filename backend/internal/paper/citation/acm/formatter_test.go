package acm

import (
	"context"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
)

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

	want := "Andreas Bauer et al.. " +
		"2023. " +
		"Code Review Guidelines for GUI-based Testing Artifacts. " +
		"Information and Software Technology 163, 107299. " +
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

	want := "Kai Petersen and Jan M. Gerken. " +
		"2024. " +
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

func TestFormatterAuthorsWithMoreThanThreeAuthors(t *testing.T) {
	t.Parallel()

	repository := memory.NewRepository()

	paper, err := repository.GetByDOI(
		context.Background(),
		"10.48550/ARXIV.2010.03525",
	)
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}

	want := "Paul Ralph et al.."

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

	want := "Malte Finsterwalder. " +
		"Automating Acceptance Tests for GUI Applications " +
		"in an Extreme Programming Environment. " +
		"https://doi.org/10.1000/empty"

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

	want := "Paul-Christian Bürkner. " +
		"2017. " +
		"Brms: An R Package for Bayesian Multilevel Models Using Stan. " +
		"Journal of Statistical Software 80, 1, 1–28. " +
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

	want := "Sultan Alyahya. " +
		"2020. " +
		"Crowdsourced Software Testing: A Systematic Literature Review. " +
		"Information and Software Technology 127, 106363. " +
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
