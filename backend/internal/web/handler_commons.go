package web

import (
	"html/template"
	"net/http"
)

type toastData struct {
	Message string
}

func renderTemplate(w http.ResponseWriter, r *http.Request, tmpl *template.Template, data any) {
	templateName := "base"
	if r.Header.Get("HX-Request") == "true" {
		templateName = "app"
	}

	if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func renderSuccessToast(w http.ResponseWriter, tmpl *template.Template, message string) error {
	templateName := "shared/partials/toast/success"
	return renderToast(w, tmpl, templateName, toastData{Message: message})
}

func renderInfoToast(w http.ResponseWriter, tmpl *template.Template, message string) error {
	templateName := "shared/partials/toast/info"
	return renderToast(w, tmpl, templateName, toastData{Message: message})
}

func renderWarningToast(w http.ResponseWriter, tmpl *template.Template, message string) error {
	templateName := "shared/partials/toast/warning"
	return renderToast(w, tmpl, templateName, toastData{Message: message})
}

func renderErrorToast(w http.ResponseWriter, tmpl *template.Template, message string) error {
	templateName := "shared/partials/toast/error"
	return renderToast(w, tmpl, templateName, toastData{Message: message})
}

func renderToast(w http.ResponseWriter, tmpl *template.Template, templateName string, data toastData) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Retarget", "#toast-host")
	w.Header().Set("HX-Reswap", "innerHTML")

	return tmpl.ExecuteTemplate(w, templateName, data)
}
