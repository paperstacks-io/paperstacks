package web

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"strings"

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
	HasSearch     bool
	Papers 	      any
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

func handlePapersAll(tmpl *template.Template, paperService *application.PaperService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

 	data, _ := paperService.List(context.Background())
		templateName := "papers/partials/papers-table"
		if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})
}

func handlePapersSearch(
	logger *slog.Logger,
	pageTmpl *template.Template,
	tmpl *template.Template,
	paperService *application.PaperService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		title := normalizeQueryParam(r.URL.Query().Get("title"))
		keyword := normalizeQueryParam(r.URL.Query().Get("keyword"))
		sortBy := normalizeQueryParam(r.URL.Query().Get("sortBy"))

		sortBy, desc := strings.CutPrefix(sortBy, "-")

		isAllowedSortKey := sortBy == "title" || sortBy == "year"
		if sortBy != "" && !isAllowedSortKey {
			http.Error(w, "invalid 'sortBy': allowed values are 'title', 'year'", http.StatusBadRequest)
			return
		}

		hasSearch := title != "" || keyword != "" || sortBy != ""

		data := pageData{
			Title:         "Search",
			AppVersion:    "v0.1.0",
			NavItems:      navItems("/app/search"),
			AppTargetID:   "app-shell",
			PageContentID: "page-content",
			SearchTitle:   title,
			HasSearch:     hasSearch,
			Papers:        nil,
		}
		
		isHX := r.Header.Get("HX-Request") == "true"
		hxTarget := r.Header.Get("HX-Target")
		if isHX && hxTarget == "results" {
			if !data.HasSearch {
				w.WriteHeader(http.StatusOK)
				return
			}

			papers, err := paperService.Search(context.Background(), title, keyword, sortBy, desc)
			if err != nil {
				logger.Error("read papers", "error", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := tmpl.ExecuteTemplate(w, "papers/partials/papers-table", papers); err != nil {
				logger.Error("render search results", "error", err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			return
		}

		if hasSearch  {
			papers, err := paperService.Search(context.Background(), title, keyword, sortBy, desc)
			if err != nil {
				logger.Error("read papers", "error", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			data.Papers = papers
		}

		if err := pageTmpl.ExecuteTemplate(w, "base", data); err != nil {
			logger.Error("render search page", "error", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func normalizeQueryParam(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
