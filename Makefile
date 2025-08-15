# Makefile for fx-gin-scaffold
.PHONY: all build clean run test lint swagger help dev deps

# Variables
APP_NAME=fx-gin-scaffold
BUILD_DIR=./bin
MAIN_FILE=./cmd/server/main.go
GOPATH=$(shell go env GOPATH)

# Default target
all: clean lint test build

## Development Commands

dev: ## Run the application in development mode with hot reload
	@echo "Starting development server..."
	@go run $(MAIN_FILE)

deps: ## Download and install dependencies
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

build: ## Build the application
	@echo "Building application..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "Build completed: $(BUILD_DIR)/$(APP_NAME)"

run: build ## Build and run the application
	@echo "Running application..."
	@$(BUILD_DIR)/$(APP_NAME)

## Testing Commands

test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-repo: ## Run repository tests only
	@echo "Running repository tests..."
	@go test -v ./internal/repo/...

## Code Quality Commands

lint: ## Run code linting
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

## Documentation Commands

swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@mkdir -p docs/swagger
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g ./cmd/server/main.go -o ./docs/swagger; \
	elif [ -f "$(GOPATH)/bin/swag" ]; then \
		$(GOPATH)/bin/swag init -g ./cmd/server/main.go -o ./docs/swagger; \
	else \
		echo "swag not installed. Install it with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

## Database Commands

migrate: ## Run migrations only (production use)
	@echo "Running database migrations..."
	@go run ./cmd/migrate/main.go

check-migrations: ## Check pending migrations
	@echo "Checking pending migrations..."
	@go run ./cmd/migrate/main.go -check

migrate-dry-run: ## Show what migrations would be executed
	@echo "Showing pending migrations..."
	@go run ./cmd/migrate/main.go -dry-run

## Utility Commands

clean: ## Clean build files and caches
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@rm -f coverage.out coverage.html

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install golang.org/x/tools/cmd/goimports@latest

setup: deps install-tools ## Setup development environment
	@echo "Setting up development environment..."
	@cp .env.example .env
	@mkdir -p data
	@mkdir -p logs
	@echo "Development environment setup complete!"
	@echo "1. Edit .env file with your configuration"
	@echo "2. Run 'make dev' to start development server (migrations run automatically)"

help: ## Show this help message
	@echo "Usage:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)