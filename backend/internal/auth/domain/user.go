// Package domain defines the core domain models used authentication and authorization.
package domain

// User represents an authenticated person in the system.
type User struct {
	// Email is the user's email address as provided by the authentication system.
	Email string
}
