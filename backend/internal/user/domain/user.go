// Package domain defines the core domain models used authentication and authorization.
package domain

import (
	"net/mail"
	"strings"
)

// User represents an authenticated person in the system.
type User struct {
	// ExternalID is the user id given by the auth provider
	ExternalID string
	// Email is the user's email address as provided by the authentication system.
	Email string
}

func (u User) Normalize() User {
	u.ExternalID = strings.TrimSpace(u.ExternalID)
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))

	return u
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

	return nil
}
