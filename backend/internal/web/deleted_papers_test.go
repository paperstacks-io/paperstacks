package web

import (
	"strings"
	"testing"

	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	stackDomain "github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

func TestStackDetailShowsDeletedPaperNotice(t *testing.T) {
	data := struct {
		Stack stackDomain.Stack
		SelectedPaper paperDomain.Paper
		CreatedAt string
		UpdatedAt string
	}{Stack: stackDomain.Stack{
		UUID: "stack",
		Name: "Research",
		DeletedPapers: []stackDomain.DeletedPaper{{UUID: "removed", Title: "Old <paper>"}},
	}}
	var output strings.Builder
	if err := testWebTemplate(t).ExecuteTemplate(&output, "stacks/detail", data); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Old &lt;paper&gt;") || !strings.Contains(output.String(), "no longer available") {
		t.Fatal("deleted paper notice missing or title not escaped")
	}
	if strings.Contains(output.String(), "/papers/removed") {
		t.Fatal("deleted paper must not have an active paper link")
	}
}
