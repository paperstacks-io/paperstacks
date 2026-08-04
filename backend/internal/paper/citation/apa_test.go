package application

import (
	"context"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
)

func TestPaperAPACitationWithoutAuthors(t *testing.T) {
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

	want := "A paper without authors. (2024). Journal of Software Testing, 10(2), 100–110. https://doi.org/10.1000/example"

	got := FormatAPA(paper)

	if got != want {
		t.Fatalf("APACitation() = %q, want %q", got, want)
	}
}

func TestFormatAPAWithMultipleAuthors(t *testing.T) {
	t.Parallel()

	repo := memory.NewRepository()

	paper, err := repo.GetByDOI(
		context.Background(),
		"10.1016/j.infsof.2023.107299",
	)
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}

	want := "Bauer, A., Coppola, R., Alégroth, E., & Gorschek, T. (2023). Code Review Guidelines for GUI-based Testing Artifacts. Information and Software Technology, 163, 107299. https://doi.org/10.1016/j.infsof.2023.107299"

	got := FormatAPA(paper)

	if got != want {
		t.Errorf("FormatAPA() = %q, want %q", got, want)
	}
}

func TestPaperAPACitationWithTwoAuthors(t *testing.T) {
	t.Parallel()

	repository := memory.NewRepository()

	paper, err := repository.GetByDOI(
		context.Background(),
		"10.1016/j.infsof.2024.107611",
	)
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}

	want := "Petersen, K., & Gerken, J. M. (2024). On the Road to Interactive LLM-based Systematic Mapping Studies. Information and Software Technology, 107611. https://doi.org/10.1016/j.infsof.2024.107611"

	got := FormatAPA(paper)

	if got != want {
		t.Fatalf("FormatAPA() = %q, want %q", got, want)
	}
}

func TestFormatAPAAuthorsWithMoreThanTwentyAuthors(t *testing.T) {
	t.Parallel()

	repository := memory.NewRepository()

	paper, err := repository.GetByDOI(
		context.Background(),
		"10.48550/ARXIV.2010.03525",
	)
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}

	want := "Ralph, P., bin Ali, N., Baltes, S., Bianculli, D., Diaz, J., Dittrich, Y., Ernst, N., Felderer, M., Feldt, R., Filieri, A., de França, B. B. N., Furia, C. A., Gay, G., Gold, N., Graziotin, D., He, P., Hoda, R., Juristo, N., Kitchenham, B., … Vegas, S."

	got := formatAPAAuthors(paper.Authors)

	if got != want {
		t.Fatalf("formatAPAAuthors() = %q, want %q", got, want)
	}
}

func TestPaperAPACitationWithoutPublicationYear(t *testing.T) {
	t.Parallel()

	repository := memory.NewRepository()

	paper, err := repository.GetByUUID(
		context.Background(),
		"5966a651-6ed5-5bbc-bfe0-c76331d373a6",
	)
	if err != nil {
		t.Fatalf("GetByUUID() error = %v", err)
	}

	want := "Finsterwalder, M. (n.d.). Automating Acceptance Tests for GUI Applications in an Extreme Programming Environment. 4. https://doi.org/10.1000/empty"

	got := FormatAPA(paper)

	if got != want {
		t.Fatalf("FormatAPA() = %q, want %q", got, want)
	}
}

func TestPaperAPACitationWithHyphenatedFirstName(t *testing.T) {
	t.Parallel()

	repository := memory.NewRepository()

	paper, err := repository.GetByDOI(
		context.Background(),
		"10.18637/jss.v080.i01",
	)
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}

	want := "Bürkner, P.-C. (2017). Brms: An R Package for Bayesian Multilevel Models Using Stan. Journal of Statistical Software, 80(1), 1–28. https://doi.org/10.18637/jss.v080.i01"

	got := FormatAPA(paper)

	if got != want {
		t.Fatalf("FormatAPA() = %q, want %q", got, want)
	}
}

func TestPaperAPACitationWithoutDOI(t *testing.T) {
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

	want := "Miller, A. (2024). A paper without a DOI. Journal of Software Testing, 10(2), 100–110."

	got := FormatAPA(paper)

	if got != want {
		t.Fatalf("APACitation() = %q, want %q", got, want)
	}
}

func TestPaperAPACitationWithoutIssue(t *testing.T) {
	t.Parallel()

	repository := memory.NewRepository()

	paper, err := repository.GetByDOI(
		context.Background(),
		"10.1016/j.infsof.2020.106363",
	)
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}

	want := "Alyahya, S. (2020). Crowdsourced Software Testing: A Systematic Literature Review. Information and Software Technology, 127, 106363. https://doi.org/10.1016/j.infsof.2020.106363"

	got := FormatAPA(paper)

	if got != want {
		t.Fatalf("FormatAPA() = %q, want %q", got, want)
	}
}

func TestFormatAPASourceWithVolumeAndIssue(t *testing.T) {
	t.Parallel()

	repository := memory.NewRepository()
	ctx := context.Background()

	withVolumeAndIssue, err := repository.GetByDOI(
		ctx,
		"10.18637/jss.v080.i01",
	)
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}

	withVolumeWithoutIssue, err := repository.GetByDOI(
		ctx,
		"10.1016/j.infsof.2020.106363",
	)
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}

	withoutVolumeAndIssue, err := repository.GetByDOI(
		ctx,
		"10.1109/ESEM56168.2023.10304792",
	)
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}
	t.Logf(
		"withoutVolumeAndIssue.Type = %q",
		withoutVolumeAndIssue.Type,
	)

	tests := []struct {
		name      string
		metadata  domain.Metadata
		want      string
		paperType string
	}{
		{
			name:      "volume and issue",
			metadata:  withVolumeAndIssue.Metadata,
			want:      "Journal of Statistical Software, 80(1), 1–28.",
			paperType: withVolumeAndIssue.Type,
		},
		{
			name:      "volume without issue",
			metadata:  withVolumeWithoutIssue.Metadata,
			want:      "Information and Software Technology, 127, 106363.",
			paperType: withVolumeWithoutIssue.Type,
		},
		{
			name: "issue without volume",
			metadata: domain.Metadata{
				PublishedIn: "Journal of Software Testing",
				Issue:       "2",
				Pages:       "100-110",
			},
			want:      "Journal of Software Testing, (2), 100–110.",
			paperType: "article",
		},
		{
			name:      "without volume and issue",
			metadata:  withoutVolumeAndIssue.Metadata,
			want:      "2023 ACM/IEEE International Symposium on Empirical Software Engineering and Measurement (ESEM), 1–7.",
			paperType: withoutVolumeAndIssue.Type,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatAPASource(
				tt.paperType,
				tt.metadata,
			)

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
