package web

import (
	"context"
	"html/template"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/paper/application"
	"github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
)

func handleIndex(tmpl *template.Template, navItems []navItem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		data := pageData{
			NavItems:      navItems,
			AppTargetID:   "app-shell",
			PageContentID: "page-content",
		}

		templateName := "layout"
		if r.Header.Get("HX-Request") == "true" {
			templateName = "app"
		}

		if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})
}

func handlePapersAll(tmpl *template.Template) http.Handler {
	service := application.NewPaperService(memory.NewRepository())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		data, _ := service.List(context.Background())
		templateName := "list-papers"
		if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})
}
