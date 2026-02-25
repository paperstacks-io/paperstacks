package server

import (
	"context"
	"net/http"
)

func AddRoute(
	mux *http.ServeMux,
	ctx context.Context,
) http.Handler {
	mux.Handle("/healthz", handleHealthz())
	return mux
}
