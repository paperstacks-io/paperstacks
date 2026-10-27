package web

import (
	"html/template"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strconv"
	"strings"

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

func handlePaperDownloadPDF(
	logger *slog.Logger,
	paperService *paperApp.PaperService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paperUUID := r.PathValue("paperUUID")
		if paperUUID == "" {
			http.Error(w, "missing paper uuid", http.StatusBadRequest)
			return
		}

		paper, err := paperService.GetByUUID(r.Context(), paperUUID)
		if err != nil {
			logger.Error("get paper before PDF download", "paperUUID", paperUUID, "error", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if paper.PDFStorageKey == "" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		pdf, err := paperService.OpenPDF(r.Context(), paper.PDFStorageKey)
		if err != nil {
			logger.Error("open paper PDF", "paperUUID", paperUUID, "error", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer pdf.Body.Close()

		w.Header().Set("Content-Type", paperDomain.PDFContentType)
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
			"filename": pdf.FileName,
		}))
		if _, err := io.Copy(w, pdf.Body); err != nil {
			logger.Error("stream paper PDF", "paperUUID", paperUUID, "error", err.Error())
		}
	})
}
