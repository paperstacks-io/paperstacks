package ports

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	"github.com/paperstacks.io/paperstacks/internal/paper/application"
	"github.com/paperstacks.io/paperstacks/internal/paper/infrastructure/persistence/memory"
)

type HttpServer struct {
	application application.Application
}

func NewHttpServer(app application.Application) *HttpServer {
	return &HttpServer{
		application: app,
	}
}

func (h *HttpServer) HandleReadPapers(ctx context.Context, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				w.Header().Set("Allow", http.MethodGet)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			papers, err := h.application.Queries.ReadPapers.Handle(ctx)
			if err != nil {
				if errors.Is(err, memory.ErrPaperAlreadyExists) {
					logger.Error("read papers", "error", "papers "+err.Error())
					http.Error(w, "papers "+err.Error(), http.StatusNotFound)
					return
				}

				logger.Error("read papers", "error", "papers "+err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := server.Encode(w, r, http.StatusOK, server.PapersToResponse(papers)); err != nil {
				logger.Error("encode paper_ response", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}
