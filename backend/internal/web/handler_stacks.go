package web

import (
	"errors"
	"html/template"
	"io"
	"log/slog"
	"mime"
	"net/http"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	paperApp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	"github.com/paperstacks.io/paperstacks/internal/paper/bibliography"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	stackApp "github.com/paperstacks.io/paperstacks/internal/stack/application"
	stackDomain "github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userApp "github.com/paperstacks.io/paperstacks/internal/user/application"
)

const invalidStackNameMessage = "Stack name must be 1-80 characters and contain only letters, numbers, spaces, or - _ . , : ' & / ( ) + #."

const stacksPageURL = "/app/stacks/page"

const (
	oneMiB int64 = 1 << 20
	tenMiB int64 = 10 << 20
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

		stack, err := stackService.GetByUUID(ctx, id)
		if err != nil {
			if isHTMX(r) {
				hxRedirect(w, stacksPageURL, http.StatusOK)
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
			pageData:      newPageData(r, hankoAPIURL),
			Stack:         stack,
			SelectedPaper: selectedPaper,
			CreatedAt:     stack.CreatedAt.Format("2006-01-02 15:04"),
			UpdatedAt:     stack.UpdatedAt.Format("2006-01-02 15:04"),
		}

		renderTemplate(w, r, tmpl, data)
	})
}

func handleStackBibLaTeXExport(
	logger *slog.Logger,
	stackService *stackApp.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stackUUID := r.PathValue("uuid")
		stack, err := stackService.GetByUUID(r.Context(), stackUUID)
		if err != nil {
			if errors.Is(err, stackDomain.ErrStackNotFound) {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			logger.Error("get stack for BibLaTeX export", "stackUUID", stackUUID, "error", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		document, err := bibliography.ExportBibLaTeX(stack.Papers)
		if err != nil {
			logger.Error("export stack as BibLaTeX", "stackUUID", stackUUID, "error", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/x-bibtex; charset=utf-8")
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
			"filename": stack.Name + ".bib",
		}))
		_, _ = w.Write(document)
	})
}

