GOCMD=go
BINARY_DIR=.
BUILD_PKG=github.com/paperstacks.io/paperstacks/internal/common/build
GO_VERSION ?= $(shell awk '/^go / { print $$2 }' go.mod)
VERSION ?= $(if $(APP_VERSION),$(APP_VERSION),$(shell git describe --tags --abbrev=0 2>/dev/null || echo development))
GIT_HASH ?= $(if $(APP_GIT_HASH),$(APP_GIT_HASH),$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown))
BUILD_LDFLAGS=-ldflags "-X $(BUILD_PKG).Version=$(VERSION) -X $(BUILD_PKG).GitHash=$(GIT_HASH)"

.PHONY: deps
deps: ## Install all Go dependencies
	$(GOCMD) mod tidy

.PHONY: clean
clean: ## Clean up all build artifacts
	$(GOCMD) clean
	rm -f $(BINARY_DIR)/server $(BINARY_DIR)/server.exe

.PHONY: test
test: ## Run all unit tests
	$(GOCMD) test ./... -coverprofile=coverage.out

.PHONY: test-integration
test-integration: ## Run all integration tests
	INTEGRATION=1 $(GOCMD) test -parallel=1 -run '^TestIntegration' ./... -coverprofile=coverage.out

.PHONY: build
build: build-server ## Build all

.PHONY: build-server
build-server: ## Build server application
	mkdir -p $(BINARY_DIR)
	$(GOCMD) build $(BUILD_LDFLAGS) -o $(BINARY_DIR)/server ./cmd/server/

.PHONY: build-docker
build-docker: ## Build Docker image for server
	@COMPOSE_CMD=$$(if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then \
		printf 'docker compose'; \
	elif command -v podman >/dev/null 2>&1 && podman compose version >/dev/null 2>&1; then \
		printf 'podman compose'; \
	fi) && \
	test -n "$$COMPOSE_CMD" || { echo "Neither docker compose nor podman compose is installed"; exit 1; } && \
	echo "Using $$COMPOSE_CMD with Go version $(GO_VERSION) from go.mod" && \
	GO_VERSION=$(GO_VERSION) \
	APP_VERSION=$(VERSION) \
	APP_GIT_HASH=$(GIT_HASH) \
	$$COMPOSE_CMD build paperstacks

.PHONY: build-github-output
build-github-output: ## Print build values as GitHub Actions outputs
	@test -n '$(GO_VERSION)' || { printf '%s\n' 'GO_VERSION is empty' >&2; exit 1; }
	@test -n '$(VERSION)' || { printf '%s\n' 'VERSION is empty' >&2; exit 1; }
	@test -n '$(GIT_HASH)' || { printf '%s\n' 'GIT_HASH is empty' >&2; exit 1; }
	@printf 'go-version=%s\n' '$(GO_VERSION)'
	@printf 'app-version=%s\n' '$(VERSION)'
	@printf 'app-git-hash=%s\n' '$(GIT_HASH)'

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
