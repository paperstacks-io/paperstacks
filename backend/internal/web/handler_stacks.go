package web

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/common/build"
	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	paperApp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	stackApp "github.com/paperstacks.io/paperstacks/internal/stack/application"
	stackDomain "github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

const invalidStackNameMessage = "Stack name must be 1-80 characters and contain only letters, numbers, spaces, or - _ . , : ' & / ( ) + #."

const stacksPageURL = "/app/stacks/page"

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

func NewStacksListData(result stackDomain.SearchResult, opts stackDomain.SearchOptions) stacksListData {
	return stacksListData{
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
}

type stacksPageData struct {
	pageData
	StacksCountTotal  int
	StacksCountPublic int
}

func handleStacksDetailPage(
	logger *slog.Logger,
	tmpl *template.Template,
	hankoAPIURL string,
	stackService *stackApp.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		ctx := r.Context()
		id := r.PathValue("uuid")

		session, ok := commonauth.SessionFromContext(ctx)
		if !ok {
			session = &commonauth.Session{}
		}

		stack, err := stackService.GetByUUID(ctx, id)
		if err != nil {
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", stacksPageURL)
				w.WriteHeader(http.StatusOK)
				return
			}

			http.Redirect(w, r, stacksPageURL, http.StatusSeeOther)

			return
		}

		selectedPaper := paperDomain.Paper{}
		if len(stack.Papers) > 0 {
			selectedPaper = stack.Papers[0]
		}

		data := struct {
			pageData
			Stack         stackDomain.Stack
			SelectedPaper paperDomain.Paper
			CreatedAt     string
			UpdatedAt     string
		}{
			pageData: pageData{
				AppVersion:  build.Version,
				AppGitHash:  build.GitHash,
				PageName:    pageNameFromPath(r.URL.Path),
				HankoAPIURL: hankoAPIURL,
				Session:     *session,
			},
			Stack:         stack,
			SelectedPaper: selectedPaper,
			CreatedAt:     stack.CreatedAt.Format("2006-01-02 15:04"),
			UpdatedAt:     stack.UpdatedAt.Format("2006-01-02 15:04"),
		}

		renderTemplate(w, r, tmpl, data)
	})
}

func handleStackPaperInfo(
	logger *slog.Logger,
	tmpl *template.Template,
	paperService *paperApp.PaperService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		ctx := r.Context()
		paperUUID := r.PathValue("paperUUID")

		if paperUUID == "" {
			http.Error(w, "missing paper uuid", http.StatusBadRequest)
			return
		}

		paper, err := paperService.GetByUUID(ctx, paperUUID)
		if err != nil {
			if err == paperDomain.ErrPaperNotFound {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			logger.Error("get paper by UUID", "error", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "stacks/partials/paper-info", paper); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})
}

func handleStacksPage(
	logger *slog.Logger,
	tmpl *template.Template,
	hankoAPIURL string,
	stackService *stackApp.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		session, ok := commonauth.SessionFromContext(r.Context())
		if !ok {
			session = &commonauth.Session{}
		}

		counter, err := stackService.CountPublic(r.Context())
		if err != nil {
			logger.Error("count public stacks", "error", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		data := stacksPageData{
			pageData: pageData{
				AppVersion:  build.Version,
				AppGitHash:  build.GitHash,
				PageName:    pageNameFromPath(r.URL.Path),
				HankoAPIURL: hankoAPIURL,
				Session:     *session,
			},
			StacksCountTotal:  0,
			StacksCountPublic: counter,
		}

		renderTemplate(w, r, tmpl, data)
	})
}

