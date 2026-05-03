package domain

import (
	"strings"
	"time"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

type Stack struct {
	// UUID (Version 4) uniquely identifies a stack across the entire system.
	UUID string

	// Owner identifies the user that owns the stack.
	// It represents the user's unique identifier (e.g., username).
	Owner string

	// Name is the name of the stack.
	// It must be unique within the scope of the owner.
	Name string

	// Description provides additional context about the stack,
	// such as its purpose, contents, or usage notes.
	Description string

	// Tags contains a list of keywords used for organizing and filtering stacks.
	Tags []string

	Papers []domain.Paper

	// CreatedAt records when the stack was initially created.
	CreatedAt time.Time

	// UpdatedAt records the last time the stack was modified.
	UpdatedAt time.Time
}

func (s Stack) Normalize() Stack {
	s.UUID = strings.TrimSpace(s.UUID)
	s.Owner = strings.TrimSpace(s.Owner)
	s.Name = strings.TrimSpace(s.Name)
	s.Description = strings.TrimSpace(s.Description)

	for i, tag := range s.Tags {
		s.Tags[i] = strings.TrimSpace(tag)
	}
	return s
}
