package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/paperstacks.io/paperstacks/internal/common/server"
	"github.com/paperstacks.io/paperstacks/internal/paper/application"
	"github.com/paperstacks.io/paperstacks/internal/paper/infrastructure/persistence/memory"
)

type Server struct {
	application application.Application
}

func NewServer(app application.Application) *Server {
	return &Server{
		application: app,
	}
}

func (s *Server) HandleReadPapers(ctx context.Context, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				w.Header().Set("Allow", http.MethodGet)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			papers, err := s.application.Queries.ReadPapers.Handle(ctx)
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

			if err := server.Encode(w, r, http.StatusOK, PapersToResponse(papers)); err != nil {
				logger.Error("encode paper_ response", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}

func (s *Server) HandleReadPaper(ctx context.Context, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				w.Header().Set("Allow", http.MethodGet)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			// No DTO needed for a single ID request
			id := r.PathValue("paperId")
			if id == "" {
				http.Error(w, "missing paper_ id", http.StatusBadRequest)
				return
			}

			p, err := s.application.Queries.ReadPaper.Handle(ctx, id)
			if err != nil {
				if errors.Is(err, memory.ErrPaperNotFound) {
					logger.Error("read paper_", "paperId", id, "error", "paper_ "+err.Error())
					http.Error(w, "paper "+err.Error(), http.StatusNotFound)
					return
				}

				logger.Error("read paper", "paperId", id, "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			response := PaperToResponse(p)
			if err := server.Encode(w, r, http.StatusOK, response); err != nil {
				logger.Error("encode paper response", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)
}

func (s *Server) HandleCreatePaper(ctx context.Context, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.Header().Set("Allow", http.MethodPost)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			req, err := server.Decode[CreatePaperRequest](r)
			if err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}

			p := req.ToDomain()
			if err := s.application.Commands.CreatePaper.Handle(ctx, p); err != nil {
				if errors.Is(err, memory.ErrPaperAlreadyExists) {
					logger.Error("create paper_", "doi", p.DOI, "error", err)
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}

				logger.Error("create paper_", "doi", p.DOI, "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
		},
	)
}

func (s *Server) HandleDeletePaper(ctx context.Context, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				w.Header().Set("Allow", http.MethodDelete)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			id := r.PathValue("paperId")
			if id == "" {
				http.Error(w, "missing paper_ id", http.StatusBadRequest)
				return
			}

			if err := s.application.Commands.DeletePaper.Handle(ctx, id); err != nil {
				if errors.Is(err, memory.ErrPaperNotFound) {
					logger.Error("delete paper_", "id", id, "error", err)
					http.Error(w, "paper_ not found", http.StatusNotFound)
					return
				}

				logger.Error("delete paper_", "id", id, "error", err)
				http.Error(w, "internal server_ error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		},
	)
}

func (s *Server) HandleUpdatePaper(ctx context.Context, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				w.Header().Set("Allow", http.MethodPut)
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			id := r.PathValue("paperId")
			if id == "" {
				http.Error(w, "missing paper_ id", http.StatusBadRequest)
				return
			}

			req, err := server.Decode[UpdatePaperRequest](r)
			if err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}

			p := req.ToDomain()
			if err := s.application.Commands.UpdatePaper.Handle(ctx, id, p); err != nil {
				if errors.Is(err, memory.ErrPaperNotFound) {
					logger.Error("update paper_", "id", id, "error", "paper_ "+err.Error())
					http.Error(w, " not found", http.StatusNotFound)
					return
				}

				logger.Error("update paper_", "id", id, "error", err)
				http.Error(w, "failed to update paper_", http.StatusInternalServerError)
				return
			}
		},
	)
}
