package web

import (
	"html/template"
	"net/http"
	"net/url"
	"strings"
)

type toastData struct {
	Message string
}

func renderTemplate(w http.ResponseWriter, r *http.Request, tmpl *template.Template, data any) {
	templateName := "base"
	if isHTMX(r) {
		templateName = "app"
	}

	if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func hxRedirect(w http.ResponseWriter, location string, status int) {
	w.Header().Set("HX-Redirect", location)
	w.WriteHeader(status)
}

func currentHTMXPath(r *http.Request) string {
	currentPath := r.URL.Path
	if hxCurrentURL := r.Header.Get("HX-Current-URL"); hxCurrentURL != "" {
		if u, err := url.Parse(hxCurrentURL); err == nil {
			currentPath = strings.TrimPrefix(u.Path, "/app")
		}
	}

	return currentPath
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

func pageNameFromPath(path string) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return "home"
	}

	pageName, _, _ := strings.Cut(path, "?")

	return pageName
}

func normalizeFormParam(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