func handleStackBibLaTeXImport(
	logger *slog.Logger,
	tmpl *template.Template,
	stackService *stackApp.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session, ok := commonauth.SessionFromContext(ctx)
		if !ok || session == nil || !session.IsValid {
			_ = renderErrorToast(w, tmpl, "Please log in before importing papers.")
			return
		}

		stackUUID := r.PathValue("uuid")
		stack, err := stackService.GetByUUID(ctx, stackUUID)
		if err != nil {
			if !errors.Is(err, stackDomain.ErrStackNotFound) {
				logger.Error("get stack before BibLaTeX import", "stackUUID", stackUUID, "error", err.Error())
			}
			_ = renderErrorToast(w, tmpl, "The requested stack could not be found.")
			return
		}
		if stack.Owner.ExternalID != session.UserID {
			_ = renderErrorToast(w, tmpl, "You are not allowed to import papers into this stack.")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, tenMiB+oneMiB)
		if err := r.ParseMultipartForm(oneMiB); err != nil {
			if r.MultipartForm != nil {
				_ = r.MultipartForm.RemoveAll()
			}
			_ = renderErrorToast(w, tmpl, "Upload a .bib file no larger than 10 MiB.")
			return
		}
		defer r.MultipartForm.RemoveAll()

		file, _, err := r.FormFile("bib_file")
		if err != nil {
			_ = renderErrorToast(w, tmpl, "Choose a .bib file to import.")
			return
		}
		defer file.Close()

		source, err := io.ReadAll(io.LimitReader(file, tenMiB+1))
		if err != nil {
			logger.Error("read BibLaTeX upload", "stackUUID", stackUUID, "error", err.Error())
			_ = renderErrorToast(w, tmpl, "The uploaded file could not be read.")
			return
		}

		switch {
		case len(source) == 0:
			_ = renderErrorToast(w, tmpl, "The uploaded .bib file is empty.")
			return
		case int64(len(source)) > tenMiB:
			_ = renderErrorToast(w, tmpl, "Choose a .bib file no larger than 10 MiB.")
			return
		}

		candidates, importErr := bibliography.ImportBibLaTeX(source)
		if importErr != nil {
			if errors.Is(importErr, bibliography.ErrInvalidBibLaTeX) {
				_ = renderErrorToast(w, tmpl, "The file is not valid BibLaTeX. Check its syntax and try again.")
			} else {
				logger.Error("import BibLaTeX", "stackUUID", stackUUID, "error", importErr.Error())
				_ = renderErrorToast(w, tmpl, "The import stopped because of a server error.")
			}
			return
		}

		result, err := stackService.Import(ctx, stackUUID, candidates.Imported)
		data := struct {
			AlreadyInStack       []paperDomain.Paper
			CreatedPaperAndAdded []paperDomain.Paper
			ExistingPaperAdded   []paperDomain.Paper
			Failed               []bibliography.PaperEntry
		}{
			AlreadyInStack:       result.AlreadyInStack,
			CreatedPaperAndAdded: result.CreatedPaperAndAdded,
			ExistingPaperAdded:   result.ExistingPaperAdded,
			Failed:               candidates.Failed,
		}
		if err != nil {
			logger.Error("stack service import failed", "stackUUID", stackUUID, "error", err.Error())
			_ = renderErrorToast(w, tmpl, "The import stopped because of a server error.")

		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "stacks/partials/biblatex-import-results", data); err != nil {
			logger.Error("render BibLaTeX import result", "error", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
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

		counter, err := stackService.CountPublic(r.Context())
		if err != nil {
			logger.Error("count public stacks", "error", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		data := stacksPageData{
			pageData:          newPageData(r, hankoAPIURL),
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

func handleSidebarStackCreate(
	logger *slog.Logger,
	tmpl *template.Template,
	userService *userApp.UserService,
	stackService *stackApp.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		session, ok := commonauth.SessionFromContext(ctx)
		if !ok || session == nil || !session.IsValid {
			renderErrorToast(w, tmpl, "Unauthorized. Please log in to create a stack.")
			return
		}

		owner, err := userService.GetByExternalID(ctx, session.UserID)
		if err != nil {
			logger.Error("read sidebar stack owner", "userId", session.UserID, "error", err)
			renderErrorToast(w, tmpl, "Failed to create stack: "+err.Error())
			return
		}

		name := r.FormValue("name")
		stack, err := stackService.CreateByName(ctx, name, owner)
		if err != nil {
			logger.Error("create sidebar stack", "error", err.Error())

			switch {
			case errors.Is(err, stackDomain.ErrStackAlreadyExists):
				renderErrorToast(w, tmpl, "A stack with this name already exists.")
			case errors.Is(err, stackDomain.ErrInvalidName):
				renderErrorToast(w, tmpl, invalidStackNameMessage)
			default:
				renderErrorToast(w, tmpl, "Failed to create stack: "+err.Error())
			}
			return
		}

		hxRedirect(w, "/app/stacks/detail/"+stack.UUID, http.StatusCreated)
	})
}

func handleStackDelete(
	logger *slog.Logger,
	tmpl *template.Template,
	stackService *stackApp.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		session, ok := commonauth.SessionFromContext(ctx)
		if !ok || session == nil || !session.IsValid {
			renderErrorToast(w, tmpl, "Unauthorized. Please log in to delete this stack.")
			return
		}

		stackUUID := r.PathValue("uuid")
		stack, err := stackService.GetByUUID(ctx, stackUUID)
		if err != nil {
			logger.Error("get stack before delete", "error", err.Error())
			renderErrorToast(w, tmpl, "Failed to delete stack: "+err.Error())
			return
		}

		if stack.Owner.ExternalID != session.UserID {
			renderErrorToast(w, tmpl, "You are not allowed to delete this stack.")
			return
		}

		if err := stackService.Delete(ctx, stackUUID); err != nil {
			logger.Error("delete stack", "error", err.Error())
			renderErrorToast(w, tmpl, "Failed to delete stack: "+err.Error())
			return
		}

		if isHTMX(r) {
			hxRedirect(w, stacksPageURL, http.StatusOK)
			return
		}

		http.Redirect(w, r, stacksPageURL, http.StatusSeeOther)
	})
}

func handleStackPaperRemove(
	logger *slog.Logger,
	tmpl *template.Template,
	stackService *stackApp.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		session, ok := commonauth.SessionFromContext(ctx)
		if !ok || session == nil || !session.IsValid {
			renderErrorToast(w, tmpl, "Unauthorized. Please log in to remove papers from this stack.")
			return
		}

		stackUUID := r.PathValue("uuid")
		paperUUID := r.PathValue("paperUUID")

		stack, err := stackService.GetByUUID(ctx, stackUUID)
		if err != nil {
			logger.Error("get stack before paper removal", "stackUUID", stackUUID, "error", err.Error())
			renderErrorToast(w, tmpl, "Failed to remove paper from stack: "+err.Error())
			return
		}

		if stack.Owner.ExternalID != session.UserID {
			renderErrorToast(w, tmpl, "You are not allowed to remove papers from this stack.")
			return
		}

		if err := stackService.RemovePaper(ctx, stackUUID, paperUUID); err != nil {
			logger.Error("remove paper from stack", "stackUUID", stackUUID, "paperUUID", paperUUID, "error", err.Error())
			renderErrorToast(w, tmpl, "Failed to remove paper from stack: "+err.Error())
			return
		}

		detailURL := "/app/stacks/detail/" + stackUUID
		if isHTMX(r) {
			hxRedirect(w, detailURL, http.StatusOK)
			return
		}

		http.Redirect(w, r, detailURL, http.StatusSeeOther)
	})
}

func handleStackPublicSettingUpdate(
	logger *slog.Logger,
	tmpl *template.Template,
	stackService *stackApp.StackService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		session, ok := commonauth.SessionFromContext(ctx)
		if !ok || session == nil || !session.IsValid {
			renderErrorToast(w, tmpl, "Unauthorized. Please log in to update this stack.")
			return
		}

		stackUUID := r.PathValue("uuid")

		stack, err := stackService.GetByUUID(ctx, stackUUID)
		if err != nil {
			logger.Error("get stack before public setting update", "error", err.Error())
			renderErrorToast(w, tmpl, "Failed to stack setting: "+err.Error())
			return
		}

		if stack.Owner.ExternalID != session.UserID {
			renderErrorToast(w, tmpl, "You are not allowed to update this stack.")
			return
		}

		stack.IsPublic = r.FormValue("is_public") == "on"
		if _, err := stackService.Update(ctx, stack); err != nil {
			logger.Error("update stack public setting", "error", err.Error())
			renderErrorToast(w, tmpl, "Failed to stack setting: "+err.Error())
			return
		}

		renderSuccessToast(w, tmpl, "Stack changes saved")
	})
}
