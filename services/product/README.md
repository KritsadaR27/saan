# Product Service

A comprehensive product management microservice for the SaaN system, implementing the Master Data Protection Pattern with advanced pricing, VIP access control, and availability management.

## Features

- **Master Data Protection**: Separates external system data from admin-controlled fields
- **Advanced Pricing**: Supports base, VIP, bulk, and promotional pricing
- **VIP Access Control**: Manages VIP-only products and special pricing
- **Availability Management**: Real-time inventory tracking and availability control
- **Category Management**: Hierarchical product categorization
- **Search & Filtering**: Powerful product search with multiple filters
- **Caching**: Redis-based caching for improved performance
- **Event Streaming**: Kafka integration for real-time updates
- **Clean Architecture**: Domain-driven design with clear separation of concerns

## Architecture

```
cmd/
├── main.go                    # Application entry point

internal/
├── domain/                    # Domain layer
│   ├── entity/               # Domain entities
│   └── repository/           # Repository interfaces
├── application/              # Application layer
│   └── service/             # Business logic services
├── infrastructure/          # Infrastructure layer
│   ├── config/             # Configuration
│   ├── database/           # Database connection & repositories
│   ├── cache/              # Redis cache implementation
│   ├── events/             # Event streaming (Kafka, NoOp)
│   └── loyverse/           # External Loyverse integration
└── transport/               # Transport layer
    └── http/               # HTTP handlers and middleware

migrations/                  # Database migrations
```

## Getting Started

### Prerequisites

- Go 1.23+
- PostgreSQL 13+
- Redis 6+
- Kafka (optional, for events)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd services/product
```

2. Install dependencies:
```bash
make install
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Run database migrations:
```bash
make migrate-up
```

5. Start the service:
```bash
make run
```

### Environment Variables

```bash
# Server
PORT=8083
ENVIRONMENT=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=saan_products
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Kafka
KAFKA_BROKERS=localhost:9092

# External Services
LOYVERSE_SERVICE_URL=http://localhost:8080
ORDER_SERVICE_URL=http://localhost:8081
INVENTORY_SERVICE_URL=http://localhost:8082

# Security
JWT_SECRET=your-secret-key
INTERNAL_API_KEY=internal-api-key

# Cache TTL (seconds)
CACHE_PRODUCT_TTL=3600
CACHE_PRICE_TTL=1800
CACHE_INVENTORY_TTL=300
```

## API Documentation

### Product Management

#### Create Product
```http
POST /api/v1/products
Content-Type: application/json

{
  "name": "Sample Product",
  "description": "Product description",
  "sku": "PROD-001",
  "barcode": "1234567890123",
  "base_price": 99.99,
  "unit": "piece",
  "is_active": true,
  "is_vip_only": false
}
```

#### Get Product
```http
GET /api/v1/products/{id}
```

#### List Products
```http
GET /api/v1/products?limit=50&offset=0&category_id={uuid}&is_active=true
```

#### Search Products
```http
GET /api/v1/products?q=search_term&limit=50
```

#### Update Product
```http
PUT /api/v1/products/{id}
Content-Type: application/json

{
  "name": "Updated Product Name",
  "base_price": 119.99
}
```

#### Delete Product
```http
DELETE /api/v1/products/{id}
```

### Availability Management

#### Get Product Availability
```http
GET /api/v1/products/{id}/availability
```

#### Update Product Availability
```http
POST /api/v1/products/{id}/availability
Content-Type: application/json

{
  "location_id": "uuid",
  "is_available": false,
  "reason": "Out of stock"
}
```

### Pricing Management

#### Get Product Pricing
```http
GET /api/v1/pricing/products/{id}?quantity=10&customer_group=wholesale&vip_level=gold
```

#### Set VIP Pricing
```http
POST /api/v1/pricing/products/{id}/vip
Content-Type: application/json

{
  "vip_tier_id": "uuid",
  "price": 79.99
}
```

#### Set Bulk Pricing
```http
POST /api/v1/pricing/products/{id}/bulk
Content-Type: application/json

{
  "min_quantity": 10,
  "max_quantity": 50,
  "price": 89.99
}
```

#### Set Promotional Pricing
```http
POST /api/v1/pricing/products/{id}/promotional
Content-Type: application/json

{
  "price": 69.99,
  "valid_from": "2024-01-01T00:00:00Z",
  "valid_to": "2024-01-31T23:59:59Z",
  "promotion_name": "New Year Sale"
}
```

### Internal APIs

#### Get Product (Internal)
```http
GET /api/v1/internal/products/{id}
X-API-Key: internal-api-key
```

#### Get Products Batch
```http
POST /api/v1/internal/products/batch
Content-Type: application/json
X-API-Key: internal-api-key

{
  "product_ids": ["uuid1", "uuid2", "uuid3"]
}
```

