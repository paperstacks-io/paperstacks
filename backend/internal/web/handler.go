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
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
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
	Stacks        []navStackItem
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

type stacksListData struct {
	Items         []stackDomain.Stack
	Total         int
	Page          int
	PageSize      int
	HasNext       bool
	PrevPage      int
	NextPage      int
	SearchOptions stackDomain.SearchOptions
	Pagination    []PaginationItem
}

type alertData struct {
	Message string
}

func handleIndex(
	logger *slog.Logger,
	tmpl *template.Template,
	navItems []navItem,
	hankoAPIURL string,
	stackService *stackApp.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		session, ok := commonauth.SessionFromContext(r.Context())
		if !ok {
			session = &commonauth.Session{}
		}

		navStacks := []navStackItem{}
		if session.IsValid {
			stacks, err := stackService.List(r.Context(), session.UserID)
			if err != nil {
				logger.Error("read sidebar stacks", "userId", session.UserID, "error", err.Error())
			}

			navStacks = navStackItems(stacks)
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
			Stacks:        navStacks,
		}

		templateName := "base"
		if r.Header.Get("HX-Request") == "true" {
			templateName = "app"
		}

		if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
			logger.Error("render page", "error", err.Error())
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
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		search := normalizeFormParam(r.FormValue("search"))
		sortByRaw := normalizeFormParam(r.FormValue("sortBy"))
		pageStr := normalizeFormParam(r.FormValue("page"))

		sortBy, _ := strings.CutPrefix(sortByRaw, "+")
		sortBy, desc := strings.CutPrefix(sortBy, "-")

		page, _ := strconv.Atoi(pageStr)
		opts := stackDomain.SearchOptions{
			Query:  search,
			SortBy: sortBy,
			Desc:   desc,
			Page:   page,
		}

		result, err := stackService.Search(r.Context(), opts)
		if err != nil {
			logger.Error("read stacks", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := stacksListData{
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

		templateName := "stacks/partials/stacks-list"
		if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func handleStacksCreate(
	logger *slog.Logger,
	tmpl *template.Template,
	stackService *stackApp.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		const (
			createStackErrorTarget   = "#create_stack_error"
			createStackSuccessTarget = "#create_stack_success"
		)

		render := func(status int, target string, templateName string, data alertData) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("HX-Retarget", target)
			w.Header().Set("HX-Reswap", "innerHTML")
			w.WriteHeader(status)

			if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
				logger.Error("render stack create", "error", err.Error())
			}
		}

		renderError := func(status int, message string) {
			render(status, createStackErrorTarget, "stacks/partials/alert-error", alertData{
				Message: message,
			})
		}

		renderSuccess := func(message string) {
			render(http.StatusCreated, createStackSuccessTarget, "stacks/partials/toast-success", alertData{
				Message: message,
			})
		}

		session, ok := commonauth.SessionFromContext(ctx)
		if !ok || session == nil || !session.IsValid {
			renderError(http.StatusUnauthorized, "Unauthorized. Please log in to create a stack.")
			return
		}

		name := r.FormValue("name")
		isPublic := r.FormValue("is_public") == "on"

		user := userDomain.NewUser(session.UserID, session.Email)

		stack := stackDomain.NewStack(name, user)
		stack.IsPublic = isPublic

		if err := stackService.Create(ctx, *stack); err != nil {
			logger.Error("create stack", "error", err.Error())

			if err == stackDomain.ErrStackAlreadyExists {
				renderError(http.StatusConflict, "A stack with the name '"+name+"' already exists.")
				return
			}

			renderError(http.StatusUnprocessableEntity, "Failed to create stack. Please try again.")
			return
		}

		renderSuccess("Stack '" + name + "' created successfully.")
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
