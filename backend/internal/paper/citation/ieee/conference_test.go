package ieee

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

	want := "in 2023 ACM/IEEE International Symposium on " +
		"Empirical Software Engineering and Measurement (ESEM), " +
		"pp. 1–7"

	got := formatConferenceSource(
		shared.NewSource(paper),
	)

	if got != want {
		t.Fatalf(
			"formatConferenceSource() = %q, want %q",
			got,
			want,
		)
	}
}
