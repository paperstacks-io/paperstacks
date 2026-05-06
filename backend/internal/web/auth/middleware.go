package auth

import "net/http"

func AuthMiddleware(validator SessionValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("hanko")
			if err != nil {
				http.Redirect(w, r, "/app/auth", http.StatusSeeOther)
				return
			}

			// Validate the session token
			isValid, err := validator.ValidateSession(cookie.Value)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if !isValid {
				http.Redirect(w, r, "/app/auth", http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
