// Package domain defines the core domain models used authentication and authorization.
package domain

import (
	"errors"
	"net/mail"
)

var (
	ErrInvalidUser = errors.New("invalid User")
)

// User represents an authenticated person in the system.
type User struct {
	// ExternalID is the user id given by the auth provider
	ExternalID string
	// Email is the user's email address as provided by the authentication system.
	Email string
}

func (u *User) Validate() error {
	if u.ExternalID == "" {
		return ErrInvalidUser
	}

	_, err := mail.ParseAddress(u.Email)
	if err != nil {
		return ErrInvalidUser
	}

	return nil
}
