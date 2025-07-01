# Loyverse Webhook Service Makefile

.PHONY: build run test clean docker-build docker-run

# Variables
BINARY_NAME=loyverse-webhook
DOCKER_IMAGE=saan/loyverse-webhook
DOCKER_TAG=latest

# Build the application
build:
	go build -o bin/$(BINARY_NAME) ./cmd/main.go

# Run the application locally (for development only)
run: build
	@echo "⚠️  Running locally - Remember to use docker-compose in production"
	./bin/$(BINARY_NAME)

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

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Download dependencies
deps:
	go mod download
	go mod verify

# Tidy dependencies
tidy:
	go mod tidy

# Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Run Docker container (for testing only)
docker-run: docker-build
	@echo "⚠️  Running single container - Use docker-compose for full system"
	docker run -p 8093:8093 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

# Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Generate go.sum
go-sum:
	go mod download
	go mod verify

# Help
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application locally (dev only)"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  deps          - Download dependencies"
	@echo "  tidy          - Tidy dependencies"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container (testing only)"
	@echo "  install-tools - Install development tools"
	@echo "  help          - Show this help"
