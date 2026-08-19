package domain

import (
	"strings"
	"testing"

	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

func TestStackContainsPaperWithDOI(t *testing.T) {
	t.Parallel()

	stack := Stack{Papers: []paperDomain.Paper{{DOI: "10.1000/example"}}}
	if !stack.ContainsPaperWithDOI("10.1000/example") {
		t.Error("ContainsPaperWithDOI() = false, want true")
	}
	if stack.ContainsPaperWithDOI("10.1000/missing") {
		t.Error("ContainsPaperWithDOI() = true, want false")
	}
}

func TestStackValidateAcceptsValidNames(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"Machine Learning",
		"NLP/IR & Retrieval (2026)",
		"C++ #notes: Smith, et al.",
		strings.Repeat("a", maxStackNameRunes),
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stack := NewStack(name, userDomain.User{ExternalID: "user"})
			if err := stack.Validate(); err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
		})
	}
}

func TestStackValidateRejectsInvalidNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
	}{
		{name: "empty", want: ""},
		{name: "too long", want: strings.Repeat("a", maxStackNameRunes+1)},
		{name: "angle bracket", want: "bad<name"},
		{name: "emoji", want: "research 🚀"},
		{name: "newline", want: "research\nnotes"},
		{name: "newline", want: "func main() {}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stack := NewStack("valid", userDomain.User{ExternalID: "user"})
			stack.Name = tt.want

			if err := stack.Validate(); err != ErrInvalidName {
				t.Fatalf("Validate() error = %v, want %v", err, ErrInvalidName)
			}
		})
	}
}
