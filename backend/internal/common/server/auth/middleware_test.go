package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testSessionService struct {
	session *Session
	err     error
	token   string
}

func (s *testSessionService) ResolveSession(_ context.Context, token string) (*Session, error) {
	s.token = token
	return s.session, s.err
}

func (s *testSessionService) LogoutSession(context.Context, string) error {
	return nil
}

func TestSessionMiddlewareAttachesSessionFromBearerToken(t *testing.T) {
	service := &testSessionService{session: &Session{UserID: "user-1", IsValid: true}}

	called := false
	handler := SessionMiddleware(service)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		session, ok := SessionFromContext(r.Context())
		if !ok {
			t.Fatal("session missing from context")
		}
		if session.UserID != "user-1" {
			t.Fatalf("session user id = %q, want user-1", session.UserID)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer session-token")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if !called {
		t.Fatal("next handler was not called")
	}
	if service.token != "session-token" {
		t.Fatalf("resolved token = %q, want session-token", service.token)
	}
}

func TestSessionMiddlewareAttachesSessionFromCookie(t *testing.T) {
	service := &testSessionService{session: &Session{UserID: "user-1", IsValid: true}}

	handler := SessionMiddleware(service)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, ok := SessionFromContext(r.Context())
		if !ok {
			t.Fatal("session missing from context")
		}
		if session.UserID != "user-1" {
			t.Fatalf("session user id = %q, want user-1", session.UserID)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "hanko", Value: "cookie-token"})
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if service.token != "cookie-token" {
		t.Fatalf("resolved token = %q, want cookie-token", service.token)
	}
}

func TestSessionMiddlewareContinuesWithoutSession(t *testing.T) {
	service := &testSessionService{err: errors.New("invalid token")}

	handler := SessionMiddleware(service)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := SessionFromContext(r.Context()); ok {
			t.Fatal("session should not be attached")
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	handler.ServeHTTP(httptest.NewRecorder(), req)
}

func TestRequireAuthAPIMiddlewareRejectsMissingSession(t *testing.T) {
	handler := RequireAuthAPIMiddleware()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuthAPIMiddlewareAllowsValidSession(t *testing.T) {
	called := false
	handler := RequireAuthAPIMiddleware()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ContextWithSession(req.Context(), &Session{IsValid: true}))
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if !called {
		t.Fatal("next handler was not called")
	}
}