## Master Data Protection

The Product Service implements the Master Data Protection Pattern to ensure data integrity when syncing from external systems like Loyverse:

### Protected Fields

**Source Fields** (managed by external systems):
- `loyverse_id`
- `name`
- `description`
- `sku`
- `barcode`
- `base_price`
- `unit`
- `weight`
- `is_active`

**Admin Fields** (protected from external updates):
- `is_vip_only`
- `category_id` (can be overridden)
- `tags`
- Custom pricing rules
- Availability overrides

### Sync Behavior

1. **New Products**: All fields are populated from external system
2. **Existing Products**: Only source fields are updated
3. **Manual Override**: Admin can mark products for manual control
4. **Conflict Resolution**: Admin fields always take precedence

## Events

The service publishes events to Kafka for downstream consumption:

### Product Events
- `product.created`
- `product.updated`
- `product.deleted`
- `product.synced`

### Pricing Events
- `price.changed`

### Inventory Events
- `stock.changed`
- `inventory.low`
- `inventory.alert`

## Caching Strategy

### Product Cache
- **TTL**: 1 hour
- **Key Pattern**: `product:{id}`
- **Invalidation**: On update/delete

### Price Cache
- **TTL**: 30 minutes
- **Key Pattern**: `price:{product_id}:{customer_id}`
- **Invalidation**: On price changes

### Availability Cache
- **TTL**: 5 minutes
- **Key Pattern**: `availability:{product_id}:{location_id}`
- **Invalidation**: On stock changes

### Search Cache
- **TTL**: 10 minutes
- **Key Pattern**: `products:{filter_hash}`
- **Invalidation**: On product changes

## Database Schema

### Products Table
```sql
CREATE TABLE products (
    id UUID PRIMARY KEY,
    loyverse_id VARCHAR(255) UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    sku VARCHAR(255) UNIQUE NOT NULL,
    barcode VARCHAR(255) UNIQUE,
    category_id UUID,
    base_price DECIMAL(10,2) NOT NULL,
    unit VARCHAR(50) NOT NULL,
    weight DECIMAL(10,3),
    is_active BOOLEAN DEFAULT TRUE,
    is_vip_only BOOLEAN DEFAULT FALSE,
    tags TEXT[],
    data_source_type VARCHAR(50) NOT NULL,
    data_source_id VARCHAR(255),
    last_synced_at TIMESTAMP WITH TIME ZONE,
    is_manual_override BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    version INTEGER DEFAULT 1
);
```

### Categories Table
```sql
CREATE TABLE categories (
    id UUID PRIMARY KEY,
    loyverse_id VARCHAR(255) UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    parent_id UUID,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    data_source_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Prices Table
```sql
CREATE TABLE prices (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL,
    price_type VARCHAR(50) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'THB',
    min_quantity INTEGER,
    max_quantity INTEGER,
    vip_tier_id UUID,
    valid_from TIMESTAMP WITH TIME ZONE,
    valid_to TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Inventory Table
```sql
CREATE TABLE inventory (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL,
    location_id UUID NOT NULL,
    stock_level DECIMAL(10,3) NOT NULL DEFAULT 0,
    reserved_level DECIMAL(10,3) NOT NULL DEFAULT 0,
    available_level DECIMAL(10,3) NOT NULL DEFAULT 0,
    low_stock_threshold DECIMAL(10,3),
    is_available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id, location_id)
);
```

## Development

### Running Tests
```bash
make test
make test-coverage
```

### Linting
```bash
make lint
```

### Database Operations
```bash
# Reset database
make db-reset

# Create new migration
make migrate-create NAME=add_new_field

# Run migrations
make migrate-up
make migrate-down
```

### Docker Development
```bash
# Build Docker image
make docker-build

# Run container
make docker-run

# Docker Compose
make docker-compose-up
make docker-compose-down
```

## Deployment

### Production Build
```bash
make prod-build
```

### Docker Deployment
```bash
docker build -t product-service:latest .
docker run -p 8083:8083 --env-file .env product-service:latest
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: product-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: product-service
  template:
    metadata:
      labels:
        app: product-service
    spec:
      containers:
      - name: product-service
        image: product-service:latest
        ports:
        - containerPort: 8083
        env:
        - name: PORT
          value: "8083"
        # Add other environment variables
```

## Monitoring

### Health Check
```http
GET /health
```

### Metrics
The service exposes metrics for monitoring:
- Request latency
- Error rates
- Cache hit rates
- Database connection pool status

### Logging
Structured JSON logging with correlation IDs for request tracing.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run linting and tests
6. Submit a pull request

## License

[Add license information]
