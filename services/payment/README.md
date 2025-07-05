# 💳 SAAN Payment Service

A comprehensive payment processing service with enhanced Loyverse POS integration, multi-store management, and delivery context tracking. Built following SAAN Clean Architecture standards.

## 🎯 Features

### 🏪 Multi-Store Management
- **Store Assignment**: Automatic assignment of payments to optimal Loyverse stores
- **Load Balancing**: Intelligent distribution based on store capacity and workload
- **Store Analytics**: Comprehensive analytics per store location

### 💰 Payment Processing
- **Multiple Methods**: Cash, bank transfer, COD, digital wallets
- **Multiple Channels**: Loyverse POS, SAAN App, Chat, Delivery, Web Portal
- **Payment Timing**: Prepaid and Cash on Delivery (COD)
- **Status Tracking**: Real-time payment status updates

### 🚚 Delivery Integration
- **COD Context**: Full delivery context tracking for COD payments
- **GPS Tracking**: Pickup and delivery location tracking
- **Driver Management**: Driver assignment and tracking
- **Real-time Updates**: Live delivery status updates

### 📊 Data Retrieval (3 Types)

#### Type 1: Store-based Queries
```bash
# Get payments for a specific store
GET /api/v1/stores/{store_id}/payments

# Get store analytics
GET /api/v1/stores/{store_id}/analytics?date_from=2024-01-01&date_to=2024-01-31
```

#### Type 2: Customer-based Queries
```bash
# Get customer payments
GET /api/v1/customers/{customer_id}/payments

# Get customer payment history
GET /api/v1/customers/{customer_id}/payment-history

# Get customer payment statistics
GET /api/v1/customers/{customer_id}/payment-stats
```

#### Type 3: Order-based Queries
```bash
# Get order payments
GET /api/v1/orders/{order_id}/payments

# Get order payment summary
GET /api/v1/orders/{order_id}/payment-summary

# Get order payment timeline
GET /api/v1/orders/{order_id}/payment-timeline
```

## 🏗️ Architecture

```
├── cmd/                    # Application entry points
├── internal/
│   ├── application/        # Application layer
│   │   ├── dto/           # Data Transfer Objects
│   │   └── usecase/       # Business logic use cases
│   ├── domain/            # Domain layer
│   │   ├── entity/        # Domain entities
│   │   └── repository/    # Repository interfaces
│   ├── infrastructure/    # Infrastructure layer
│   │   ├── config/        # Configuration
│   │   ├── repository/    # Repository implementations
│   │   ├── cache/         # Redis cache
│   │   ├── events/        # Event publishing
│   │   └── external/      # External service clients
│   └── transport/         # Transport layer
│       └── http/          # HTTP handlers and routes
├── migrations/            # Database migrations
└── docs/                 # Documentation
```

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 14+
- Redis 6+
- Docker & Docker Compose

### Installation

1. **Clone and setup**
```bash
git clone <repository>
cd services/payment
make setup-dev
```

2. **Configure environment**
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. **Start dependencies**
```bash
docker-compose up -d postgres redis
```

4. **Run migrations**
```bash
make migrate-up
```

5. **Start the service**
```bash
make run
```

## 📋 API Documentation

### Core Payment Operations

#### Create Payment
```bash
POST /api/v1/payments
Content-Type: application/json

{
  "order_id": "123e4567-e89b-12d3-a456-426614174000",
  "customer_id": "123e4567-e89b-12d3-a456-426614174001",
  "payment_method": "cod_cash",
  "payment_channel": "saan_app",
  "payment_timing": "cod",
  "amount": 150.00,
  "currency": "THB",
  "delivery_context": {
    "delivery_id": "123e4567-e89b-12d3-a456-426614174002",
    "delivery_address": "123 Main St, Bangkok",
    "estimated_arrival": "2024-01-15T14:30:00Z"
  }
}
```

#### Update Payment Status
```bash
PUT /api/v1/payments/{payment_id}/status
Content-Type: application/json

{
  "status": "completed",
  "loyverse_receipt_id": "RCP123456",
  "loyverse_payment_type": "cash"
}
```

### Store-based Queries (Type 1)

#### Get Store Payments
```bash
GET /api/v1/stores/STORE001/payments?status=completed&limit=50&offset=0
```

