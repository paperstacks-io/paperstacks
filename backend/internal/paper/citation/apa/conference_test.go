package apa

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
	withPublisherAndPages := shared.NewSource(paper)

	tests := []struct {
		name   string
		source shared.Source
		want   string
	}{
		{
			name:   "publisher and pages",
			source: withPublisherAndPages,
			want: "In 2023 ACM/IEEE International Symposium on " +
				"Empirical Software Engineering and Measurement (ESEM) " +
				"(pp. 1–7). IEEE.",
		},
		{
			name: "without publisher",
			source: shared.Source{
				PublishedIn: "Proceedings of the Software Testing Conference",
				Pages:       "20-30",
			},
			want: "In Proceedings of the Software Testing Conference " +
				"(pp. 20–30).",
		},
		{
			name: "without pages",
			source: shared.Source{
				PublishedIn: "Proceedings of the Software Testing Conference",
				Publisher:   "ACM",
			},
			want: "In Proceedings of the Software Testing Conference. ACM.",
		},
		{
			name: "without proceedings",
			source: shared.Source{
				Publisher: "IEEE",
				Pages:     "1-7",
			},
			want: "IEEE.",
		},
		{
			name:   "without source",
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
