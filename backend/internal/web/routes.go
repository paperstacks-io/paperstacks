package web

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"slices"

	"github.com/paperstacks.io/paperstacks/internal/server/middleware"
)

//go:embed assets/* all:templates
var content embed.FS

type pageData struct {
	Title string
}

func AddRoute(mux *http.ServeMux, logger *slog.Logger) error {
	templateFiles, err := templateFiles(content)
	if err != nil {
		return err
	}

	tmpl, err := template.ParseFS(content, templateFiles...)
	if err != nil {
		return fmt.Errorf("parse web templates: %w", err)
	}

	assets, err := fs.Sub(content, "assets")
	if err != nil {
		return fmt.Errorf("load web assets: %w", err)
	}

	defaultMiddle := middleware.NewDefault(logger)

	mux.Handle(http.MethodGet+" /{$}", defaultMiddle(handleIndex(tmpl)))
	mux.Handle(http.MethodGet+" /app/assets/", defaultMiddle(http.StripPrefix("/app/assets/", http.FileServerFS(assets))))

	return nil
}

func templateFiles(content fs.FS) ([]string, error) {
	files := make([]string, 0)
	err := fs.WalkDir(content, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		switch filepath.Ext(path) {
		case ".html", ".gohtml":
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list web templates: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("list web templates: no supported templates found")
	}

	slices.Sort(files)

	return files, nil
}

func handleIndex(tmpl *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if err := tmpl.ExecuteTemplate(w, "layout", pageData{Title: "Paperstacks"}); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	})
}
