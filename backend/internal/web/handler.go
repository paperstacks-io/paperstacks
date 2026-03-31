package web

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/paper/application"
)

type pageData struct {
	Title         string
	AppVersion    string
	PageName      string
	NavItems      []navItem
	AppTargetID   string
	PageContentID string
	SearchTitle   string
}

func handleIndex(tmpl *template.Template, navItems []navItem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		data := pageData{
			AppVersion:    "v0.1.0",
			NavItems:      navItems,
			AppTargetID:   "app-shell",
			PageContentID: "page-content",
			SearchTitle:   "",
		}

		templateName := "base"
		if r.Header.Get("HX-Request") == "true" {
			templateName = "app"
		}

		if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})
}

func handlePapersSearch(
	logger *slog.Logger,
	tmpl *template.Template,
	paperService *application.PaperService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		search := r.FormValue("search")

		data, _ := paperService.Search(context.Background(), search, search, "title", false)
		templateName := "papers/partials/papers-table"
		if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})
}
