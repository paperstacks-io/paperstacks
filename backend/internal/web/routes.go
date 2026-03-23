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
	Title         string
	PageName      string
	NavItems      []navItem
	AppTargetID   string
	PageContentID string
}

type navItem struct {
	Label  string
	Path   string
	Active bool
}

func navItems(activePath string) []navItem {
	items := []navItem{
		{Label: "Home", Path: "/app/"},
		{Label: "Papers", Path: "/app/papers"},
		{Label: "Search", Path: "/app/search"},
		{Label: "Settings", Path: "/app/settings"},
	}

	for i := range items {
		items[i].Active = items[i].Path == activePath
	}

	return items
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

	homeTemplate, err := pageTemplateSet(tmpl, "home")
	if err != nil {
		return err
	}

	assets, err := fs.Sub(content, "assets")
	if err != nil {
		return fmt.Errorf("load web assets: %w", err)
	}

	defaultMiddle := middleware.NewDefault(logger)

	mux.Handle(http.MethodGet+" /app/{$}", defaultMiddle(handleIndex(homeTemplate, navItems("/app/"))))

	for _, page := range []struct {
		path     string
		template string
	}{
		{path: "/app/papers", template: "paper"},
		{path: "/app/search", template: "search"},
		{path: "/app/settings", template: "settings"},
	} {
		pageTemplate, err := pageTemplateSet(tmpl, page.template)
		if err != nil {
			return err
		}

		mux.Handle(http.MethodGet+" "+page.path, defaultMiddle(handleIndex(pageTemplate, navItems(page.path))))
	}
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

func pageTemplateSet(base *template.Template, pageTemplate string) (*template.Template, error) {
	cloned, err := base.Clone()
	if err != nil {
		return nil, fmt.Errorf("clone web templates: %w", err)
	}

	if _, err := cloned.Parse(`{{define "page-body"}}{{template "` + pageTemplate + `" .}}{{end}}`); err != nil {
		return nil, fmt.Errorf("parse page template alias: %w", err)
	}

	return cloned, nil
}
