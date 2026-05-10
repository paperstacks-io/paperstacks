package auth

import "net/http"

func AuthMiddleware(service SessionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("hanko")
			if err != nil {
				http.Redirect(w, r, "/app/auth", http.StatusSeeOther)
				return
			}

			session, err := service.ResolveSession(r.Context(), cookie.Value)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if session == nil || !session.IsValid {
				http.Redirect(w, r, "/app/auth", http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r.WithContext(ContextWithSession(r.Context(), session)))
		})
	}
}
