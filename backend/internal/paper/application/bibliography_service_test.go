package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
)

func TestBibliographyServiceExportBibLaTeXPreservesRequestedOrder(t *testing.T) {
	t.Parallel()

	service := NewBibliographyService(memory.NewRepository())
	firstUUID := "0f324174-926b-585d-b121-3a1e3f7fee0b"
	secondUUID := "36583bb4-8cdc-554e-bcf5-f67b60d0b290"

	got, err := service.ExportBibLaTeX(context.Background(), []string{" " + firstUUID + " ", secondUUID})
	if err != nil {
		t.Fatalf("ExportBibLaTeX() error = %v", err)
	}

	first := strings.Index(string(got), "@article{10.1109/52.28121,")
	second := strings.Index(string(got), "@article{10.48550/ARXIV.1709.08439,")
	if first == -1 || second == -1 || first >= second {
		t.Fatalf("ExportBibLaTeX() did not preserve request order:\n%s", got)
	}
}

func TestBibliographyServiceExportBibLaTeXReturnsMissingPaper(t *testing.T) {
	t.Parallel()

	service := NewBibliographyService(memory.NewRepository())

	_, err := service.ExportBibLaTeX(context.Background(), []string{"a4b065f1-1b88-4f50-a7fe-1177f3489fcf"})
	if !errors.Is(err, domain.ErrPaperNotFound) {
		t.Fatalf("ExportBibLaTeX() error = %v, want %v", err, domain.ErrPaperNotFound)
	}
}
