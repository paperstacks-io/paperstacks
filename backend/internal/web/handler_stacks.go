package web

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/common/build"
	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	paperApp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	stackApp "github.com/paperstacks.io/paperstacks/internal/stack/application"
	stackDomain "github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

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

		if id == "" {
			http.Error(w, "missing stack uuid", http.StatusBadRequest)
			return
		}

		stack, err := stackService.GetByUUID(ctx, id)
		if err != nil {
			logger.Error("get stack by UUID", "error", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

func handleStacksMyPage(
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

func handleStacksSearchPublic(
	logger *slog.Logger,
	tmpl *template.Template,
	stackService *stackApp.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		opts := searchOptionsFromRequest(r)
		result, err := stackService.Search(r.Context(), opts)
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
