# Customer Service (Port 8110)

The Customer Service is a core microservice in the SaaN system responsible for managing customer data, addresses, Thai address lookups, customer tiers, and integration with Loyverse POS system.

## Features

### Core Customer Management
- **CRUD Operations**: Create, read, update, and delete customers
- **Customer Search**: Search by email, phone, or ID
- **Customer Validation**: Comprehensive data validation
- **Soft Delete**: Maintains data integrity with soft deletion

### Address Management
- **Multiple Addresses**: Support for multiple addresses per customer (home, work, billing, shipping)
- **Thai Address Integration**: Integration with Thai administrative divisions
- **Address Validation**: Postal code and address line validation
- **Default Address**: Set and manage default addresses per customer

### Customer Tier System
- **Automatic Tier Calculation**: Based on total spending
- **Tier Levels**: Bronze, Silver, Gold, Platinum, Diamond
- **Tier Thresholds**: Configurable spending thresholds
- **Tier Events**: Published when tier changes occur

### Loyverse Integration
- **Customer Sync**: Sync customers with Loyverse POS
- **Auto-Creation**: Create Loyverse customers when orders are paid
- **ID Mapping**: Maintain mapping between internal and Loyverse customer IDs
- **Search Integration**: Find existing customers in Loyverse

### Caching & Performance
- **Redis Caching**: Cache frequently accessed customer data
- **Thai Address Cache**: Cache Thai address lookups
- **Configurable TTL**: Adjustable cache time-to-live

### Event Publishing
- **Customer Events**: Publish events for customer lifecycle
- **Tier Events**: Notify when customer tiers change
- **Loyverse Events**: Track Loyverse synchronization
- **Kafka Integration**: Reliable event delivery

## API Endpoints

### Customer Management
```
POST   /api/v1/customers                    # Create customer
GET    /api/v1/customers                    # List customers (with filters)
GET    /api/v1/customers/:id                # Get customer by ID
PUT    /api/v1/customers/:id                # Update customer
DELETE /api/v1/customers/:id                # Delete customer (soft)
GET    /api/v1/customers/search/email       # Search by email
GET    /api/v1/customers/search/phone       # Search by phone
```

### Address Management
```
POST   /api/v1/customers/:id/addresses            # Add address
PUT    /api/v1/customers/:id/addresses/:addr_id   # Update address
DELETE /api/v1/customers/:id/addresses/:addr_id   # Delete address
POST   /api/v1/customers/:id/addresses/:addr_id/default # Set default
```

### Thai Address Lookup
```
GET    /api/v1/addresses/thai/search              # Search Thai addresses
GET    /api/v1/addresses/thai/postal/:code        # Get by postal code
```

### Loyverse Integration
```
POST   /api/v1/customers/:id/sync/loyverse        # Sync with Loyverse
```

### Health Check
```
GET    /health                                    # Service health
```

## Database Schema

### customers
- `id` (UUID, PK) - Unique customer identifier
- `first_name` (VARCHAR) - Customer first name
- `last_name` (VARCHAR) - Customer last name
- `email` (VARCHAR, UNIQUE) - Customer email
- `phone` (VARCHAR, UNIQUE) - Customer phone
- `date_of_birth` (DATE) - Customer birth date
- `gender` (VARCHAR) - Customer gender
- `tier` (ENUM) - Customer tier level
- `loyverse_id` (VARCHAR, UNIQUE) - Loyverse customer ID
- `total_spent` (DECIMAL) - Total customer spending
- `order_count` (INTEGER) - Number of orders
- `last_order_date` (TIMESTAMP) - Last order date
- `delivery_route_id` (UUID, FK) - Assigned delivery route
- `is_active` (BOOLEAN) - Soft delete flag
- `created_at` (TIMESTAMP) - Creation time
- `updated_at` (TIMESTAMP) - Last update time

### customer_addresses
- `id` (UUID, PK) - Unique address identifier
- `customer_id` (UUID, FK) - Reference to customer
- `type` (ENUM) - Address type (home, work, billing, shipping)
- `address_line1` (VARCHAR) - Primary address line
- `address_line2` (VARCHAR) - Secondary address line
- `thai_address_id` (UUID, FK) - Reference to Thai address
- `postal_code` (VARCHAR) - Postal code
- `latitude` (DECIMAL) - GPS latitude
- `longitude` (DECIMAL) - GPS longitude
- `is_default` (BOOLEAN) - Default address flag
- `delivery_notes` (TEXT) - Delivery instructions
- `is_active` (BOOLEAN) - Soft delete flag
- `created_at` (TIMESTAMP) - Creation time
- `updated_at` (TIMESTAMP) - Last update time

