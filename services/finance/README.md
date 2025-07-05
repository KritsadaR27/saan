# Finance Service

A comprehensive financial management microservice implementing the **Profit First** methodology for the SAAN business management platform.

## 🎯 Overview

The Finance Service provides complete cash flow management, profit allocation, expense tracking, and financial reporting capabilities for branches and vehicles in the SAAN ecosystem.

### Key Features

- ✅ **End-of-Day Cash Processing** - Automated daily cash summaries with Profit First allocations
- ✅ **Profit First Implementation** - Configurable profit allocation rules (5% profit, 50% owner pay, 15% tax, 30% operating)
- ✅ **Expense Management** - Categorized expense tracking with receipt support
- ✅ **Cash Transfer System** - Batch processing for supplier payments and transfers
- ✅ **Real-time Cash Flow** - Live tracking of cash movements with running balances
- ✅ **Financial Reconciliation** - Manual cash reconciliation workflow
- ✅ **Multi-Entity Support** - Separate tracking for branches and vehicles

## 🏗️ Architecture

The service follows **Clean Architecture** principles with clear separation of concerns:

```
finance/
├── cmd/
│   ├── main.go              # Application entry point
│   ├── migrate/             # Database migration runner
│   └── testsetup/           # Test database utilities
├── internal/
│   ├── domain/              # Business entities and interfaces
│   ├── application/         # Business logic services
│   ├── infrastructure/      # External concerns (DB, cache, etc.)
│   └── transport/           # HTTP handlers and routing
├── migrations/              # Database schema migrations
└── .env.example            # Environment configuration template
```

### Domain Entities

- **DailyCashSummary** - End-of-day cash summary with allocations
- **ProfitAllocationRule** - Configurable allocation percentages
- **CashTransfer** - Individual transfer records
- **CashTransferBatch** - Grouped transfers for processing
- **ExpenseEntry** - Manual expense entries
- **CashFlowRecord** - Real-time cash flow tracking

## 🚀 Getting Started

### Prerequisites

- Go 1.23+
- PostgreSQL 13+
- Redis 6+
- Docker & Docker Compose (for development)

### Environment Setup

1. Copy the environment template:
```bash
cp .env.example .env
```

2. Configure your environment variables:
```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=finance_db
DB_SSLMODE=disable

# Redis Configuration  
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Service Configuration
PORT=8085
GIN_MODE=release
```

### Database Setup

1. Create the database:
```sql
CREATE DATABASE finance_db;
```

2. Run migrations:
```bash
go run cmd/migrate/main.go
```

### Running the Service

#### Development Mode
```bash
go run cmd/main.go
```

#### Using Docker Compose (from project root)
```bash
docker-compose up finance-service
```

#### Production Build
```bash
go build -o finance-service cmd/main.go
./finance-service
```

## 📖 API Documentation

The service exposes a REST API on port **8085** with the following endpoints:

### Core Operations

#### Process End of Day
```http
POST /api/finance/end-of-day
Content-Type: application/json

{
  "business_date": "2024-01-15",
  "branch_id": "uuid",
  "vehicle_id": "uuid",
  "total_sales": 1500.00,
  "cod_collections": 300.00,
  "opening_cash": 100.00
}
```

#### Add Expense Entry
```http
POST /api/finance/expenses
Content-Type: application/json

{
  "summary_id": "uuid",
  "category": "fuel",
  "description": "Gasoline for delivery truck",
  "amount": 75.50,
  "entered_by": "user_uuid"
}
```

#### Create Transfer Batch
```http
POST /api/finance/transfers/batch
Content-Type: application/json

{
  "branch_id": "uuid",
  "transfers": [
    {
      "transfer_type": "supplier_payment",
      "recipient_name": "ABC Supplier",
      "recipient_account": "12345678",
      "amount": 500.00,
      "description": "Weekly inventory payment"
    }
  ],
  "authorized_by": "user_uuid"
}
```

#### Get Cash Status
```http
GET /api/finance/cash-status
```

#### Reconcile Cash
```http
POST /api/finance/reconcile
Content-Type: application/json

{
  "summary_id": "uuid",
  "actual_cash": 2456.78,
  "reconciled_by": "user_uuid",
  "notes": "End of day reconciliation"
}
```

