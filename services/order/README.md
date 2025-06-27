# Order Service

A microservice for managing orders in the Saan System, built with Go and following Clean Architecture principles.

## Architecture

This service follows Clean Architecture with the following layers:

```
services/order/
├── cmd/                        # Entry point: main.go
│   └── main.go
├── internal/
│   ├── domain/                 # Domain Model + Interface
│   │   ├── order.go           # Order and OrderItem entities
│   │   ├── repository.go      # Repository interfaces
│   │   └── errors.go          # Domain errors
│   ├── application/           # Business Logic / Use cases
│   │   ├── order_service.go   # Order service implementation
│   │   └── dto/
│   │       └── order_dto.go   # Data transfer objects
│   ├── transport/             # Input Adapters
│   │   ├── http/
│   │   │   ├── handler.go     # HTTP handlers
│   │   │   ├── routes.go      # Route definitions
│   │   │   └── middleware/    # HTTP middleware
│   │   └── grpc/              # gRPC transport (future)
│   └── infrastructure/        # Output Adapters
│       ├── repository/
│       │   └── postgres_order_repository.go
│       ├── config/
│       │   └── loader.go      # Configuration loader
│       └── db/
│           └── postgres.go    # Database connection
├── migrations/                # Database migrations
│   └── 001_create_orders.sql
├── pkg/                       # Shared utilities
│   └── logger/
│       └── logger.go
├── Dockerfile
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

## Features

- **Order Management**: Create, read, update, delete orders
- **Order Status Tracking**: Manage order lifecycle with status transitions
- **Order Items**: Support for multiple items per order
- **Customer Orders**: Retrieve orders by customer
- **Pagination**: List orders with pagination support
- **Clean Architecture**: Separated concerns with dependency injection
- **Database**: PostgreSQL with connection pooling
- **Logging**: Structured logging with JSON format
- **Validation**: Request validation using Go validators
- **Error Handling**: Proper error handling and HTTP status codes
- **Health Check**: Health check endpoint for monitoring

## API Endpoints

### Health Check
- `GET /health` - Service health check

### Orders
- `POST /api/v1/orders` - Create a new order
- `GET /api/v1/orders` - List orders with pagination
- `GET /api/v1/orders/:id` - Get order by ID
- `PUT /api/v1/orders/:id` - Update order
- `DELETE /api/v1/orders/:id` - Delete order (only pending orders)
- `PATCH /api/v1/orders/:id/status` - Update order status
- `GET /api/v1/orders/status/:status` - Get orders by status

### Customer Orders
- `GET /api/v1/customers/:customerId/orders` - Get orders for a customer

## Order Status Lifecycle

```
pending → confirmed → processing → shipped → delivered
    ↓         ↓           ↓
cancelled cancelled  cancelled
    ↓
refunded (from delivered)
```

## Environment Variables

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=order_db
DB_SSLMODE=disable

# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Logger Configuration
LOG_LEVEL=info
LOG_FORMAT=json
```

## Quick Start

### Prerequisites
- Go 1.23+
- PostgreSQL 12+
- Docker (optional)

### Local Development

1. **Clone and navigate to the service:**
   ```bash
   cd services/order
   ```

2. **Install dependencies:**
   ```bash
   make deps
   ```

3. **Set up database:**
   ```bash
   # Create database
   createdb order_db
   
   # Run migrations
   make migrate-up
   ```

4. **Set environment variables:**
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=password
   export DB_NAME=order_db
   export DB_SSLMODE=disable
   ```

5. **Build and run:**
   ```bash
   make run
   ```

The service will start on `http://localhost:8080`

### Using Docker

1. **Build and run with Docker:**
   ```bash
   make docker-run
   ```

### Using Make Commands

```bash
# Show all available commands
make help

# Build the application
make build

# Run tests
make test

# Run with coverage
make test-coverage

# Format code
make format

# Lint code
make lint

# Build Docker image
make docker-build

# Development setup
make dev-setup

# Install development tools
make install-tools

# Hot reload development (requires air)
make dev
```

## API Examples

### Create Order
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "123e4567-e89b-12d3-a456-426614174000",
    "shipping_address": "123 Main St, City, State 12345",
    "billing_address": "123 Main St, City, State 12345",
    "notes": "Handle with care",
    "items": [
      {
        "product_id": "123e4567-e89b-12d3-a456-426614174001",
        "quantity": 2,
        "unit_price": 29.99
      }
    ]
  }'
```

### Get Order
```bash
curl http://localhost:8080/api/v1/orders/123e4567-e89b-12d3-a456-426614174000
```

### Update Order Status
```bash
curl -X PATCH http://localhost:8080/api/v1/orders/123e4567-e89b-12d3-a456-426614174000/status \
  -H "Content-Type: application/json" \
  -d '{"status": "confirmed"}'
```

### List Orders
```bash
curl "http://localhost:8080/api/v1/orders?page=1&page_size=10"
```

## Database Schema

### Orders Table
```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    total_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    shipping_address TEXT NOT NULL,
    billing_address TEXT NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Order Items Table
```sql
CREATE TABLE order_items (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10,2) NOT NULL CHECK (unit_price >= 0),
    total_price DECIMAL(10,2) NOT NULL CHECK (total_price >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test ./internal/application/...
```

## Monitoring

The service provides:
- Health check endpoint at `/health`
- Structured JSON logging
- Request/response logging middleware
- Error tracking and recovery

## Contributing

1. Follow the Clean Architecture principles
2. Write tests for new features
3. Use the provided Makefile commands
4. Follow Go naming conventions
5. Add proper error handling and logging

## Deployment

The service can be deployed using:
- Docker containers
- Kubernetes (recommended for production)
- Binary deployment
- Docker Compose (for development)

See the main project's `DEPLOYMENT.md` for detailed deployment instructions.
