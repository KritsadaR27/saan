# Saan System - Development Commands

.PHONY: help dev build test lint deploy clean

# Default target
help: ## Show this help message
	@echo "Saan System - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
dev: ## Start all services in development mode
	@echo "🚀 Starting Saan System in development mode..."
	docker-compose up -d
	@echo "✅ All services started!"
	@echo "🌐 Web App: http://localhost:3000"
	@echo "🔧 Admin: http://localhost:3001"
	@echo "💬 Chat API: http://localhost:8001"

dev-logs: ## Show logs for all development services
	docker-compose logs -f

dev-stop: ## Stop all development services
	docker-compose down

# Individual service development
chat-dev: ## Start chat service in development mode
	@echo "💬 Starting Chat Service..."
	cd services/chat && go run cmd/main.go

order-dev: ## Start order service in development mode
	@echo "📋 Starting Order Service..."
	cd services/order && go run cmd/main.go

inventory-dev: ## Start inventory service in development mode
	@echo "📦 Starting Inventory Service..."
	cd services/inventory && go run cmd/main.go

web-dev: ## Start web frontend in development mode
	@echo "🌐 Starting Web App..."
	cd apps/web && npm run dev

admin-dev: ## Start admin frontend in development mode
	@echo "🔧 Starting Admin Dashboard..."
	cd apps/admin && npm run dev

# Build
build: ## Build all services and applications
	@echo "🔨 Building all services..."
	docker-compose build
	@echo "✅ Build completed!"

build-services: ## Build all Go services
	@echo "🔨 Building Go services..."
	cd services/chat && go build -o bin/chat cmd/main.go
	cd services/order && go build -o bin/order cmd/main.go
	cd services/inventory && go build -o bin/inventory cmd/main.go
	cd services/delivery && go build -o bin/delivery cmd/main.go
	cd services/finance && go build -o bin/finance cmd/main.go
	@echo "✅ Go services built!"

build-apps: ## Build all frontend applications
	@echo "🔨 Building frontend apps..."
	cd apps/web && npm run build
	cd apps/admin && npm run build
	@echo "✅ Frontend apps built!"

# Testing
test: ## Run all tests
	@echo "🧪 Running all tests..."
	$(MAKE) test-services
	$(MAKE) test-apps
	@echo "✅ All tests completed!"

test-services: ## Run Go service tests
	@echo "🧪 Testing Go services..."
	cd services/chat && go test ./...
	cd services/order && go test ./...
	cd services/inventory && go test ./...
	cd services/delivery && go test ./...
	cd services/finance && go test ./...

test-apps: ## Run frontend tests
	@echo "🧪 Testing frontend apps..."
	cd apps/web && npm test
	cd apps/admin && npm test

# Code Quality
lint: ## Run linting for all code
	@echo "🔍 Running linters..."
	$(MAKE) lint-services
	$(MAKE) lint-apps
	@echo "✅ Linting completed!"

lint-services: ## Lint Go services
	@echo "🔍 Linting Go services..."
	cd services/chat && golangci-lint run
	cd services/order && golangci-lint run
	cd services/inventory && golangci-lint run
	cd services/delivery && golangci-lint run
	cd services/finance && golangci-lint run

lint-apps: ## Lint frontend applications
	@echo "🔍 Linting frontend apps..."
	cd apps/web && npm run lint
	cd apps/admin && npm run lint

format: ## Format all code
	@echo "💅 Formatting code..."
	cd services/chat && go fmt ./...
	cd services/order && go fmt ./...
	cd services/inventory && go fmt ./...
	cd services/delivery && go fmt ./...
	cd services/finance && go fmt ./...
	cd apps/web && npm run format
	cd apps/admin && npm run format

# Database
db-migrate: ## Run database migrations
	@echo "🗄️ Running database migrations..."
	docker-compose exec postgres psql -U saan -d saan_db -f /migrations/init.sql

db-seed: ## Seed database with sample data
	@echo "🌱 Seeding database..."
	docker-compose exec postgres psql -U saan -d saan_db -f /seeds/sample_data.sql

db-reset: ## Reset database (WARNING: destroys all data)
	@echo "⚠️ Resetting database..."
	docker-compose down -v
	docker-compose up -d postgres
	sleep 5
	$(MAKE) db-migrate
	$(MAKE) db-seed

# Dependencies
deps: ## Install all dependencies
	@echo "📦 Installing dependencies..."
	$(MAKE) deps-services
	$(MAKE) deps-apps
	@echo "✅ Dependencies installed!"

deps-services: ## Install Go dependencies
	@echo "📦 Installing Go dependencies..."
	cd services/chat && go mod download
	cd services/order && go mod download
	cd services/inventory && go mod download
	cd services/delivery && go mod download
	cd services/finance && go mod download

deps-apps: ## Install Node.js dependencies
	@echo "📦 Installing Node.js dependencies..."
	cd apps/web && npm install
	cd apps/admin && npm install
	cd packages/ui && npm install
	cd packages/types && npm install
	cd packages/utils && npm install

# Deployment
deploy: ## Deploy to production
	@echo "🚀 Deploying to production..."
	docker-compose -f docker-compose.prod.yml up -d
	@echo "✅ Deployment completed!"

deploy-k8s: ## Deploy to Kubernetes
	@echo "🚀 Deploying to Kubernetes..."
	kubectl apply -f infrastructure/k8s/
	@echo "✅ Kubernetes deployment completed!"

# Monitoring
logs: ## Show logs from all services
	docker-compose logs -f

logs-chat: ## Show logs from chat service
	docker-compose logs -f chat

logs-order: ## Show logs from order service
	docker-compose logs -f order

logs-web: ## Show logs from web app
	docker-compose logs -f web

# Cleanup
clean: ## Clean up build artifacts and containers
	@echo "🧹 Cleaning up..."
	docker-compose down
	docker system prune -f
	cd services/chat && go clean
	cd services/order && go clean
	cd services/inventory && go clean
	cd services/delivery && go clean
	cd services/finance && go clean
	@echo "✅ Cleanup completed!"

# Utilities
status: ## Show status of all services
	@echo "📊 Saan System Status:"
	docker-compose ps

health: ## Check health of all services
	@echo "🏥 Health Check:"
	@curl -s http://localhost:8001/health || echo "❌ Chat service down"
	@curl -s http://localhost:8002/health || echo "❌ Order service down"
	@curl -s http://localhost:8003/health || echo "❌ Inventory service down"
	@curl -s http://localhost:3000/api/health || echo "❌ Web app down"
	@echo "✅ Health check completed!"

# Development tools
shell-chat: ## Open shell in chat service container
	docker-compose exec chat sh

shell-postgres: ## Open PostgreSQL shell
	docker-compose exec postgres psql -U saan -d saan_db

shell-redis: ## Open Redis shell
	docker-compose exec redis redis-cli
