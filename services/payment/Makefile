# Makefile for Payment Service
# SAAN System - Payment Service with Loyverse Integration

.PHONY: help build run test clean lint docker-build docker-run migrate-up migrate-down deps

# Variables
SERVICE_NAME := payment-service
BINARY_NAME := payment-service
DOCKER_IMAGE := saan/payment-service
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development commands
deps: ## Install dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

build: ## Build the service binary
	@echo "Building $(SERVICE_NAME)..."
	CGO_ENABLED=0 GOOS=linux go build -ldflags "$(LDFLAGS)" -a -installsuffix cgo -o bin/$(BINARY_NAME) cmd/main.go

run: ## Run the service locally
	@echo "Running $(SERVICE_NAME)..."
	go run cmd/main.go

run-dev: ## Run the service in development mode with auto-reload
	@echo "Running $(SERVICE_NAME) in development mode..."
	air -c .air.toml

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linters
	@echo "Running linters..."
	golangci-lint run ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Database commands
migrate-up: ## Run database migrations up
	@echo "Running database migrations up..."
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up

migrate-down: ## Run database migrations down
	@echo "Running database migrations down..."
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" down

migrate-create: ## Create new migration file (usage: make migrate-create NAME=migration_name)
	@echo "Creating new migration: $(NAME)"
	migrate create -ext sql -dir migrations -seq $(NAME)

db-reset: ## Reset database (drop and recreate)
	@echo "Resetting database..."
	make migrate-down
	make migrate-up

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run --rm -p 8087:8087 \
		-e DB_HOST=host.docker.internal \
		-e DB_PORT=5432 \
		-e DB_USER=saan_user \
		-e DB_PASSWORD=saan_password \
		-e DB_NAME=saan_payment \
		-e REDIS_HOST=host.docker.internal \
		-e REDIS_PORT=6379 \
		$(DOCKER_IMAGE):latest

docker-push: ## Push Docker image to registry
	@echo "Pushing Docker image..."
	docker push $(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_IMAGE):latest

# Docker Compose commands
compose-up: ## Start services with docker-compose
	@echo "Starting services with docker-compose..."
	docker-compose up -d

compose-down: ## Stop services with docker-compose
	@echo "Stopping services with docker-compose..."
	docker-compose down

compose-logs: ## View docker-compose logs
	@echo "Viewing docker-compose logs..."
	docker-compose logs -f

# Development tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

gen-docs: ## Generate API documentation
	@echo "Generating API documentation..."
	swag init -g cmd/main.go -o docs/

# Load testing
load-test: ## Run load tests
	@echo "Running load tests..."
	k6 run tests/load/payment_load_test.js

# Code quality
fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

security-scan: ## Run security scan
	@echo "Running security scan..."
	gosec ./...

# Environment setup
setup-dev: ## Setup development environment
	@echo "Setting up development environment..."
	make install-tools
	make deps
	cp .env.example .env
	@echo "Please update .env file with your configuration"

# Service-specific commands
seed-data: ## Seed test data
	@echo "Seeding test data..."
	go run scripts/seed_data.go

test-endpoints: ## Test API endpoints
	@echo "Testing API endpoints..."
	@curl -X GET http://localhost:8087/health
	@echo ""
	@curl -X GET http://localhost:8087/api/v1/payments

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Production commands
deploy-staging: ## Deploy to staging environment
	@echo "Deploying to staging..."
	# Add deployment commands here

deploy-prod: ## Deploy to production environment
	@echo "Deploying to production..."
	# Add deployment commands here

# Monitoring
logs: ## View service logs
	@echo "Viewing service logs..."
	docker logs -f $(SERVICE_NAME) 2>/dev/null || echo "Service not running in Docker"

health-check: ## Check service health
	@echo "Checking service health..."
	@curl -s http://localhost:8087/health | jq .

# Integration with other services
test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go test -tags=integration ./tests/integration/...

# Payment-specific commands
test-loyverse: ## Test Loyverse integration
	@echo "Testing Loyverse integration..."
	go run scripts/test_loyverse.go

sync-stores: ## Sync Loyverse stores
	@echo "Syncing Loyverse stores..."
	go run scripts/sync_stores.go

process-cod: ## Process COD payments
	@echo "Processing COD payments..."
	go run scripts/process_cod.go

# Data retrieval testing (3 types)
test-store-data: ## Test store-based data retrieval (Type 1)
	@echo "Testing store-based data retrieval..."
	@curl -X GET "http://localhost:8087/api/v1/stores/STORE001/payments?limit=10" | jq .

test-customer-data: ## Test customer-based data retrieval (Type 2)
	@echo "Testing customer-based data retrieval..."
	@curl -X GET "http://localhost:8087/api/v1/customers/$(CUSTOMER_ID)/payments?limit=10" | jq .

test-order-data: ## Test order-based data retrieval (Type 3)
	@echo "Testing order-based data retrieval..."
	@curl -X GET "http://localhost:8087/api/v1/orders/$(ORDER_ID)/payment-summary" | jq .

# Environment variables for testing
export DB_HOST ?= localhost
export DB_PORT ?= 5432
export DB_USER ?= saan_user
export DB_PASSWORD ?= saan_password
export DB_NAME ?= saan_payment
export DB_SSLMODE ?= disable
export REDIS_HOST ?= localhost
export REDIS_PORT ?= 6379
export CUSTOMER_ID ?= 123e4567-e89b-12d3-a456-426614174000
export ORDER_ID ?= 123e4567-e89b-12d3-a456-426614174001
