.PHONY: help build test test-unit test-integration test-race lint vet fmt clean run dev air install-tools docker-test release release-linux release-darwin release-all

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=phanes
BUILD_DIR=tmp
RELEASE_DIR=dist
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION)"

# Supported platforms for release builds
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

run: build ## Build and run the binary
	@./$(BUILD_DIR)/$(BINARY_NAME)

dev: ## Run with air for hot reload (requires air to be installed)
	@if ! command -v air > /dev/null; then \
		echo "Error: air is not installed. Run 'make install-tools' to install it."; \
		exit 1; \
	fi
	@air

air: dev ## Alias for dev target

test: ## Run all tests
	@echo "Running all tests..."
	@go test -v ./...

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@go test -v ./internal/...

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	@if [ -d "test/integration" ]; then \
		go test -v ./test/integration/...; \
	else \
		echo "No integration tests found in test/integration/"; \
	fi

test-e2e: ## Run E2E tests in Docker container
	@echo "Running E2E tests in Docker container..."
	@docker-compose -f docker-compose.test.yml build
	@docker-compose -f docker-compose.test.yml run --rm phanes-test go test -v ./test/integration/... -run "E2E"

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	@go test -race -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -coverprofile=$(COVERAGE_FILE) ./...
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

lint: ## Run golangci-lint (if available)
	@if command -v golangci-lint > /dev/null; then \
		echo "Running golangci-lint..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Skipping lint check."; \
	fi

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

fmt: ## Format code with gofmt
	@echo "Formatting code..."
	@go fmt ./...

check: fmt vet lint test ## Run all checks (format, vet, lint, test)

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@rm -f build-errors.log
	@go clean

install-tools: ## Install development tools (air, golangci-lint)
	@echo "Installing development tools..."
	@if ! command -v air > /dev/null; then \
		echo "Installing air..."; \
		go install github.com/air-verse/air@latest; \
	else \
		echo "air is already installed"; \
	fi
	@if ! command -v golangci-lint > /dev/null; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin latest; \
	else \
		echo "golangci-lint is already installed"; \
	fi

docker-test: ## Run tests in Docker containers
	@./scripts/dev.sh test

docker-build: ## Build Docker containers for testing
	@./scripts/dev.sh build

docker-up: ## Start Docker containers
	@./scripts/dev.sh up

docker-down: ## Stop Docker containers
	@./scripts/dev.sh down

docker-logs: ## Show Docker container logs
	@./scripts/dev.sh logs

# Release targets
release-clean: ## Clean release artifacts
	@echo "Cleaning release artifacts..."
	@rm -rf $(RELEASE_DIR)

release-dir: ## Create release directory
	@mkdir -p $(RELEASE_DIR)

release-linux-amd64: release-dir ## Build for Linux amd64
	@echo "Building for linux/amd64..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_linux_amd64/$(BINARY_NAME) .
	@cd $(RELEASE_DIR) && tar -czvf $(BINARY_NAME)_$(VERSION)_linux_amd64.tar.gz $(BINARY_NAME)_$(VERSION)_linux_amd64
	@rm -rf $(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_linux_amd64

release-linux-arm64: release-dir ## Build for Linux arm64
	@echo "Building for linux/arm64..."
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_linux_arm64/$(BINARY_NAME) .
	@cd $(RELEASE_DIR) && tar -czvf $(BINARY_NAME)_$(VERSION)_linux_arm64.tar.gz $(BINARY_NAME)_$(VERSION)_linux_arm64
	@rm -rf $(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_linux_arm64

release-darwin-amd64: release-dir ## Build for macOS amd64
	@echo "Building for darwin/amd64..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_amd64/$(BINARY_NAME) .
	@cd $(RELEASE_DIR) && tar -czvf $(BINARY_NAME)_$(VERSION)_darwin_amd64.tar.gz $(BINARY_NAME)_$(VERSION)_darwin_amd64
	@rm -rf $(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_amd64

release-darwin-arm64: release-dir ## Build for macOS arm64 (Apple Silicon)
	@echo "Building for darwin/arm64..."
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_arm64/$(BINARY_NAME) .
	@cd $(RELEASE_DIR) && tar -czvf $(BINARY_NAME)_$(VERSION)_darwin_arm64.tar.gz $(BINARY_NAME)_$(VERSION)_darwin_arm64
	@rm -rf $(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_arm64

release: release-clean release-linux-amd64 release-linux-arm64 release-darwin-amd64 release-darwin-arm64 ## Build release binaries for all platforms
	@echo ""
	@echo "Release builds complete. Artifacts in $(RELEASE_DIR)/"
	@ls -la $(RELEASE_DIR)/

release-checksums: release ## Generate checksums for release artifacts
	@echo "Generating checksums..."
	@cd $(RELEASE_DIR) && sha256sum *.tar.gz > checksums.txt
	@cat $(RELEASE_DIR)/checksums.txt

