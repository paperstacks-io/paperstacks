// Package domain defines the core domain models used authentication and authorization.
package domain

import (
	"net/mail"
	"strings"
	"time"
)

// User represents an authenticated person in the system.
type User struct {
	// ExternalID is the user id given by the auth provider
	ExternalID string

	// Email is the user's email address as provided by the authentication system.
	Email string

	// ORCID is a unique, persistent identifier free of charge to researchers
	ORCID string

	// CreatedAt records when the user was initially created.
	CreatedAt time.Time

	// UpdatedAt records the last time the user was modified.
	UpdatedAt time.Time
}

func NewUser(externalID, email string) User {
	now := time.Now()
	return User{
		ExternalID: strings.TrimSpace(externalID),
		Email:      strings.ToLower(strings.TrimSpace(email)),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func (u User) Validate() error {
	if strings.TrimSpace(u.ExternalID) == "" {
		return ErrInvalidUser
	}

	email := strings.TrimSpace(u.Email)
	if email == "" {
		return ErrInvalidUser
	}

	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return ErrInvalidUser
	}

	orcid := strings.TrimSpace(u.ORCID)
	if orcid != "" && len(orcid) != len("0000-0000-0000-0000") {
		return ErrInvalidUser
	}

	return nil
}
