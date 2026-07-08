GOCMD=go
BINARY_DIR=.
BUILD_PKG=github.com/paperstacks.io/paperstacks/internal/common/build
VERSION ?= $(if $(APP_VERSION),$(APP_VERSION),$(shell git describe --tags 2>/dev/null || echo development))
GIT_HASH ?= $(if $(APP_GIT_HASH),$(APP_GIT_HASH),$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown))
BUILD_TIME ?= $(if $(APP_BUILD_TIME),$(APP_BUILD_TIME),$(shell date -u +"%Y-%m-%dT%H:%M:%SZ"))
BUILD_LDFLAGS=-ldflags "-X $(BUILD_PKG).Version=$(VERSION) -X $(BUILD_PKG).GitHash=$(GIT_HASH) -X $(BUILD_PKG).BuildTime=$(BUILD_TIME)"

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
	@GO_VERSION=$$(grep '^go ' go.mod | cut -d' ' -f2) && \
	CONTAINER_CMD=$$(if command -v docker >/dev/null 2>&1; then \
		printf 'docker'; \
	elif command -v podman >/dev/null 2>&1; then \
		printf 'podman'; \
	fi) && \
	test -n "$$CONTAINER_CMD" || { echo "Neither docker nor podman is installed"; exit 1; } && \
	echo "Using $$CONTAINER_CMD with Go version $$GO_VERSION from go.mod" && \
	$$CONTAINER_CMD build \
		--build-arg GO_VERSION=$$GO_VERSION \
		--build-arg APP_VERSION=$(VERSION) \
		--build-arg APP_GIT_HASH=$(GIT_HASH) \
		--build-arg APP_BUILD_TIME=$(BUILD_TIME) \
		-t paperstacks .

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
