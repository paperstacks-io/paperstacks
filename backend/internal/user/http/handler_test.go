package http_test

import (
	"context"
	"encoding/json"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"os"
	"testing"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	stackApplication "github.com/paperstacks.io/paperstacks/internal/stack/application"
	stackDomain "github.com/paperstacks.io/paperstacks/internal/stack/domain"
	stackMemory "github.com/paperstacks.io/paperstacks/internal/stack/repository/memory"
	"github.com/paperstacks.io/paperstacks/internal/user/application"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
	userHttp "github.com/paperstacks.io/paperstacks/internal/user/http"
	"github.com/paperstacks.io/paperstacks/internal/user/repository/memory"
)

func newStackService() *stackApplication.StackService {
	return stackApplication.NewStackService(stackMemory.NewRepository(), nil)
}

func TestGetUserByID(t *testing.T) {
	t.Parallel()

	repo := memory.NewRepository()
	service := application.NewUserService(repo)
	created, err := service.CreateIfNotExist(context.Background(), "external-1", "one@example.com")
	if err != nil {
		t.Fatalf("CreateIfNotExist() error = %v, want nil", err)
	}

	mux := newUserMux(service, newStackService())
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

	service := application.NewUserService(memory.NewRepository())
	mux := newUserMux(service, newStackService())
	req := httptest.NewRequest(nethttp.MethodGet, "/users/missing", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != nethttp.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.Code, nethttp.StatusNotFound)
	}
}

func TestGetCurrentUserRequiresBearerToken(t *testing.T) {
	t.Parallel()

	service := application.NewUserService(memory.NewRepository())
	mux := newUserMux(service, newStackService())
	req := httptest.NewRequest(nethttp.MethodGet, "/users/me", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != nethttp.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, nethttp.StatusUnauthorized)
	}
}

func TestGetCurrentUser(t *testing.T) {
	t.Parallel()

	service := application.NewUserService(memory.NewRepository())
	created, err := service.CreateIfNotExist(context.Background(), "external-1", "one@example.com")
	if err != nil {
		t.Fatalf("CreateIfNotExist() error = %v, want nil", err)
	}

	mux := newUserMuxWithSession(service, newStackService(), testSessionService{session: &commonauth.Session{
		UserID:  created.ExternalID,
		Email:   created.Email,
		IsValid: true,
	}})
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
	if response.ExternalID != created.ExternalID {
		t.Fatalf("ExternalID = %q, want %q", response.ExternalID, created.ExternalID)
	}
}

func TestListCurrentUserStacksRequiresBearerToken(t *testing.T) {
	t.Parallel()

	service := application.NewUserService(memory.NewRepository())
	mux := newUserMux(service, newStackService())
	req := httptest.NewRequest(nethttp.MethodGet, "/users/me/stacks", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != nethttp.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, nethttp.StatusUnauthorized)
	}
}

func TestListCurrentUserStacksReturnsAllOwnedStacks(t *testing.T) {
	t.Parallel()

	userRepo := memory.NewRepository()
	stackService := newStackService()
	userService := application.NewUserService(userRepo)
	provisioner := application.NewUserProvisioner(userService, stackService)
	user, err := provisioner.Provision(context.Background(), "external-1", "one@example.com")
	if err != nil {
		t.Fatalf("CreateIfNotExist() error = %v, want nil", err)
	}

	defaultStacks, err := stackService.List(context.Background(), user.ExternalID)
	if err != nil {
		t.Fatalf("List() default stack error = %v", err)
	}
	if len(defaultStacks) != 1 {
		t.Fatalf("default stack count = %d, want 1", len(defaultStacks))
	}
	defaultStack := defaultStacks[0]
	publicStack := stackDomain.NewStack("Public Stack", user)
	publicStack.IsPublic = true
	if err := stackService.Create(context.Background(), *publicStack); err != nil {
		t.Fatalf("Create() public stack error = %v", err)
	}
	privateStack := stackDomain.NewStack("Private Stack", user)
	if err := stackService.Create(context.Background(), *privateStack); err != nil {
		t.Fatalf("Create() private stack error = %v", err)
	}
	otherUserStack := stackDomain.NewStack("Other User Stack", userDomain.User{ExternalID: "external-2"})
	otherUserStack.IsPublic = true
	if err := stackService.Create(context.Background(), *otherUserStack); err != nil {
		t.Fatalf("Create() other user stack error = %v", err)
	}

	mux := newUserMuxWithSession(userService, stackService, testSessionService{session: &commonauth.Session{
		UserID:  user.ExternalID,
		Email:   user.Email,
		IsValid: true,
	}})
	req := httptest.NewRequest(nethttp.MethodGet, "/users/me/stacks", nil)
	req.Header.Set("Authorization", "Bearer session-token")
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, nethttp.StatusOK)
	}

	var response []userHttp.StackResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response) != 3 {
		t.Fatalf("len(response) = %d, want %d", len(response), 3)
	}

	wantUUIDs := map[string]bool{
		defaultStack.UUID: true,
		publicStack.UUID:  true,
		privateStack.UUID: true,
	}
	for _, stack := range response {
		if !wantUUIDs[stack.UUID] {
			t.Fatalf("unexpected stack UUID %q", stack.UUID)
		}
		delete(wantUUIDs, stack.UUID)
	}
	if len(wantUUIDs) != 0 {
		t.Fatalf("missing stack UUIDs: %v", wantUUIDs)
	}
}

