package web

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	paperApp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

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

func handlePaperUploadPDF(
	logger *slog.Logger,
	tmpl *template.Template,
	paperService *paperApp.PaperService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, ok := commonauth.SessionFromContext(r.Context())
		if !ok || session == nil || !session.IsValid {
			_ = renderErrorToast(w, tmpl, "Please log in before uploading a PDF.")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, paperDomain.PDFMaxFileSize+oneMiB)
		if err := r.ParseMultipartForm(oneMiB); err != nil {
			if r.MultipartForm != nil {
				_ = r.MultipartForm.RemoveAll()
			}
			_ = renderErrorToast(w, tmpl, "Upload a PDF file no larger than 10 MiB.")
			return
		}
		defer r.MultipartForm.RemoveAll()

		file, header, err := r.FormFile("file")
		if err != nil {
			_ = renderErrorToast(w, tmpl, "Choose a PDF file to upload.")
			return
		}
		defer file.Close()

		_, err = paperService.UploadPDF(r.Context(), paperDomain.UploadPDFOptions{
			PaperUUID: r.PathValue("uuid"),
			FileName:  header.Filename,
			File:      file,
			Size:      header.Size,
		})
		if err != nil {
			switch {
			case errors.Is(err, paperDomain.ErrFileSizeExceeded):
				_ = renderErrorToast(w, tmpl, "Upload a PDF file no larger than 10 MiB.")
			case errors.Is(err, paperDomain.ErrInvalidFileType):
				_ = renderErrorToast(w, tmpl, "The uploaded file is not a valid PDF.")
			case errors.Is(err, paperDomain.ErrPaperNotFound):
				_ = renderErrorToast(w, tmpl, "The requested paper could not be found.")
			default:
				logger.Error("upload paper PDF", "paperUUID", r.PathValue("uuid"), "error", err.Error())
				_ = renderErrorToast(w, tmpl, "The PDF could not be uploaded. Please try again.")
			}
			return
		}

		_ = renderSuccessToast(w, tmpl, "PDF uploaded successfully.")
	})
}
