package domain

import (
	"strings"
	"time"

	userDomain "github.com/paperstacks.io/paperstacks/internal/auth/domain"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

type Stack struct {
	// UUID (Version 4) uniquely identifies a stack across the entire system.
	UUID string

	// Name is the name of the stack.
	// It must be unique within the scope of the owner.
	Name string

	// Owner identifies the user that owns the stack.
	// It represents the user's unique identifier (e.g., username).
	Owner userDomain.User

	// Papers contains all papers that belong to this stack.
	Papers []paperDomain.Paper

	// Visibility defines whether the stack is public or private.
	IsPublic bool

	// CreatedAt records when the stack was initially created.
	CreatedAt time.Time

	// UpdatedAt records the last time the stack was modified.
	UpdatedAt time.Time
}

func NewStack(name string, owner userDomain.User) *Stack {
	now := time.Now()
	name = strings.TrimSpace(name)

	return &Stack{
		UUID:      uuid.NewString(),
		Name:      name,
		Owner:     owner,
		Papers:    []paperDomain.Paper{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
