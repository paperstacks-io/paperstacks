package web

import (
	"html/template"
	"net/http"
)

func handlePage(tmpl *template.Template, title string, pageName string, navItems []navItem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		data := pageData{
			Title:           title,
			PageName:        pageName,
			NavItems:        navItems,
			ContentTargetID: "page-content",
		}

		templateName := "layout"
		if r.Header.Get("HX-Request") == "true" {
			templateName = "page-content"
		}

		if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})
}
