package auth

import (
	"net/http"
	"net/url"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/auth"
)

func RequireAuthWebMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, ok := commonauth.SessionFromContext(r.Context())
			if !ok || session == nil || !session.IsValid {
				next := url.QueryEscape("/app" + r.URL.RequestURI())
				http.Redirect(w, r, "/app/auth?next="+next, http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
