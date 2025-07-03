# Product Service Makefile

.PHONY: build run test clean install deps migrate-up migrate-down docker-build docker-run

# Variables
SERVICE_NAME := product-service
DOCKER_IMAGE := $(SERVICE_NAME):latest
POSTGRES_URL := postgres://postgres:password@localhost:5432/saan_products?sslmode=disable
MIGRATION_DIR := ./migrations

# Build the service
build:
	go build -o bin/$(SERVICE_NAME) cmd/main.go

# Run the service
run:
	go run cmd/main.go

# Install dependencies
install:
	go mod download
	go mod tidy

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install development dependencies
deps:
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Database migrations
migrate-up:
	migrate -path $(MIGRATION_DIR) -database "$(POSTGRES_URL)" up

migrate-down:
	migrate -path $(MIGRATION_DIR) -database "$(POSTGRES_URL)" down

migrate-create:
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME)

# Database operations
db-reset:
	migrate -path $(MIGRATION_DIR) -database "$(POSTGRES_URL)" down -all
	migrate -path $(MIGRATION_DIR) -database "$(POSTGRES_URL)" up

db-seed:
	# TODO: Add seed data script
	@echo "Seeding database..."

# Linting
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...
	go mod tidy

# Docker operations
docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	docker run -p 8083:8083 --env-file .env $(DOCKER_IMAGE)

docker-compose-up:
	docker-compose up -d

docker-compose-down:
	docker-compose down

# Development
dev:
	go run cmd/main.go

# Production build
prod-build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/$(SERVICE_NAME) cmd/main.go

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the service"
	@echo "  run            - Run the service"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  clean          - Clean build artifacts"
	@echo "  install        - Install dependencies"
	@echo "  deps           - Install development dependencies"
	@echo "  migrate-up     - Run database migrations up"
	@echo "  migrate-down   - Run database migrations down"
	@echo "  migrate-create - Create a new migration (use NAME=name)"
	@echo "  db-reset       - Reset database (down all, then up)"
	@echo "  db-seed        - Seed database with sample data"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  dev            - Run in development mode"
	@echo "  prod-build     - Build for production"
	@echo "  help           - Show this help message"
