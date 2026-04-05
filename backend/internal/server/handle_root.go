package server

import (
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
			resp := RootResponse{
				Message: "Welcome to paperstacks.io",
				Links: map[string]Link{
					"self": {
						Href: "/",
					},
					"web-frontend": {
						Href: "/web",
					},
					"api": {
						Href: "/api",
					},
				},
			}

			if err := server.Encode(w, r, http.StatusOK, resp); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}
