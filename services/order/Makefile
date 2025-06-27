# Order Service Makefile

# Variables
SERVICE_NAME=order-service
BINARY_NAME=bin/$(SERVICE_NAME)
DOCKER_IMAGE=saan/$(SERVICE_NAME)
DOCKER_TAG=latest

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Directories
CMD_DIR=./cmd
PKG_DIR=./...

.PHONY: help build clean test run docker-build docker-run docker-push deps lint format

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build the application
build: ## Build the application
	@echo "Building $(SERVICE_NAME)..."
	@mkdir -p bin
	@$(GOBUILD) -o $(BINARY_NAME) $(CMD_DIR)
	@echo "Build complete: $(BINARY_NAME)"

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf bin/
	@echo "Clean complete"

# Run tests
test: ## Run tests
	@echo "Running tests..."
	@$(GOTEST) -v $(PKG_DIR)

# Run tests with coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@$(GOTEST) -v -coverprofile=coverage.out $(PKG_DIR)
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run the application
run: build ## Build and run the application
	@echo "Running $(SERVICE_NAME)..."
	@./$(BINARY_NAME)

# Install/update dependencies
deps: ## Install/update dependencies
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy

# Lint the code
lint: ## Lint the code
	@echo "Linting..."
	@golangci-lint run

# Format the code
format: ## Format the code
	@echo "Formatting..."
	@gofmt -s -w .
	@$(GOCMD) mod tidy

# Build Docker image
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# Run Docker container
docker-run: docker-build ## Build and run Docker container
	@echo "Running Docker container..."
	@docker run --rm -p 8080:8080 \
		-e DB_HOST=localhost \
		-e DB_PORT=5432 \
		-e DB_USER=postgres \
		-e DB_PASSWORD=password \
		-e DB_NAME=order_db \
		-e DB_SSLMODE=disable \
		-e SERVER_HOST=0.0.0.0 \
		-e SERVER_PORT=8080 \
		-e LOG_LEVEL=info \
		-e LOG_FORMAT=json \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Push Docker image
docker-push: docker-build ## Build and push Docker image
	@echo "Pushing Docker image..."
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

# Development setup
dev-setup: deps ## Setup development environment
	@echo "Setting up development environment..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development setup complete"

# Run migrations (requires database connection)
migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	@psql -h ${DB_HOST:-localhost} -p ${DB_PORT:-5432} -U ${DB_USER:-postgres} -d ${DB_NAME:-order_db} -f migrations/001_create_orders.sql

# Create migration file
migrate-create: ## Create a new migration file (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then echo "Error: NAME is required. Usage: make migrate-create NAME=migration_name"; exit 1; fi
	@echo "Creating migration: $(NAME)"
	@touch migrations/$(shell date +%Y%m%d%H%M%S)_$(NAME).sql
	@echo "Migration file created"

# Hot reload for development (requires air)
dev: ## Run with hot reload (requires 'air' to be installed)
	@air

# Install development tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development tools installed"

# Build for multiple platforms
build-all: ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	@mkdir -p bin
	@GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(SERVICE_NAME)-linux-amd64 $(CMD_DIR)
	@GOOS=windows GOARCH=amd64 $(GOBUILD) -o bin/$(SERVICE_NAME)-windows-amd64.exe $(CMD_DIR)
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) -o bin/$(SERVICE_NAME)-darwin-amd64 $(CMD_DIR)
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) -o bin/$(SERVICE_NAME)-darwin-arm64 $(CMD_DIR)
	@echo "Multi-platform build complete"

# Check for security vulnerabilities
security-check: ## Check for security vulnerabilities
	@echo "Checking for security vulnerabilities..."
	@govulncheck ./...
