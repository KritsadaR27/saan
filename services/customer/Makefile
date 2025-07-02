.PHONY: build run test clean docker-build docker-run migrate-up migrate-down deps

# Variables
SERVICE_NAME=customer-service
DOCKER_IMAGE=saan/customer-service
GO_FILES=$(shell find . -name '*.go' -type f)

# Default target
all: build

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build the application
build:
	go build -o bin/$(SERVICE_NAME) cmd/main.go

# Run the application
run:
	go run cmd/main.go

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Test coverage
coverage: test
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker build
docker-build:
	docker build -t $(DOCKER_IMAGE):latest .

# Docker run
docker-run: docker-build
	docker run -p 8110:8110 --env-file .env $(DOCKER_IMAGE):latest

# Database migrations up
migrate-up:
	migrate -path migrations -database "postgres://saan_user:saan_password@localhost:5432/saan_customer?sslmode=disable" up

# Database migrations down
migrate-down:
	migrate -path migrations -database "postgres://saan_user:saan_password@localhost:5432/saan_customer?sslmode=disable" down

# Create new migration
migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

# Lint code
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Install tools
install-tools:
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Development setup
dev-setup: deps install-tools
	cp .env.example .env
	@echo "Development environment setup complete"
	@echo "Please configure .env file with your settings"

# Hot reload for development
dev:
	air -c .air.toml
