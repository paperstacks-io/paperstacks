package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
)

func TestServiceCreateNormalizesAndValidatesPaper(t *testing.T) {
	t.Parallel()

	service := NewPaperService(memory.NewRepository())

	err := service.Create(context.Background(), domain.Paper{
		DOI:   " 10.1000/example ",
		Title: " Example Paper ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := service.GetByDOI(context.Background(), "10.1000/example")
	if err != nil {
		t.Fatalf("GetByDOI() error = %v", err)
	}

	_, err = uuid.Parse(got.UUID)
	if got.UUID == "" || err != nil {
		t.Fatalf("expected valid uuid, got %q", got.UUID)
	}

	if got.DOI != "10.1000/example" {
		t.Fatalf("stored DOI = %q, want %q", got.DOI, "10.1000/example")
	}

	if got.Title != "Example Paper" {
		t.Fatalf("stored title = %q, want %q", got.Title, "Example Paper")
	}
}

func TestServiceUpdateRejectsMismatchedUUID(t *testing.T) {
	t.Parallel()

	service := NewPaperService(memory.NewRepository())

	err := service.Update(context.Background(), "a4b065f1-1b88-4f50-a7fe-1177f3489fcf", domain.Paper{
		UUID:  "a4b065f1-other",
		DOI:   "10.9999/other",
		Title: "Updated",
	})
	if err != domain.ErrUUIDMismatch {
		t.Fatalf("Update() error = %v, want %v", err, domain.ErrUUIDMismatch)
	}
}
