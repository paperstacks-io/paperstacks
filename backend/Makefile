GOCMD=go
BINARY_DIR=.

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
	INTEGRATION=1 $(GOCMD) test -run '^TestIntegration' ./... -coverprofile=coverage.out

.PHONY: build
build: build-server ## Build all

.PHONY: build-server
build-server: ## Build server application
	mkdir -p $(BINARY_DIR)
	$(GOCMD) build -o $(BINARY_DIR)/server ./cmd/server/

.PHONY: build-docker
build-docker: ## Build Docker image for server
	@GO_VERSION=$$(grep '^go ' go.mod | cut -d' ' -f2) && \
	echo "Using Go version $$GO_VERSION from go.mod" && \
	docker build --build-arg GO_VERSION=$$GO_VERSION -t paperstacks .

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

