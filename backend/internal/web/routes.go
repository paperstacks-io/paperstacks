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

	"github.com/paperstacks.io/paperstacks/internal/common/config"
	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	"github.com/paperstacks.io/paperstacks/internal/common/server/middleware"
	paperApp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	stackApp "github.com/paperstacks.io/paperstacks/internal/stack/application"
	stackDomain "github.com/paperstacks.io/paperstacks/internal/stack/domain"
	webauth "github.com/paperstacks.io/paperstacks/internal/web/auth"
)

//go:embed assets/* all:templates
var content embed.FS

type navItem struct {
	Label  string
	Path   string
	Active bool
}

type navStackItem struct {
	Name    string
	Size    int
	Path    string
	Active  bool
	HasMore bool
}

func navItems(activePath string) []navItem {
	prefix := "/app"
	items := []navItem{
		{Label: "Home", Path: prefix + "/"},
		{Label: "Papers", Path: prefix + "/papers"},
		{Label: "Stacks", Path: prefix + "/stacks"},
		{Label: "Search", Path: prefix + "/search"},
		{Label: "Settings", Path: prefix + "/settings"},
	}

	for i := range items {
		items[i].Active = items[i].Path == activePath
	}

	return items
}

func navStackItems(stacks []stackDomain.Stack) []navStackItem {
	testPaths := []string{
		"/papers",
		"/stacks",
		"/search",
		"/settings",
	}

	prefix := "/app"
	items := make([]navStackItem, 0, 4)

	for i, stack := range stacks {
		if i >= 4 {
			break
		}

		items = append(items, navStackItem{
			Name: stack.Name,
			Size: len(stack.Papers),
			Path: prefix + testPaths[i],
		})
	}

	if len(items) > 0 {
		items[len(items)-1].HasMore = len(stacks) > 4
	}

	return items
}

func AddRoute(
	mux *http.ServeMux,
	cfg config.Config,
	logger *slog.Logger,
	paperService *paperApp.PaperService,
	stackService *stackApp.StackService,
	sessionService commonauth.SessionService,
) error {
	templateFiles, err := templateFiles(content)
	if err != nil {
		return err
	}

	tmpl := template.Must(template.ParseFS(content, templateFiles...))

	homeTemplate, err := pageTemplateSet(tmpl, "home")
	if err != nil {
		return err
	}

	assets, err := fs.Sub(content, "assets")
	if err != nil {
		return fmt.Errorf("load web assets: %w", err)
	}

	defaultMiddle := middleware.NewDefault(logger, sessionService)
	requireAuthMiddle := webauth.RequireAuthWebMiddleware()

	mux.Handle(
		http.MethodGet+" /{$}",
		defaultMiddle(handleIndex(
			logger,
			homeTemplate,
			navItems("/app/"),
			cfg.HankoAPIURL,
			stackService,
		)),
	)

	for _, page := range []struct {
		path         string
		template     string
		requiresAuth bool
	}{
		{path: "/papers", template: "paper", requiresAuth: false},
		{path: "/stacks", template: "stack", requiresAuth: false},
		{path: "/search", template: "search", requiresAuth: false},
		{path: "/settings", template: "settings", requiresAuth: true},
		{path: "/auth", template: "auth", requiresAuth: false},
	} {
		pageTemplate, err := pageTemplateSet(tmpl, page.template)
		if err != nil {
			return err
		}

		pageHandler := handleIndex(
			logger,
			pageTemplate,
			navItems(page.path),
			cfg.HankoAPIURL,
			stackService,
		)

		if page.requiresAuth {
			pageHandler = requireAuthMiddle(pageHandler)
		}
		pageHandler = defaultMiddle(pageHandler)

		mux.Handle(http.MethodGet+" "+page.path, pageHandler)
	}

	mux.Handle(http.MethodPost+" /auth/logout", defaultMiddle(handleLogout(logger, sessionService)))

	mux.Handle(http.MethodGet+" /assets/", defaultMiddle(http.StripPrefix("/assets/", http.FileServerFS(assets))))
	mux.Handle(http.MethodPost+" /papers/search", defaultMiddle(handlePapersSearch(logger, tmpl, paperService)))
	mux.Handle(http.MethodPost+" /stacks/search", defaultMiddle(handleStacksSearch(logger, tmpl, stackService)))
	mux.Handle(http.MethodPost+" /stacks/create", defaultMiddle(handleStacksCreate(logger, tmpl, stackService)))

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
