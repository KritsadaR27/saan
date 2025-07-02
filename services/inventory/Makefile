# Inventory Service Makefile

# Variables
BINARY_NAME=inventory-service
DOCKER_IMAGE=saan/inventory-service
VERSION=1.0.0

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(GOBIN)/$(BINARY_NAME) cmd/main.go

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	@go run cmd/main.go

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(GOBIN)
	@go clean

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Lint code
lint:
	@echo "Running linter..."
	@golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Build Docker image
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE):$(VERSION)..."
	@docker build -t $(DOCKER_IMAGE):$(VERSION) .
	@docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8082:8082 --name $(BINARY_NAME) $(DOCKER_IMAGE):$(VERSION)

# Stop Docker container
docker-stop:
	@echo "Stopping Docker container..."
	@docker stop $(BINARY_NAME) || true
	@docker rm $(BINARY_NAME) || true

# Push Docker image
docker-push:
	@echo "Pushing Docker image..."
	@docker push $(DOCKER_IMAGE):$(VERSION)
	@docker push $(DOCKER_IMAGE):latest

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@go mod download
	@go mod tidy

# Generate API documentation
docs:
	@echo "Generating API documentation..."
	@swag init -g cmd/main.go

# Run development server with hot reload
dev:
	@echo "Starting development server with hot reload..."
	@air

# Run database migrations
migrate-up:
	@echo "Running database migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" up

# Rollback database migrations
migrate-down:
	@echo "Rolling back database migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" down 1

# Create new migration
migrate-create:
	@echo "Creating new migration: $(name)"
	@migrate create -ext sql -dir migrations -seq $(name)

# Health check
health:
	@echo "Checking service health..."
	@curl -f http://localhost:8082/health || exit 1

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest

.PHONY: build run clean deps test test-coverage lint fmt docker-build docker-run docker-stop docker-push dev-setup docs dev migrate-up migrate-down migrate-create health install-tools
