package http_test

import (
	"context"
	"encoding/json"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/user/application"
	userHttp "github.com/paperstacks.io/paperstacks/internal/user/http"
	"github.com/paperstacks.io/paperstacks/internal/user/repository/memory"
)

func TestGetUserByID(t *testing.T) {
	t.Parallel()

	repo := memory.NewRepository()
	service := application.NewUserService(repo, "", nil)
	created, err := service.CreateIfNotExist(context.Background(), "external-1", "one@example.com")
	if err != nil {
		t.Fatalf("CreateIfNotExist() error = %v, want nil", err)
	}

	mux := newUserMux(service)
	req := httptest.NewRequest(nethttp.MethodGet, "/users/external-1", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, nethttp.StatusOK)
	}

	var response userHttp.UserResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ExternalID != created.ExternalID {
		t.Fatalf("ExternalID = %q, want %q", response.ExternalID, created.ExternalID)
	}
}

func TestGetUserByIDReturnsNotFound(t *testing.T) {
	t.Parallel()

	service := application.NewUserService(memory.NewRepository(), "", nil)
	mux := newUserMux(service)
	req := httptest.NewRequest(nethttp.MethodGet, "/users/missing", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != nethttp.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.Code, nethttp.StatusNotFound)
	}
}

func TestGetCurrentUserRequiresBearerToken(t *testing.T) {
	t.Parallel()

	service := application.NewUserService(memory.NewRepository(), "", nil)
	mux := newUserMux(service)
	req := httptest.NewRequest(nethttp.MethodGet, "/users/me", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != nethttp.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, nethttp.StatusUnauthorized)
	}
}

func TestGetCurrentUser(t *testing.T) {
	t.Parallel()

	authServer := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer session-token" {
			t.Fatalf("Authorization = %q, want bearer token", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"user_id":"external-1","emails":[{"address":"one@example.com","is_primary":true}]}`))
	}))
	t.Cleanup(authServer.Close)

	service := application.NewUserService(memory.NewRepository(), authServer.URL, authServer.Client())
	mux := newUserMux(service)
	req := httptest.NewRequest(nethttp.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer session-token")
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, nethttp.StatusOK)
	}

	var response userHttp.UserResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ExternalID != "external-1" {
		t.Fatalf("ExternalID = %q, want %q", response.ExternalID, "external-1")
	}
}

func newUserMux(service *application.UserService) *nethttp.ServeMux {
	mux := nethttp.NewServeMux()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	userHttp.AddUserRoute(mux, logger, service)
	return mux
}
