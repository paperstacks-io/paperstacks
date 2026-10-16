package web

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/paperstacks.io/paperstacks/internal/common/build"
	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	docApp "github.com/paperstacks.io/paperstacks/internal/document/application"
	"github.com/paperstacks.io/paperstacks/internal/document/domain"
	paperApp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
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
		result, err := paperService.Search(r.Context(), opts)
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

func handleDocumentUpload(
	logger *slog.Logger,
	documentService *docApp.DocumentService,
	templates *template.Template,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		session, ok := commonauth.SessionFromContext(ctx)
		if !ok || session == nil || !session.IsValid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			if errors.Is(err, http.ErrMissingFile) {
				http.Error(w, "file is required", http.StatusBadRequest)
				return
			}
			http.Error(w, "invalid multipart form data", http.StatusBadRequest)
			return
		}
		defer file.Close()

		paperUUID := r.FormValue("paper_uuid")
		if paperUUID == "" {
			http.Error(w, "paper_uuid is required", http.StatusBadRequest)
			logger.Warn("PDF upload missing paper UUID")

			if err := renderErrorToast(w, templates, "Paper UUID is required."); err != nil {
				logger.Error("render upload toast", "error", err)
			}
			return
		}

		fileName := header.Filename
		if fileName == "" {
			http.Error(w, "file_name is required", http.StatusBadRequest)
			return
		}

		opts := domain.UploadOptions{
			PaperUUID: paperUUID,
			FileName:  fileName,
			UserID:    session.UserID,
			File:      file,
			Size:      header.Size,
			Config: domain.PDFConfiguration{
				MaxFileSize:     10 * 1024 * 1024,
				Signature:       "%PDF-",
				ContentType:     "application/pdf",
				SniffBufferSize: 512,
			},
		}

		_, err = documentService.Upload(ctx, opts)
		if err != nil {
			if errors.Is(err, domain.ErrFileSizeExceeded) || errors.Is(err, domain.ErrInvalidFileType) || errors.Is(err, paperDomain.ErrPaperNotFound) {
				if err := renderErrorToast(w, templates, err.Error()); err != nil {
					logger.Error("render upload toast", "error", err)
				}
				return
			}
			logger.Error("upload document", "user_id", session.UserID, "error", err)
			if err := renderErrorToast(w, templates, "Failed to upload PDF."); err != nil {
				logger.Error("render upload toast", "error", err)
			}
			return
		}

		logger.Info("PDF uploaded successfully", "paper_uuid", paperUUID, "file_name", header.Filename)

		if err := renderSuccessToast(w, templates, "PDF uploaded successfully."); err != nil {
			logger.Error("render upload toast", "error", err)
		}
	})
}
