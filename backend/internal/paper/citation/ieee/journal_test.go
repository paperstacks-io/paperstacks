package ieee

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
			want: "Journal of Statistical Software, vol. 80, " +
				"no. 1, pp. 1–28",
		},
		{
			name:   "volume without issue",
			source: shared.NewSource(withVolumeWithoutIssue),
			want: "Information and Software Technology, " +
				"vol. 127, p. 106363",
		},
		{
			name:   "without volume and issue",
			source: shared.NewSource(withoutVolumeAndIssue),
			want: "Information and Software Technology, " +
				"p. 107611",
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
