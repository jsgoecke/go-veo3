.PHONY: build test lint security clean install run coverage pre-build check all help

BINARY_NAME=veo3
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GOFLAGS=-ldflags "-X github.com/jasongoecke/go-veo3/pkg/cli.Version=$(VERSION) -X github.com/jasongoecke/go-veo3/pkg/cli.BuildTime=$(BUILD_TIME)"

# Build platforms
PLATFORMS=linux darwin windows
ARCHITECTURES=amd64 arm64

help: ## Show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

all: check build ## Run all checks and build

check: lint security test ## Run all quality checks (lint, security, test)

pre-build: lint test-unit ## Run pre-build checks (lint and unit tests only)

build: pre-build ## Build the binary (runs lint and unit tests first)
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	go build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/veo3
	@echo "Build complete: $(BINARY_NAME)"

build-all: pre-build ## Build for all platforms and architectures
	@echo "Building for all platforms..."
	@mkdir -p dist
	@for os in $(PLATFORMS); do \
		for arch in $(ARCHITECTURES); do \
			output="dist/$(BINARY_NAME)-$$os-$$arch"; \
			if [ "$$os" = "windows" ]; then output="$$output.exe"; fi; \
			echo "Building $$output..."; \
			GOOS=$$os GOARCH=$$arch go build $(GOFLAGS) -o $$output ./cmd/veo3 || true; \
		done \
	done
	@echo "Multi-platform build complete. Artifacts in dist/"

run: ## Run the application
	go run ./cmd/veo3

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v ./tests/unit/...

test: ## Run all tests (unit + integration)
	@echo "Running all tests..."
	go test -v ./...

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	go test -v -race ./...

test-coverage: ## Run tests with coverage report (requires 80% minimum)
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out -covermode=atomic ./...
	@echo "\nCoverage summary:"
	@go tool cover -func=coverage.out | grep total
	@echo "\nChecking 80% coverage threshold..."
	@go tool cover -func=coverage.out | grep total | awk '{if ($$3+0 < 80.0) {print "Coverage below 80% threshold: " $$3; exit 1} else {print "Coverage meets 80% threshold: " $$3}}'

coverage: test-coverage ## Alias for test-coverage
	@go tool cover -html=coverage.out -o coverage.html
	@echo "HTML coverage report generated: coverage.html"

lint: ## Run linters
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install from https://golangci-lint.run/"; exit 1)
	golangci-lint run --timeout 5m

security: ## Run security scans
	@echo "Running security scans..."
	@which gosec > /dev/null || (echo "Installing gosec..."; go install github.com/securego/gosec/v2/cmd/gosec@latest)
	gosec -fmt=text -out=security-report.txt ./... || true
	@echo "Security scan complete. Report: security-report.txt"

fmt: ## Format code
	@echo "Formatting code..."
	gofmt -w -s .
	goimports -w .

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	go mod tidy

clean: ## Clean build artifacts
	@echo "Cleaning..."
	go clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -f security-report.txt
	rm -rf dist/

install: build ## Install the binary
	go install $(GOFLAGS) ./cmd/veo3

.DEFAULT_GOAL := help
