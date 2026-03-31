package web

import (
	"context"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/common/build"
	"github.com/paperstacks.io/paperstacks/internal/paper/application"
)

type pageData struct {
	Title         string
	AppVersion    string
	AppGitHash    string
	AppBuildTime  string
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
			AppVersion:    build.Version,
			AppGitHash:    build.GitHash,
			AppBuildTime:  build.BuildTime,
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

		search := normalizeQueryParam(r.FormValue("search"))
		sortRaw := normalizeQueryParam(r.FormValue("sortBy"))

		sortBy, desc := sortRaw, false
		if s, ok := strings.CutPrefix(sortRaw, "-"); ok {
			sortBy, desc = s, true
		} else if s, ok := strings.CutPrefix(sortRaw, "+"); ok {
			sortBy = s
		}

		data, err := paperService.Search(context.Background(), search, search, sortBy, desc)
		if err != nil {
			logger.Error("read papers", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for i := range data {
			log.Println("paper", data[i].PublicationYear)
		}

		templateName := "papers/partials/papers-table"
		if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})
}

func normalizeQueryParam(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
