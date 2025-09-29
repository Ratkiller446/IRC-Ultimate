.PHONY: all build test test-race lint clean help

# Default target
all: build

# Application name
APP_NAME := irc-client

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	go build -o $(APP_NAME) .

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -race -timeout 30s ./...

# Lint the code (includes go vet, gofmt check)
lint:
	@echo "Running linting checks..."
	@echo "Running go vet..."
	go vet ./...
	@echo "Checking code formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "Code is not formatted properly. Run 'go fmt ./...' to fix:"; \
		gofmt -l .; \
		exit 1; \
	fi
	@echo "All linting checks passed!"

# Format the code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(APP_NAME)
	go clean

# Run all quality checks (lint + test + test-race)
check: lint test test-race
	@echo "All quality checks passed!"

# Help target
help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  test       - Run tests"
	@echo "  test-race  - Run tests with race detection"
	@echo "  lint       - Run linting checks (go vet + gofmt)"
	@echo "  fmt        - Format code with go fmt"
	@echo "  clean      - Clean build artifacts"
	@echo "  check      - Run all quality checks (lint + test + test-race)"
	@echo "  help       - Show this help message"
	@echo "  all        - Default target (same as build)"