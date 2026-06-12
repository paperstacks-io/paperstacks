package web

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/common/build"
	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	paperApp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	stackApp "github.com/paperstacks.io/paperstacks/internal/stack/application"
	stackDomain "github.com/paperstacks.io/paperstacks/internal/stack/domain"
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
	Session       commonauth.Session
}

type papersListData struct {
	Items         []paperDomain.Paper
	Total         int
	Page          int
	PageSize      int
	HasNext       bool
	PrevPage      int
	NextPage      int
	SearchOptions paperDomain.SearchOptions
	Pagination    []PaginationItem
}

func handleIndex(tmpl *template.Template, navItems []navItem, hankoAPIURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		session, ok := commonauth.SessionFromContext(r.Context())
		if !ok {
			session = &commonauth.Session{}
		}

		data := pageData{
			AppVersion:    build.Version,
			AppGitHash:    build.GitHash,
			AppBuildTime:  build.BuildTime,
			NavItems:      navItems,
			AppTargetID:   "app-shell",
			PageContentID: "page-content",
			HankoAPIURL:   hankoAPIURL,
			Session:       *session,
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
	paperService *paperApp.PaperService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		search := normalizeFormParam(r.FormValue("search"))
		sortByRaw := normalizeFormParam(r.FormValue("sortBy"))
		pageStr := normalizeFormParam(r.FormValue("page"))

		sortBy, _ := strings.CutPrefix(sortByRaw, "+")
		sortBy, desc := strings.CutPrefix(sortBy, "-")

		page, _ := strconv.Atoi(pageStr)
		opts := paperDomain.SearchOptions{
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

func handleStacksSearch(
	logger *slog.Logger,
	tmpl *template.Template,
	stackService *stackApp.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("handleStacksSearch called", "method", r.Method, "path", r.URL.Path)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		logger.Info("authorization header", "value", r.Header.Get("Authorization"))

		result, err := stackService.ListAllPublic(r.Context())
		if err != nil {
			logger.Error("read stacks", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := struct {
			Items []stackDomain.Stack
		}{
			Items: result,
		}

		if err := tmpl.ExecuteTemplate(w, "stacks/partials/stacks-list", data); err != nil {
			logger.Error("render stacks list", "error", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func handleLogout(
	logger *slog.Logger,
	sessionService commonauth.SessionService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, ok := commonauth.SessionFromContext(r.Context())
		if !ok {
			session = &commonauth.Session{}
		}

		err := sessionService.LogoutSession(r.Context(), session.Token)
		if err != nil {
			logger.Error("error while logout", "error", err.Error())
			http.Error(w, "failed to logout session", http.StatusInternalServerError)
			return
		}

		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/app/")
			w.WriteHeader(http.StatusOK)
			return
		}

		http.Redirect(w, r, "/app/", http.StatusSeeOther)
	})
}

func normalizeFormParam(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
