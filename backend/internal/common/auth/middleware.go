package auth

import (
	"net/http"
	"strings"
)

func SessionMiddleware(service SessionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := sessionToken(r)

			session, err := service.ResolveSession(r.Context(), token)
			if err != nil || session == nil || !session.IsValid {
				next.ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r.WithContext(ContextWithSession(r.Context(), session)))
		})
	}
}

func sessionToken(r *http.Request) string {
	cookie, err := r.Cookie("hanko")
	if err == nil {
		return cookie.Value
	}

	if token, ok := bearerToken(r.Header.Get("Authorization")); ok {
		return token
	}

	return ""
}

func RequireAuthAPIMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, ok := SessionFromContext(r.Context())
			if !ok || session == nil || !session.IsValid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func bearerToken(header string) (string, bool) {
	token, ok := strings.CutPrefix(strings.TrimSpace(header), "Bearer ")
	if !ok {
		return "", false
	}

	token = strings.TrimSpace(token)
	return token, token != ""
}
