package server

import (
	"fmt"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
)

func handleRoot() http.Handler {
	type Link struct {
		Href string `json:"href"`
	}

	type RootResponse struct {
		Message string          `json:"message"`
		Links   map[string]Link `json:"_links"`
	}
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			if forwardedProto := r.Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
				scheme = forwardedProto
			}

			host := r.Host
			if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
				host = forwardedHost
			}

			baseURL := fmt.Sprintf("%s://%s", scheme, host)

			resp := RootResponse{
				Message: "Welcome to paperstacks.io",
				Links: map[string]Link{
					"self": {
						Href: baseURL + "/",
					},
					"web-frontend": {
						Href: baseURL + "/web",
					},
					"api": {
						Href: baseURL + "/api",
					},
					"healthz": {
						Href: baseURL + "/healthz",
					},
				},
			}

			if err := server.Encode(w, r, http.StatusOK, resp); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}
