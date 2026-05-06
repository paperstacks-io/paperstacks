package web

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/common/build"
	"github.com/paperstacks.io/paperstacks/internal/paper/application"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
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
	HankoAPIURL   string
}

type papersListData struct {
	Items         []domain.Paper
	Total         int
	Page          int
	PageSize      int
	HasNext       bool
	PrevPage      int
	NextPage      int
	SearchOptions domain.SearchOptions
	Pagination    []PaginationItem
}

func handleIndex(tmpl *template.Template, navItems []navItem, hankoAPIURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		data := pageData{
			AppVersion:    build.Version,
			AppGitHash:    build.GitHash,
			AppBuildTime:  build.BuildTime,
			NavItems:      navItems,
			AppTargetID:   "app-shell",
			PageContentID: "page-content",
			HankoAPIURL:   hankoAPIURL,
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

		search := normalizeFormParam(r.FormValue("search"))
		sortByRaw := normalizeFormParam(r.FormValue("sortBy"))
		pageStr := normalizeFormParam(r.FormValue("page"))

		sortBy, _ := strings.CutPrefix(sortByRaw, "+")
		sortBy, desc := strings.CutPrefix(sortBy, "-")

		page, _ := strconv.Atoi(pageStr)
		opts := domain.SearchOptions{
			Query:  search,
			SortBy: sortBy,
			Desc:   desc,
			Page:   page,
		}
		result, err := paperService.Search(context.Background(), opts)
		if err != nil {
			logger.Error("read papers", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := papersListData{
			Items:         result.Items,
			Total:         result.Total,
			Page:          result.Page,
			PageSize:      result.PageSize,
			HasNext:       result.HasNext,
			PrevPage:      result.Page - 1,
			NextPage:      result.Page + 1,
			SearchOptions: opts,
			Pagination:    BuildPagination(result.Total, result.PageSize, result.Page),
		}
		templateName := "papers/partials/papers-list"
		if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})
}

func normalizeFormParam(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