func TestListCurrentUserStacksReturnsUnauthorizedForInvalidToken(t *testing.T) {
	t.Parallel()

	stackService := newStackService()
	service := application.NewUserService(memory.NewRepository())
	mux := newUserMux(service, stackService)
	req := httptest.NewRequest(nethttp.MethodGet, "/users/me/stacks", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != nethttp.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, nethttp.StatusUnauthorized)
	}
}

func TestListUserStacksReturnsPublicStacks(t *testing.T) {
	t.Parallel()

	userRepo := memory.NewRepository()
	stackService := newStackService()
	userService := application.NewUserService(userRepo)
	user, err := userService.CreateIfNotExist(context.Background(), "external-1", "one@example.com")
	if err != nil {
		t.Fatalf("CreateIfNotExist() error = %v, want nil", err)
	}
	publicStack := stackDomain.NewStack("Public Stack", user)
	publicStack.IsPublic = true
	if err := stackService.Create(context.Background(), *publicStack); err != nil {
		t.Fatalf("Create() public stack error = %v", err)
	}
	privateStack := stackDomain.NewStack("Private Stack", user)
	if err := stackService.Create(context.Background(), *privateStack); err != nil {
		t.Fatalf("Create() private stack error = %v", err)
	}
	otherUserStack := stackDomain.NewStack("Other User Stack", userDomain.User{ExternalID: "external-2"})
	otherUserStack.IsPublic = true
	if err := stackService.Create(context.Background(), *otherUserStack); err != nil {
		t.Fatalf("Create() other user stack error = %v", err)
	}

	mux := newUserMux(userService, stackService)
	req := httptest.NewRequest(nethttp.MethodGet, "/users/external-1/stacks", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, nethttp.StatusOK)
	}

	var response []userHttp.StackResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response) != 1 {
		t.Fatalf("len(response) = %d, want %d", len(response), 1)
	}
	if response[0].UUID != publicStack.UUID {
		t.Fatalf("UUID = %q, want %q", response[0].UUID, publicStack.UUID)
	}
}

func TestListUserStacksReturnsNotFoundForMissingUser(t *testing.T) {
	t.Parallel()

	userService := application.NewUserService(memory.NewRepository())
	stackService := newStackService()
	mux := newUserMux(userService, stackService)
	req := httptest.NewRequest(nethttp.MethodGet, "/users/missing/stacks", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != nethttp.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.Code, nethttp.StatusNotFound)
	}
}

func newUserMux(service *application.UserService, stackService *stackApplication.StackService) *nethttp.ServeMux {
	return newUserMuxWithSession(service, stackService, testSessionService{})
}

func newUserMuxWithSession(
	service *application.UserService,
	stackService *stackApplication.StackService,
	sessionService commonauth.SessionService,
) *nethttp.ServeMux {
	mux := nethttp.NewServeMux()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	userHttp.AddUserRoute(mux, logger, service, stackService, sessionService)
	return mux
}

type testSessionService struct {
	session *commonauth.Session
}

func (s testSessionService) ResolveSession(context.Context, string) (*commonauth.Session, error) {
	return s.session, nil
}

func (testSessionService) LogoutSession(context.Context, string) error {
	return nil
}