func handleStacksSearchByOwner(
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

		opts := searchOptionsFromRequest(r)
		result, err := stackService.SearchByOwner(r.Context(), session.UserID, opts)
		if err != nil {
			logger.Error("read stacks", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := NewStacksListData(result, opts)
		templateName := "stacks/partials/stacks-list"
		if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func handleStacksStatsByOwner(
	logger *slog.Logger,
	tmpl *template.Template,
	stackService *stackApp.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		ctx := r.Context()

		session, ok := commonauth.SessionFromContext(ctx)
		if !ok {
			session = &commonauth.Session{}
		}
		stats, err := stackService.GetStatsByOwner(ctx, session.UserID)
		if err != nil {
			logger.Error("read stacks", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templateName := "stacks/partials/stats-my"
		if err := tmpl.ExecuteTemplate(w, templateName, stats); err != nil {
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

		if err := stackService.CreateByName(ctx, name, session.UserID); err != nil {
			logger.Error("create stack", "error", err.Error())

			switch {
			case errors.Is(err, stackDomain.ErrStackAlreadyExists):
				renderError(http.StatusConflict, "A stack with the name '"+name+"' already exists.")
			case errors.Is(err, stackDomain.ErrInvalidName):
				renderError(http.StatusUnprocessableEntity, invalidStackNameMessage)
			default:
				renderError(http.StatusUnprocessableEntity, "Failed to create stack. Please try again.")
			}

			return
		}

		renderSuccess("Stack '" + name + "' created successfully.")
	})
}

func handleSidebarStackCreate(
	logger *slog.Logger,
	tmpl *template.Template,
	stackService *stackApp.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		renderError := func(status int, message string) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("HX-Retarget", "#toast-host")
			w.Header().Set("HX-Reswap", "innerHTML")
			w.WriteHeader(status)

			if err := tmpl.ExecuteTemplate(w, "shared/partials/toast/error", alertData{
				Message: message,
			}); err != nil {
				logger.Error("render sidebar stack create error", "error", err.Error())
			}
		}

		session, ok := commonauth.SessionFromContext(ctx)
		if !ok || session == nil || !session.IsValid {
			renderError(http.StatusUnauthorized, "Unauthorized. Please log in to create a stack.")
			return
		}

		user := userDomain.NewUser(session.UserID, session.Email)
		stack := stackDomain.NewStack(r.FormValue("name"), user)

		if err := stackService.Create(ctx, *stack); err != nil {
			logger.Error("create sidebar stack", "error", err.Error())

			switch {
			case errors.Is(err, stackDomain.ErrInvalidStack):
				renderError(http.StatusUnprocessableEntity, "Invalid stack.")
			case errors.Is(err, stackDomain.ErrInvalidName):
				renderError(http.StatusUnprocessableEntity, invalidStackNameMessage)
			case errors.Is(err, stackDomain.ErrStackAlreadyExists):
				renderError(http.StatusConflict, "A stack with this name already exists.")
			default:
				renderError(http.StatusInternalServerError, "Failed to create stack: "+err.Error())
			}
			return
		}

		w.Header().Set("HX-Redirect", "/app/stacks/detail/"+stack.UUID)
		w.WriteHeader(http.StatusCreated)
	})
}

func handleStackDelete(
	logger *slog.Logger,
	tmpl *template.Template,
	stackService *stackApp.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		renderError := func(status int, message string) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("HX-Retarget", "#toast-host")
			w.Header().Set("HX-Reswap", "innerHTML")
			w.WriteHeader(status)

			if err := tmpl.ExecuteTemplate(w, "shared/partials/toast/error", alertData{
				Message: message,
			}); err != nil {
				logger.Error("render stack delete error", "error", err.Error())
			}
		}

		session, ok := commonauth.SessionFromContext(ctx)
		if !ok || session == nil || !session.IsValid {
			renderError(http.StatusUnauthorized, "Unauthorized. Please log in to delete this stack.")
			return
		}

		stackUUID := r.PathValue("uuid")
		stack, err := stackService.GetByUUID(ctx, stackUUID)
		if err != nil {
			logger.Error("get stack before delete", "error", err.Error())
			renderError(http.StatusInternalServerError, "Failed to delete stack: "+err.Error())
			return
		}

		if stack.Owner.ExternalID != session.UserID {
			renderError(http.StatusForbidden, "You are not allowed to delete this stack.")
			return
		}

		if err := stackService.Delete(ctx, stackUUID); err != nil {
			logger.Error("delete stack", "error", err.Error())
			renderError(http.StatusInternalServerError, "Failed to delete stack: "+err.Error())
			return
		}

		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", stacksPageURL)
			w.WriteHeader(http.StatusOK)
			return
		}

		http.Redirect(w, r, stacksPageURL, http.StatusSeeOther)
	})
}
