package web

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/common/build"
	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	stackApp "github.com/paperstacks.io/paperstacks/internal/stack/application"
	stackDomain "github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

type pageData struct {
	Title       string
	AppVersion  string
	AppGitHash  string
	PageName    string
	HankoAPIURL string
	Session     commonauth.Session
}

func newPageData(r *http.Request, hankoAPIURL string) pageData {
	session, ok := commonauth.SessionFromContext(r.Context())
	if !ok || session == nil {
		session = &commonauth.Session{}
	}

	return pageData{
		AppVersion:  build.Version,
		AppGitHash:  build.GitHash,
		PageName:    pageNameFromPath(r.URL.Path),
		HankoAPIURL: hankoAPIURL,
		Session:     *session,
	}
}

func handlePage(tmpl *template.Template, hankoAPIURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		data := newPageData(r, hankoAPIURL)

		renderTemplate(w, r, tmpl, data)
	})
}

func handlePartialWithoutData(tmpl *template.Template, templateName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if err := tmpl.ExecuteTemplate(w, templateName, nil); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func handleSidebarStacks(
	logger *slog.Logger,
	tmpl *template.Template,
	stackService *stackApp.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		session, ok := commonauth.SessionFromContext(r.Context())
		if !ok {
			session = &commonauth.Session{}
		}

		opts := stackDomain.SearchOptions{
			Query:    "",
			SortBy:   "name",
			Page:     1,
			PageSize: 50,
		}
		result, err := stackService.SearchByOwner(r.Context(), session.UserID, opts)
		if err != nil {
			logger.Error("query stacks", "error", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		currentPath := currentHTMXPath(r)

		data := struct {
			stackDomain.SearchResult
			PageName string
		}{
			SearchResult: result,
			PageName:     pageNameFromPath(currentPath),
		}

		templateName := "shared/partials/sidebar/stacks-list"
		if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func searchOptionsFromRequest(r *http.Request) stackDomain.SearchOptions {
	search := normalizeFormParam(r.FormValue("search"))
	sortByRaw := normalizeFormParam(r.FormValue("sortBy"))
	pageStr := normalizeFormParam(r.FormValue("page"))

	sortBy, _ := strings.CutPrefix(sortByRaw, "+")
	sortBy, desc := strings.CutPrefix(sortBy, "-")

	page, _ := strconv.Atoi(pageStr)
	return stackDomain.SearchOptions{
		Query:  search,
		SortBy: sortBy,
		Desc:   desc,
		Page:   page,
	}
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

		if isHTMX(r) {
			hxRedirect(w, "/app/", http.StatusOK)
			return
		}

		http.Redirect(w, r, "/app/", http.StatusSeeOther)
	})
}
