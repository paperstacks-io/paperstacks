// Package main runs the paperstacks HTTP server binary.
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/paperstacks.io/paperstacks/internal/common/build"
	"github.com/paperstacks.io/paperstacks/internal/common/config"
	"github.com/paperstacks.io/paperstacks/internal/common/objectstorage"
	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	docApp "github.com/paperstacks.io/paperstacks/internal/document/application"
	docHttp "github.com/paperstacks.io/paperstacks/internal/document/http"
	docMem "github.com/paperstacks.io/paperstacks/internal/document/repository/memory"
	doiApp "github.com/paperstacks.io/paperstacks/internal/doi/application"
	doiHttp "github.com/paperstacks.io/paperstacks/internal/doi/http"
	paperApp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	paperHttp "github.com/paperstacks.io/paperstacks/internal/paper/http"
	paperMem "github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
	"github.com/paperstacks.io/paperstacks/internal/server"
	stackApp "github.com/paperstacks.io/paperstacks/internal/stack/application"
	stackHttp "github.com/paperstacks.io/paperstacks/internal/stack/http"
	stackMem "github.com/paperstacks.io/paperstacks/internal/stack/repository/memory"
	userApp "github.com/paperstacks.io/paperstacks/internal/user/application"
	userHttp "github.com/paperstacks.io/paperstacks/internal/user/http"
	userMem "github.com/paperstacks.io/paperstacks/internal/user/repository/memory"
	"github.com/paperstacks.io/paperstacks/internal/web"
)

func run(
	ctx context.Context,
	cfg config.Config,
) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	paperRepo := paperMem.NewRepository()
	paperService := paperApp.NewPaperService(paperRepo)
	bibliographyService := paperApp.NewBibliographyService(paperRepo)
	doiService := doiApp.NewDOIService(nil)
	stackService := stackApp.NewStackService(stackMem.NewRepository(), paperService)
	userService := userApp.NewUserService(userMem.NewRepository())
	userProvisioner := userApp.NewUserProvisioner(userService, stackService, cfg.HankoAPIURL, http.DefaultClient)
	sessionService := commonauth.NewHankoSessionService(cfg.HankoAPIURL, userProvisioner, http.DefaultClient)
	docRepo := docMem.NewRepository()
	docStorage := docMem.NewStorage()
	documentService := docApp.NewDocumentService(docRepo, docStorage, paperService)

	if ok, _ := cfg.ObjectStorage.Validate(); ok {
		objectStore, err := objectstorage.NewS3Store(cfg.ObjectStorage, "paper", logger)
		if err != nil {
			logger.Error("ObjectStorage config error", "error", err)
			return err
		}

		if err := objectStore.Check(ctx); err != nil {
			logger.Error("ObjectStorage check error", "error", err)
			return err
		}
	}

	rootMux := http.NewServeMux()
	apiMux := http.NewServeMux()
	webMux := http.NewServeMux()
	server.AddRoute(rootMux, ctx, logger, sessionService)
	paperHttp.AddPaperRoute(apiMux, logger, paperService, sessionService)
	doiHttp.AddDOIRoute(apiMux, logger, doiService, sessionService)
	userHttp.AddUserRoute(apiMux, logger, userService, userProvisioner, stackService, sessionService)
	stackHttp.AddStackRoute(apiMux, logger, stackService, sessionService)
	docHttp.UploadDocumentRoute(apiMux, logger, documentService, sessionService)
	web.AddRoute(webMux, cfg, logger, paperService, bibliographyService, stackService, userService, sessionService)
	rootMux.Handle("/api/", http.StripPrefix("/api", apiMux))
	rootMux.Handle("/app/", http.StripPrefix("/app", webMux))

	httpServer := &http.Server{
		Addr:         net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:      rootMux,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		slog.Info("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		httpServer.SetKeepAlivesEnabled(false)
		if err := httpServer.Shutdown(ctx); err != nil {
			slog.Error("Could not gracefully shutdown the server", slog.String("error", err.Error()))
		}
		close(done)
	}()

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "\tStarting server:")
	fmt.Fprintln(tw, "\tFrontend: \thttp://"+cfg.Host+":"+cfg.Port+"/app/")
	fmt.Fprintln(tw, "\tAPI: \thttp://"+cfg.Host+":"+cfg.Port+"/api/")
	fmt.Fprintln(tw)
	_ = tw.Flush()

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Could not listen", slog.String("port", cfg.Port), slog.String("error", err.Error()))
		return err
	}

	<-done
	slog.Info("Server stopped")

	return nil
}

const bannerLogo = `
                                 _             _          _       
                                | |           | |        (_)      
 _ __   __ _ _ __   ___ _ __ ___| |_ __ _  ___| | _____   _  ___  
| '_ \ / _' | '_ \ / _ \ '__/ __| __/ _' |/ __| |/ / __| | |/ _ \ 
| |_) | (_| | |_) |  __/ |  \__ \ || (_| | (__|   <\__ \_| | (_) |
| .__/ \__,_| .__/ \___|_|  |___/\__\__,_|\___|_|\_\___(_)_|\___/ 
| |         | |                                                   
|_|         |_|                                                   
`

func banner(w io.Writer, cfg config.Config) {
	storageStatus := "Not configured or invalid configuration"
	if ok, _ := cfg.ObjectStorage.Validate(); ok {
		storageStatus = cfg.ObjectStorage.Endpoint
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	fmt.Fprint(tw, bannerLogo)
	fmt.Fprintln(tw)
	fmt.Fprintln(tw, "  Version:\t"+build.Version)
	fmt.Fprintln(tw, "  Git hash:\t"+build.GitHash)
	fmt.Fprintln(tw)
	fmt.Fprintln(tw, "  Hanko API URL:\t"+cfg.HankoAPIURL)
	fmt.Fprintln(tw, "  Object Storage URL:\t"+storageStatus)
	fmt.Fprintln(tw)

	_ = tw.Flush()
}

func main() {
	cfg := config.New()
	banner(os.Stdout, cfg)

	if err := run(context.Background(), cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
