package server_

import (
	"net/http"
)

func handleExample() http.Handler {
	type example struct {
		Author string `json:"author"`
		Title  string `json:"title"`
		Year   int    `json:"year"`
	}

	hans := &example{
		Author: "Hans",
		Title:  "Hans' Adventures in Wonderland",
		Year:   2025,
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			err := encode(w, r, http.StatusOK, hans)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}