### Management Operations

#### Update Allocation Rule
```http
PUT /api/finance/allocation-rules/{id}
Content-Type: application/json

{
  "profit_percentage": 5.0,
  "owner_pay_percentage": 50.0,
  "tax_percentage": 15.0,
  "effective_from": "2024-01-01T00:00:00Z",
  "updated_by_user_id": "user_uuid"
}
```

#### Execute Transfer Batch
```http
POST /api/finance/transfers/batch/{id}/execute
```

For complete API documentation, see [API.md](../../../docs/services/finance/API.md).

## 🧪 Testing

### Unit Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/application/ -v
```

### Integration Tests
```bash
# Setup test database
go run cmd/testsetup/main.go setup

# Run integration tests (requires test DB)
go test -tags=integration ./...

# Cleanup test database
go run cmd/testsetup/main.go teardown
```

### Test Structure
- **Unit Tests** - Test business logic in isolation with mocks
- **Integration Tests** - Test database operations and external dependencies
- **API Tests** - Test HTTP endpoints end-to-end

## 💰 Profit First Implementation

The service implements the **Profit First** cash management system:

### Default Allocation Rules
- **Profit Account**: 5% - Set aside for business profit
- **Owner Pay Account**: 50% - Owner/operator compensation  
- **Tax Account**: 15% - Tax obligations
- **Operating Account**: 30% - Business operations

### Entity-Specific Rules
Allocation rules can be customized per:
- **Branch** - Different rules for different locations
- **Vehicle** - Mobile service specific allocations
- **Time Period** - Rules can change over time

### Allocation Process
1. **Revenue Recording** - Sales and COD collections recorded
2. **Automatic Allocation** - Amounts distributed per active rules
3. **Account Updates** - Individual account balances updated
4. **Transfer Creation** - Scheduled transfers to external accounts

## 🔧 Configuration

### Database Configuration
- **Connection Pooling** - Configurable pool size and timeouts
- **Migration Management** - Automated schema updates
- **Backup & Recovery** - Daily automated backups

### Cache Configuration  
- **Redis Integration** - Session and calculation caching
- **TTL Management** - Automatic cache expiration
- **Fallback Strategy** - Graceful degradation when cache unavailable

### Service Configuration
- **Rate Limiting** - API request throttling
- **Authentication** - JWT token validation
- **Logging** - Structured JSON logging with levels

## 🔍 Monitoring

### Health Checks
```http
GET /health
GET /ready
```

### Metrics
- Request latency and throughput
- Database connection health
- Cache hit/miss ratios
- Business metrics (daily revenue, allocations)

### Logging
All operations are logged with:
- Request tracing IDs
- User context
- Execution timing
- Error details

## 🚚 Deployment

### Docker
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o finance-service cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/finance-service .
CMD ["./finance-service"]
```

### Kubernetes
Deploy using the provided manifests in `/infrastructure/k8s/`.

### Environment Variables
Ensure all required environment variables are set in your deployment environment.

## 🛠️ Development

### Adding New Features

1. **Domain First** - Define entities and interfaces in `/internal/domain/`
2. **Business Logic** - Implement services in `/internal/application/`
3. **Infrastructure** - Add repository implementations in `/internal/infrastructure/`
4. **Transport** - Create HTTP handlers in `/internal/transport/`
5. **Tests** - Add comprehensive unit and integration tests
6. **Documentation** - Update API docs and README

### Code Standards
- Follow Go naming conventions
- Use dependency injection
- Implement proper error handling
- Add comprehensive logging
- Write tests for all business logic

## 📚 Related Documentation

- [API Documentation](../../../docs/services/finance/API.md)
- [Integration Guide](../../../docs/services/finance/INTEGRATION.md)
- [Troubleshooting](../../../docs/services/finance/TROUBLESHOOTING.md)
- [Implementation Plan](../../../docs/services/finance/IMPLEMENTATION_PLAN.md)

## 🤝 Contributing

1. Follow the project coding standards
2. Write tests for new features
3. Update documentation
4. Submit pull requests for review

## 📄 License

This project is part of the SAAN business management platform. See the main project LICENSE file for details.
