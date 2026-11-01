package http

import (
	"encoding/json"
	"testing"

	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

func TestStackResponsePreservesDeletedPaperReferences(t *testing.T) {
	stack := domain.Stack{
		UUID: "stack",
		Papers: []paperDomain.Paper{{UUID: "active", Title: "Available paper"}},
		DeletedPapers: []domain.DeletedPaper{{UUID: "removed", Title: "Deleted paper"}},
	}
	encoded, err := json.Marshal(NewStackResponse(stack))
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		Papers []struct {
			UUID string `json:"uuid"`
		} `json:"papers"`
		DeletedPapers []struct {
			UUID string `json:"uuid"`
			Title string `json:"title"`
		} `json:"deleted_papers"`
	}
	if err := json.Unmarshal(encoded, &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Papers) != 1 || response.Papers[0].UUID != "active" {
		t.Fatalf("unexpected active papers: %+v", response.Papers)
	}
	if len(response.DeletedPapers) != 1 || response.DeletedPapers[0].UUID != "removed" || response.DeletedPapers[0].Title != "Deleted paper" {
		t.Fatalf("deleted reference missing from JSON: %+v", response.DeletedPapers)
	}
}
