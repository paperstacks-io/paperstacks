package acm

import (
	"context"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/paper/citation/shared"
	"github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
)

func TestFormatConferenceSource(t *testing.T) {
	t.Parallel()

	repository := memory.NewRepository()

	paper, err := repository.GetByDOI(
		context.Background(),
		"10.1109/ESEM56168.2023.10304792",
	)
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}

	completeSource := shared.NewSource(paper)

	withoutPublisher := completeSource
	withoutPublisher.Publisher = ""

	withoutPages := completeSource
	withoutPages.Pages = ""

	withoutPublishedIn := completeSource
	withoutPublishedIn.PublishedIn = ""

	tests := []struct {
		name   string
		source shared.Source
		want   string
	}{
		{
			name:   "proceedings publisher and pages",
			source: completeSource,
			want: "In 2023 ACM/IEEE International Symposium on " +
				"Empirical Software Engineering and Measurement (ESEM). " +
				"IEEE, 1–7.",
		},
		{
			name:   "without publisher",
			source: withoutPublisher,
			want: "In 2023 ACM/IEEE International Symposium on " +
				"Empirical Software Engineering and Measurement (ESEM). " +
				"1–7.",
		},
		{
			name:   "without pages",
			source: withoutPages,
			want: "In 2023 ACM/IEEE International Symposium on " +
				"Empirical Software Engineering and Measurement (ESEM). " +
				"IEEE.",
		},
		{
			name:   "without published in",
			source: withoutPublishedIn,
			want:   "",
		},
		{
			name:   "empty source",
			source: shared.Source{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatConferenceSource(tt.source)

			if got != tt.want {
				t.Fatalf(
					"formatConferenceSource() = %q, want %q",
					got,
					tt.want,
				)
			}
		})
	}
}
