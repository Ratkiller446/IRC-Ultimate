.PHONY: all build test test-race test-coverage test-property test-integration lint clean fmt vet security benchmark coverage-gate help

# Default target
all: build

# Application name and coverage files
APP_NAME := irc-client
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

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

# Run tests with coverage
test-coverage:
        @echo "Running tests with coverage..."
        go test -v -race -coverprofile=$(COVERAGE_FILE) ./...
        go tool cover -html=$(COVERAGE_FILE) -o=$(COVERAGE_HTML)
        @echo "Coverage report generated: $(COVERAGE_HTML)"
        go tool cover -func=$(COVERAGE_FILE)

# Run property-based tests
test-property:
        @echo "Running property-based tests..."
        go test -v -race -timeout=10m ./parser -run="Property"
        go test -v -race -timeout=10m ./commands -run="Property"

# Run integration tests
test-integration:
        @echo "Running integration tests..."
        go test -v -race -tags=integration ./...

# Run benchmarks
benchmark:
        @echo "Running benchmarks..."
        go test -bench=. -benchmem ./... | tee benchmark.txt

# Run security checks
security:
        @echo "Running security checks..."
        @if command -v gosec >/dev/null 2>&1; then \
                gosec ./...; \
        else \
                echo "gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
        fi

# Check coverage gate
coverage-gate: test-coverage
        @echo "Checking coverage requirements..."
        @COVERAGE=$$(go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print substr($$3, 1, length($$3)-1)}'); \
        echo "Current coverage: $${COVERAGE}%"; \
        REQUIRED=80.0; \
        if [ $$(echo "$${COVERAGE} < $${REQUIRED}" | bc -l 2>/dev/null || echo "0") -eq 1 ]; then \
                echo "❌ Coverage $${COVERAGE}% is below required $${REQUIRED}%"; \
                exit 1; \
        else \
                echo "✅ Coverage $${COVERAGE}% meets requirement of $${REQUIRED}%"; \
        fi

# Run comprehensive tests
test-all: test test-race test-property

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
        rm -f $(COVERAGE_FILE)
        rm -f $(COVERAGE_HTML)
        rm -f benchmark.txt
        go clean

# Run all quality checks (lint + test + test-race + coverage)
check: fmt lint test test-race coverage-gate
        @echo "All quality checks passed!"

# Development workflow
dev: clean fmt lint test

# CI workflow  
ci: clean fmt lint security test-all coverage-gate

# Go vet check
vet:
        @echo "Running go vet..."
        go vet ./...

# Help target
help:
        @echo "Available targets:"
        @echo "  build           - Build the application"
        @echo "  test            - Run tests"
        @echo "  test-race       - Run tests with race detection"
        @echo "  test-coverage   - Run tests with coverage report"
        @echo "  test-property   - Run property-based tests"
        @echo "  test-integration - Run integration tests"
        @echo "  test-all        - Run all test categories"
        @echo "  benchmark       - Run performance benchmarks"
        @echo "  lint            - Run linting checks (go vet + gofmt)"
        @echo "  fmt             - Format code with go fmt"
        @echo "  vet             - Run go vet"
        @echo "  security        - Run security checks"
        @echo "  coverage-gate   - Check coverage requirements"
        @echo "  clean           - Clean build artifacts"
        @echo "  check           - Run all quality checks"
        @echo "  dev             - Development workflow"
        @echo "  ci              - CI workflow"
        @echo "  help            - Show this help message"
        @echo "  all             - Default target (same as build)"