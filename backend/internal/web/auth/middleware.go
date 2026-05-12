package auth

import (
	"net/http"
	"net/url"
)

func SessionMiddleware(service SessionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if service == nil {
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie("hanko")
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			session, err := service.ResolveSession(r.Context(), cookie.Value)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if session == nil || !session.IsValid {
				next.ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r.WithContext(ContextWithSession(r.Context(), session)))
		})
	}
}

func RequireAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, ok := SessionFromContext(r.Context())
			if !ok || session == nil || !session.IsValid {
				next := url.QueryEscape("/app" + r.URL.RequestURI())
				http.Redirect(w, r, "/app/auth?next="+next, http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
