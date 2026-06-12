package auth

import (
	"context"
	"time"
)

type Session struct {
	Token          string
	UserID         string
	Email          string
	ExpirationTime time.Time
	IsValid        bool
}

type sessionResponse struct {
	IsValid        bool      `json:"is_valid"`
	ExpirationTime time.Time `json:"expiration_time"`
	UserID         string    `json:"user_id"`
	Claims         struct {
		AMR      []string `json:"amr"`
		Audience []string `json:"audience"`
		Email    struct {
			Address    string `json:"address"`
			IsPrimary  bool   `json:"is_primary"`
			IsVerified bool   `json:"is_verified"`
		} `json:"email"`
		Expiration time.Time `json:"expiration"`
		IssuedAt   time.Time `json:"issued_at"`
		Issuer     string    `json:"issuer"`
		SessionID  string    `json:"session_id"`
		Subject    string    `json:"subject"`
	} `json:"claims"`
}

func (r sessionResponse) Session(token string) *Session {
	return &Session{
		Token:          token,
		Email:          r.Claims.Email.Address,
		UserID:         r.UserID,
		ExpirationTime: r.ExpirationTime,
		IsValid:        r.IsValid,
	}
}

type SessionService interface {
	ResolveSession(ctx context.Context, token string) (*Session, error)
	LogoutSession(ctx context.Context, token string) error
}

type sessionContextKey struct{}

func ContextWithSession(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, session)
}

func SessionFromContext(ctx context.Context) (*Session, bool) {
	session, ok := ctx.Value(sessionContextKey{}).(*Session)
	return session, ok
}
