# Gemini CLI OpenAI Go - Makefile

# Variables
BINARY_NAME=gemini-cli-go
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME).exe
VERSION=1.0.0
BUILD_TIME=$(shell date +%Y-%m-%d_%H:%M:%S)
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Default target
.PHONY: all
all: clean build

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  build-all     - Build for all platforms"
	@echo "  run           - Run the application"
	@echo "  dev           - Run in development mode with hot reload"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download dependencies"
	@echo "  update        - Update dependencies"
	@echo "  lint          - Run linter"
	@echo "  format        - Format code"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  install       - Install the binary"
	@echo "  uninstall     - Uninstall the binary"

# Build the binary
.PHONY: build
build: deps
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Build for all platforms
.PHONY: build-all
build-all: clean deps
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o build/$(BINARY_NAME)_linux_amd64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o build/$(BINARY_NAME)_windows_amd64.exe .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o build/$(BINARY_NAME)_darwin_amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o build/$(BINARY_NAME)_darwin_arm64 .

# Run the application
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Run in development mode
.PHONY: dev
dev:
	@echo "Running in development mode..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Running without hot reload..."; \
		go run .; \
	fi

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Update dependencies
.PHONY: update
update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with:"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.54.2"; \
		echo "Running go vet instead..."; \
		go vet ./...; \
	fi

# Format code
.PHONY: format
format:
	@echo "Formatting code..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not found. Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(BINARY_WINDOWS)
	rm -rf build/
	rm -f coverage.out coverage.html

# Docker build
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest

# Docker run
.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run --rm -p 8080:8080 --env-file .env $(BINARY_NAME):latest

# Install the binary
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	@if [ "$(shell uname)" = "Darwin" ] || [ "$(shell uname)" = "Linux" ]; then \
		sudo cp $(BINARY_NAME) /usr/local/bin/; \
		echo "$(BINARY_NAME) installed to /usr/local/bin/"; \
	else \
		echo "Please manually copy $(BINARY_NAME) to your PATH"; \
	fi

# Uninstall the binary
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@if [ "$(shell uname)" = "Darwin" ] || [ "$(shell uname)" = "Linux" ]; then \
		sudo rm -f /usr/local/bin/$(BINARY_NAME); \
		echo "$(BINARY_NAME) uninstalled from /usr/local/bin/"; \
	else \
		echo "Please manually remove $(BINARY_NAME) from your PATH"; \
	fi

# Create release
.PHONY: release
release: clean build-all
	@echo "Creating release..."
	mkdir -p release
	cp build/* release/
	cp README.md release/
	cp .env.example release/
	@echo "Release created in release/ directory"

# Security scan
.PHONY: security
security:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Setup development environment
.PHONY: setup-dev
setup-dev:
	@echo "Setting up development environment..."
	go install github.com/cosmtrek/air@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "Development tools installed!"

# Check environment
.PHONY: check-env
check-env:
	@echo "Checking environment..."
	@if [ ! -f .env ]; then \
		echo "⚠️  .env file not found. Copy .env.example to .env and configure it."; \
	else \
		echo "✅ .env file found"; \
	fi
	@echo "Environment variables:"
	@echo "  GCP_SERVICE_ACCOUNT: $$(if [ -n "$$GCP_SERVICE_ACCOUNT" ]; then echo 'Set'; else echo 'Not set'; fi)"
	@echo "  GEMINI_PROJECT_ID: $$(if [ -n "$$GEMINI_PROJECT_ID" ]; then echo 'Set'; else echo 'Not set'; fi)"
	@echo "  OPENAI_API_KEY: $$(if [ -n "$$OPENAI_API_KEY" ]; then echo 'Set'; else echo 'Not set'; fi)"
	@echo "  PORT: $$(if [ -n "$$PORT" ]; then echo "$$PORT"; else echo '8080 (default)'; fi)"

# Quick start
.PHONY: quick-start
quick-start: check-env deps build
	@echo "🚀 Quick start completed!"
	@echo "To run the application:"
	@echo "  1. Configure your .env file with GCP_SERVICE_ACCOUNT"
	@echo "  2. Run: make run"
	@echo "  3. Visit: http://localhost:8080"

# Show version
.PHONY: version
version:
	@echo "$(BINARY_NAME) version $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Git commit: $(GIT_COMMIT)"