### thai_addresses
- `id` (UUID, PK) - Unique identifier
- `province` (VARCHAR) - Province name
- `district` (VARCHAR) - District name
- `subdistrict` (VARCHAR) - Subdistrict name
- `postal_code` (VARCHAR) - Postal code
- `province_code` (VARCHAR) - Province code
- `district_code` (VARCHAR) - District code
- `created_at` (TIMESTAMP) - Creation time
- `updated_at` (TIMESTAMP) - Last update time

### delivery_routes
- `id` (UUID, PK) - Unique route identifier
- `name` (VARCHAR) - Route name
- `description` (TEXT) - Route description
- `is_active` (BOOLEAN) - Active status
- `created_at` (TIMESTAMP) - Creation time
- `updated_at` (TIMESTAMP) - Last update time

## Customer Tier System

### Tier Levels and Thresholds
- **Bronze**: ฿0 - ฿4,999 (Default tier)
- **Silver**: ฿5,000 - ฿19,999
- **Gold**: ฿20,000 - ฿49,999
- **Platinum**: ฿50,000 - ฿99,999
- **Diamond**: ฿100,000+

### Tier Calculation
Tiers are automatically calculated based on `total_spent` and updated when:
- A new order is completed
- Manual tier updates are performed
- Customer spending is modified

## Environment Variables

```bash
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=saan_user
DB_PASSWORD=saan_password
DB_NAME=saan_customer
DB_SSLMODE=disable

# Redis
REDIS_ADDR=redis:6379
REDIS_PASSWORD=

# Kafka
KAFKA_BROKERS=kafka:9092

# Loyverse API
LOYVERSE_API_URL=https://api.loyverse.com/v1.0
LOYVERSE_API_KEY=your_api_key

# Service
PORT=8110
GIN_MODE=release
LOG_LEVEL=info
```

## Development Setup

1. **Clone and setup**:
   ```bash
   cd services/customer
   make dev-setup
   ```

2. **Configure environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your settings
   ```

3. **Install dependencies**:
   ```bash
   make deps
   ```

4. **Run migrations**:
   ```bash
   make migrate-up
   ```

5. **Run the service**:
   ```bash
   make run
   # or for hot reload
   make dev
   ```

## Docker

### Build and run with Docker:
```bash
make docker-build
make docker-run
```

### Using Docker Compose:
```bash
# From project root
docker-compose up customer-service
```

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Lint code
make lint
```

## Architecture

The service follows Clean Architecture principles with a well-organized infrastructure layer:

```
├── cmd/                    # Application entry point
├── internal/
│   ├── domain/            # Business logic and entities
│   │   ├── entity/       # Domain entities
│   │   └── repository/   # Repository interfaces
│   ├── application/       # Use cases and services
│   ├── infrastructure/    # Infrastructure layer (unified)
│   │   ├── config/       # Configuration management
│   │   ├── database/     # Database connections & repositories
│   │   ├── cache/        # Redis cache implementation
│   │   ├── events/       # Event streaming (Kafka, NoOp)
│   │   ├── loyverse/     # External Loyverse integration
│   │   └── external/     # Other external service integrations
│   └── transport/
│       └── http/         # HTTP handlers and routes
├── migrations/            # Database migrations
└── Dockerfile            # Container configuration
```

## Integration Points

### With Order Service (8088)
- Receives order completion events
- Updates customer total_spent and tier
- Triggers Loyverse sync for new customers

### With Loyverse POS
- Creates customers when orders are paid
- Syncs customer data bidirectionally
- Maintains ID mapping

### With Payment Service (8087)
- Receives payment completion events
- Updates customer spending data
- Triggers tier recalculation

### With Shipping Service (8086)
- Provides customer address data
- Manages delivery route assignments
- Handles address validation

## Events Published

### Customer Events
- `customer.created` - When a new customer is created
- `customer.updated` - When customer data is modified
- `customer.deleted` - When a customer is soft deleted
- `customer.tier.updated` - When customer tier changes
- `customer.loyverse.synced` - When synced with Loyverse

## Monitoring and Health

### Health Check
```bash
curl http://localhost:8110/health
```

### Metrics
- Customer creation/update rates
- Tier distribution
- Loyverse sync success/failure rates
- Cache hit/miss ratios
- Database query performance

## Security

- Input validation on all endpoints
- SQL injection protection
- Rate limiting (via API Gateway)
- Authentication via JWT (when integrated)
- Secure Loyverse API key handling

## Performance Considerations

- Database connection pooling
- Redis caching for frequently accessed data
- Efficient database indexes
- Batch processing for bulk operations
- Async event publishing

## Future Enhancements

- Customer segmentation features
- Advanced analytics and reporting
- Customer loyalty program integration
- Multi-language support for Thai addresses
- Customer preferences and settings
- Integration with marketing tools
