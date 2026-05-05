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
	doiApp "github.com/paperstacks.io/paperstacks/internal/doi/application"
	doiHttp "github.com/paperstacks.io/paperstacks/internal/doi/http"
	paperApp "github.com/paperstacks.io/paperstacks/internal/paper/application"
	phttp "github.com/paperstacks.io/paperstacks/internal/paper/http"
	"github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
	"github.com/paperstacks.io/paperstacks/internal/server"
	"github.com/paperstacks.io/paperstacks/internal/web"
)

func run(
	ctx context.Context,
	cfg config.Config,
) error {
	host := cfg.Host
	port := cfg.Port

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	paperService := paperApp.NewPaperService(memory.NewRepository())
	doiService := doiApp.NewDOIService(nil)

	rootMux := http.NewServeMux()
	apiMux := http.NewServeMux()
	webMux := http.NewServeMux()
	server.AddRoute(
		rootMux,
		ctx,
		logger,
	)
	phttp.AddPaperRoute(
		apiMux,
		logger,
		paperService,
	)

	doiHttp.AddDOIRoute(apiMux, logger, doiService)
	web.AddRoute(webMux, logger, paperService)
	rootMux.Handle("/api/", http.StripPrefix("/api", apiMux))
	rootMux.Handle("/app/", http.StripPrefix("/app", webMux))

	httpServer := &http.Server{
		Addr:         net.JoinHostPort(host, port),
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
	fmt.Fprintln(tw, "\tFrontend: \thttp://"+host+":"+port+"/app/")
	fmt.Fprintln(tw, "\tAPI: \thttp://"+host+":"+port+"/api/")
	fmt.Fprintln(tw)
	_ = tw.Flush()

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Could not listen", slog.String("port", port), slog.String("error", err.Error()))
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

func banner(w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	fmt.Fprint(tw, bannerLogo)
	fmt.Fprintln(tw)
	fmt.Fprintln(tw, "  Version:\t"+build.Version)
	fmt.Fprintln(tw, "  Git hash:\t"+build.GitHash)
	fmt.Fprintln(tw, "  Build time:\t"+build.BuildTime)
	fmt.Fprintln(tw)

	_ = tw.Flush()
}

func main() {
	banner(os.Stdout)

	conf := config.New()

	if err := run(context.Background(), conf); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