#### Get Store Analytics
```bash
GET /api/v1/stores/STORE001/analytics?date_from=2024-01-01&date_to=2024-01-31
```

### Customer-based Queries (Type 2)

#### Get Customer Payments
```bash
GET /api/v1/customers/{customer_id}/payments?payment_method=cod_cash&limit=20
```

#### Get Customer Payment History
```bash
GET /api/v1/customers/{customer_id}/payment-history?limit=10
```

### Order-based Queries (Type 3)

#### Get Order Payment Summary
```bash
GET /api/v1/orders/{order_id}/payment-summary
```

**Response:**
```json
{
  "message": "Order payment summary retrieved successfully",
  "data": {
    "order_id": "123e4567-e89b-12d3-a456-426614174000",
    "total_amount": 150.00,
    "paid_amount": 150.00,
    "pending_amount": 0.00,
    "refunded_amount": 0.00,
    "currency": "THB",
    "payment_status": "fully_paid",
    "transaction_count": 1,
    "last_payment_at": "2024-01-15T14:30:00Z",
    "payment_methods": ["cod_cash"]
  }
}
```

## 🧪 Testing

### Run Tests
```bash
# Unit tests
make test

# Integration tests
make test-integration

# Load tests
make load-test

# Test coverage
make test-coverage
```

### Test Data Retrieval Types
```bash
# Test Type 1: Store-based
make test-store-data

# Test Type 2: Customer-based
make test-customer-data CUSTOMER_ID=your-customer-id

# Test Type 3: Order-based
make test-order-data ORDER_ID=your-order-id
```

## 🔧 Development

### Available Commands
```bash
make help                 # Show all available commands
make run                  # Run service locally
make run-dev              # Run with auto-reload
make build                # Build binary
make test                 # Run tests
make lint                 # Run linters
make docker-build         # Build Docker image
make migrate-up           # Run database migrations
```

### Database Operations
```bash
# Create new migration
make migrate-create NAME=add_new_feature

# Reset database
make db-reset

# Seed test data
make seed-data
```

## 🌍 Environment Variables

```bash
# Server Configuration
SERVER_PORT=8087
SERVER_ENVIRONMENT=development

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=saan_user
DB_PASSWORD=saan_password
DB_NAME=saan_payment
DB_SSLMODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Loyverse Integration
LOYVERSE_API_URL=https://api.loyverse.com/v1.0
LOYVERSE_WEBHOOK_SECRET=your-webhook-secret

# External Services
ORDER_SERVICE_URL=http://localhost:8081
CUSTOMER_SERVICE_URL=http://localhost:8082
SHIPPING_SERVICE_URL=http://localhost:8085
```

## 📊 Monitoring

### Health Check
```bash
curl http://localhost:8087/health
```

### Service Metrics
- Payment processing rates
- Store assignment efficiency
- COD collection rates
- Error rates by payment method
- Response time by query type

## 🔗 Integration

### Loyverse POS Integration
- Automatic store assignment based on capacity
- Real-time receipt synchronization
- Payment method mapping
- Store analytics aggregation

### Other SAAN Services
- **Order Service**: Payment status updates
- **Customer Service**: Payment history
- **Shipping Service**: COD context tracking
- **Chat Service**: Payment notifications

## 🛠️ Deployment

### Docker
```bash
# Build and run
make docker-build
make docker-run

# Using Docker Compose
make compose-up
```

### Production
```bash
# Deploy to staging
make deploy-staging

# Deploy to production
make deploy-prod
```

## 📈 Performance

### Optimization Features
- Database connection pooling
- Redis caching for frequent queries
- Indexed database queries
- Efficient pagination
- Batch operations support

### Scaling Considerations
- Horizontal scaling support
- Load balancer compatible
- Database read replicas ready
- Event-driven architecture

## 🔐 Security

- Input validation and sanitization
- SQL injection prevention
- Rate limiting
- Authentication middleware ready
- Audit trail logging

## 📝 Contributing

1. Follow SAAN Clean Architecture standards
2. Write comprehensive tests
3. Update documentation
4. Follow Go best practices
5. Use conventional commit messages

## 📄 License

Part of the SAAN System - All rights reserved.

---

> 💡 **Tip**: Use `make help` to see all available commands and `make test-{type}-data` to test the three data retrieval patterns!
