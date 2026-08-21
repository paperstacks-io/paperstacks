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
	userApp "github.com/paperstacks.io/paperstacks/internal/user/application"
	webauth "github.com/paperstacks.io/paperstacks/internal/web/auth"
)

//go:embed assets/* all:templates
var content embed.FS

func AddRoute(
	mux *http.ServeMux,
	cfg config.Config,
	logger *slog.Logger,
	paperService *paperApp.PaperService,
	stackService *stackApp.StackService,
	userService *userApp.UserService,
	sessionService commonauth.SessionService,
) error {
	templateFiles, err := templateFiles(content)
	if err != nil {
		return err
	}

	tmpl := template.Must(template.ParseFS(content, templateFiles...))

	assets, err := fs.Sub(content, "assets")
	if err != nil {
		return fmt.Errorf("load web assets: %w", err)
	}

	defaultMiddle := middleware.NewDefault(logger, sessionService)
	requireAuthMiddle := webauth.RequireAuthWebMiddleware()
	authenticated := func(handler http.Handler) http.Handler {
		return defaultMiddle(requireAuthMiddle(handler))
	}

	pageTemplate := func(name string) *template.Template {
		pageTemplate, err := pageTemplateSet(tmpl, name)
		if err != nil {
			panic(fmt.Errorf("load %s page template: %w", name, err))
		}

		return pageTemplate
	}

	// Pages
	homeTmpl := pageTemplate("home/page")
	mux.Handle(http.MethodGet+" /{$}", defaultMiddle(handlePage(homeTmpl, cfg.HankoAPIURL)))

	papersTmpl := pageTemplate("papers/page")
	mux.Handle(http.MethodGet+" /papers", defaultMiddle(handlePage(papersTmpl, cfg.HankoAPIURL)))
	mux.Handle(http.MethodPost+" /papers/search", defaultMiddle(handlePapersSearch(logger, tmpl, paperService)))

	stacksMyTmpl := pageTemplate("stacks/page")
	mux.Handle(http.MethodGet+" /stacks/page", authenticated(handleStacksPage(logger, stacksMyTmpl, cfg.HankoAPIURL, stackService)))
	bibImportTmpl := pageTemplate("stacks/import-biblatex")
	mux.Handle(http.MethodGet+" /stacks/{uuid}/import/biblatex", authenticated(handleStacksDetailPage(logger, bibImportTmpl, cfg.HankoAPIURL, stackService)))
	mux.Handle(http.MethodPost+" /stacks/{uuid}/import/biblatex", authenticated(handleStackBibLaTeXImport(logger, tmpl, stackService)))
	stacksCiteTmpl := pageTemplate("stacks/cite")
	mux.Handle(http.MethodGet+" /stacks/detail/{stackUUID}/papers/{paperUUID}/cite", authenticated(handleStacksPaperCitation(logger, stacksCiteTmpl, paperService, stackService)))
	stacksDetailTmpl := pageTemplate("stacks/detail")
	mux.Handle(http.MethodGet+" /stacks/detail/{uuid}", authenticated(handleStacksDetailPage(logger, stacksDetailTmpl, cfg.HankoAPIURL, stackService)))
	mux.Handle(http.MethodGet+" /stacks/detail/{uuid}/export/biblatex", authenticated(handleStackBibLaTeXExport(logger, stackService)))
	mux.Handle(http.MethodPost+" /stacks/detail/{uuid}/delete", authenticated(handleStackDelete(logger, tmpl, stackService)))
	mux.Handle(http.MethodPost+" /stacks/detail/{uuid}/settings/is-public", authenticated(handleStackPublicSettingUpdate(logger, tmpl, stackService)))
	mux.Handle(http.MethodGet+" /stacks/detail/{stackUUID}/papers/{paperUUID}", authenticated(handleStackPaperInfo(logger, tmpl, paperService)))
	mux.Handle(http.MethodPost+" /stacks/detail/{uuid}/papers/{paperUUID}/remove", authenticated(handleStackPaperRemove(logger, tmpl, stackService)))
	mux.Handle(http.MethodPost+" /stacks/search", authenticated(handleStacksSearchByOwner(logger, tmpl, stackService)))
	mux.Handle(http.MethodPost+" /stacks/stats", authenticated(handleStacksStatsByOwner(logger, tmpl, stackService)))
	mux.Handle(http.MethodPost+" /stacks/sidebar/create", authenticated(handleSidebarStackCreate(logger, tmpl, userService, stackService)))

	settingsTmpl := pageTemplate("settings/page")
	mux.Handle(http.MethodGet+" /settings", authenticated(handleSettingsPage(logger, settingsTmpl, cfg.HankoAPIURL, userService)))
	mux.Handle(http.MethodPost+" /settings/user", authenticated(handleUserSettingsUpdate(logger, tmpl, userService)))

	authTmpl := pageTemplate("auth/page")
	mux.Handle(http.MethodGet+" /auth", defaultMiddle(handlePage(authTmpl, cfg.HankoAPIURL)))
	mux.Handle(http.MethodPost+" /auth/logout", defaultMiddle(handleLogout(logger, sessionService)))

	// Shared partials
	mux.Handle(http.MethodGet+" /partials/sidebar/stacks-list", authenticated(handleSidebarStacks(logger, homeTmpl, stackService)))
	mux.Handle(http.MethodGet+" /partials/toast/not-implemented", defaultMiddle(handlePartialWithoutData(tmpl, "shared/partials/toast/not-implemented")))
	mux.Handle(http.MethodGet+" /partials/toast/changes-saved", defaultMiddle(handlePartialWithoutData(tmpl, "shared/partials/toast/changes-saved")))

	// Static content
	mux.Handle(http.MethodGet+" /assets/", defaultMiddle(http.StripPrefix("/assets/", http.FileServerFS(assets))))

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

	if _, err := cloned.Parse(`{{ define "page-body" }}{{ template "` + pageTemplate + `" . }}{{ end }}`); err != nil {
		return nil, fmt.Errorf("parse page template alias: %w", err)
	}

	return cloned, nil
}
