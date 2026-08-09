package apa

import (
	"context"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/paper/citation/shared"
	"github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
)

func TestFormatJournalSource(t *testing.T) {
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
		"10.1016/j.infsof.2024.107611",
	)
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}

	tests := []struct {
		name   string
		source shared.Source
		want   string
	}{
		{
			name:   "volume and issue",
			source: shared.NewSource(withVolumeAndIssue),
			want:   "Journal of Statistical Software, 80(1), 1–28.",
		},
		{
			name:   "volume without issue",
			source: shared.NewSource(withVolumeWithoutIssue),
			want:   "Information and Software Technology, 127, 106363.",
		},
		{
			name: "issue without volume",
			source: shared.Source{
				PublishedIn: "Journal of Software Testing",
				Issue:       "2",
				Pages:       "100-110",
			},
			want: "Journal of Software Testing, (2), 100–110.",
		},
		{
			name:   "without volume and issue",
			source: shared.NewSource(withoutVolumeAndIssue),
			want:   "Information and Software Technology, 107611.",
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

			got := formatJournalSource(tt.source)

			if got != tt.want {
				t.Fatalf(
					"formatJournalSource() = %q, want %q",
					got,
					tt.want,
				)
			}
		})
	}
}
