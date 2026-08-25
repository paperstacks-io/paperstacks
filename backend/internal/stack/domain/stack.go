package domain

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
	"uuid"

	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

const (
	minStackNameRunes = 1
	maxStackNameRunes = 80
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

	// IsPublic defines whether the stack is public or private.
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
		UUID:      uuid.New().String(),
		Name:      name,
		Owner:     owner,
		Papers:    []paperDomain.Paper{},
		IsPublic:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (s Stack) ContainsPaperWithDOI(doi string) (paperDomain.Paper, bool) {
	for _, paper := range s.Papers {
		if paper.DOI == doi {
			return paper, true
		}
	}
	return paperDomain.Paper{}, false
}

func (s Stack) Validate() error {
	_, err := uuid.Parse(s.UUID)
	if err != nil {
		return ErrInvalidStack
	}

	if err := validateStackName(s.Name); err != nil {
		return err
	}

	return nil
}

func validateStackName(name string) error {
	length := utf8.RuneCountInString(name)
	if length < minStackNameRunes || length > maxStackNameRunes {
		return ErrInvalidName
	}

	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			continue
		}

		switch r {
		case ' ', '-', '_', '.', ',', ':', '\'', '&', '/', '(', ')', '+', '#':
			continue
		default:
			return ErrInvalidName
		}
	}

	return nil
}
