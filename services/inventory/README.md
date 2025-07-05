# SaaN Inventory Service

## Overview
The Inventory Service is a comprehensive inventory management service following Clean Architecture principles. It provides inventory tracking, stock management, product lifecycle events, and analytics capabilities for the SaaN system.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │  Order Service  │    │   Chat AI       │
│  (Admin/Web)    │    │    (8081)       │    │   (8090)        │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼───────────────┐
                    │    Inventory Service        │
                    │        (8082)               │
                    │   Clean Architecture        │
                    │                             │
                    │  ┌─────────────────────┐   │
                    │  │   HTTP Interface    │   │
                    │  └─────────┬───────────┘   │
                    │  ┌─────────▼───────────┐   │
                    │  │   Application       │   │
                    │  └─────────┬───────────┘   │
                    │  ┌─────────▼───────────┐   │
                    │  │     Domain          │   │
                    │  └─────────┬───────────┘   │
                    │  ┌─────────▼───────────┐   │
                    │  │  Infrastructure     │   │
                    │  │  • Database         │   │
                    │  │  • Cache (Redis)    │   │
                    │  │  • Events (Kafka)   │   │
                    │  └─────────────────────┘   │
                    └─────────────────────────────┘

## Infrastructure (Clean Architecture)

### Events System
- **Kafka Publisher**: Publishes domain events to Kafka topics
- **Kafka Consumer**: Consumes events from other services (Loyverse sync)
- **Event Types**: Stock updates, product changes, sync events, alerts
- **Noop Publisher**: For testing and development

### Cache Layer
- **Redis Client**: Enhanced caching with TTL and pattern operations
- **Product Caching**: Individual products and product lists
- **Stock Level Caching**: Real-time stock tracking
- **Loyverse Integration**: Cache for Loyverse product data

### Database
- **PostgreSQL Connection**: Structured database configuration
- **Product Management**: CRUD operations for products
- **Stock Tracking**: Inventory levels and movements

## Key Features

### 📊 Event-Driven Architecture
- Domain events for stock updates and product changes
- Kafka integration for real-time synchronization
- Event sourcing for audit trails
- Loyverse synchronization events

### 🏪 Inventory Management
- Product lifecycle management (create, update, delete)
- Multi-store stock level tracking
- Low stock alerts and notifications
- Real-time stock adjustments

### 🔄 Real-time Updates
- Event publishing for downstream services
- Cache invalidation strategies
- Automatic sync with Loyverse POS

### 🔒 Configuration & Security
- Admin authentication for sensitive operations
- Rate limiting and CORS protection
- Health checks and monitoring

## API Endpoints

### Core Inventory APIs
```bash
# Products
GET /api/v1/inventory/products              # List all products
GET /api/v1/inventory/products/:id          # Get specific product
GET /api/v1/inventory/products/:id/stock    # Get product stock levels
GET /api/v1/inventory/search?q=query        # Search products

# Stores
GET /api/v1/inventory/stores                # List all stores
GET /api/v1/inventory/stores/:id/stock      # Get store inventory

# Categories
GET /api/v1/inventory/categories            # List all categories

# Stock Operations
GET /api/v1/inventory/stock/low             # Get low stock items
GET /api/v1/inventory/alerts                # Get inventory alerts
```

### Analytics APIs
```bash
# Dashboard
GET /api/v1/analytics/dashboard             # Main dashboard data

# Performance Analytics
GET /api/v1/analytics/performance/products  # Product performance
GET /api/v1/analytics/performance/categories # Category performance

# Trend Analysis
GET /api/v1/analytics/trends/daily          # Daily movement trends
GET /api/v1/analytics/trends/weekly         # Weekly trend analysis

# AI Suggestions
GET /api/v1/analytics/suggestions/reorder   # Intelligent reorder suggestions
```

### Admin APIs (Requires Authentication)
```bash
# System Operations
POST /api/v1/admin/sync/trigger            # Trigger manual sync
POST /api/v1/admin/cache/refresh           # Refresh cache
GET  /api/v1/admin/stats                   # System statistics
```

### Health & Monitoring
```bash
GET /health                                # Basic health check
GET /ready                                 # Readiness check
GET /ws/inventory                          # WebSocket endpoint
```

## Environment Variables

```bash
# Server Configuration
PORT=8082
GO_ENV=development

# Database
DATABASE_URL=postgres://saan:saan_password@postgres:5432/saan_db?sslmode=disable

# Redis (Cache from Loyverse)
REDIS_ADDR=redis:6379
REDIS_PASSWORD=

# Kafka (Real-time events)
KAFKA_BROKERS=kafka:9092
KAFKA_CONSUMER_GROUP=inventory-service
LOYVERSE_EVENT_TOPIC=loyverse-events

# External Services
ORDER_SERVICE_URL=http://order:8081
CHAT_SERVICE_URL=http://chatbot:8090

# Authentication
ADMIN_TOKEN=saan-dev-admin-2024-secure

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json
```

## Development

### Setup
```bash
cd services/inventory
make dev-setup
make run
```

### Testing
```bash
make test
make test-coverage
```

### Docker
```bash
make docker-build
make docker-run
```

## Data Flow

### 1. Cache Reading (Primary Data Source)
```
Loyverse API → integrations/loyverse → Redis → inventory service
```

### 2. Real-time Updates
```
Loyverse Webhooks → webhooks/loyverse-webhook → Kafka → inventory service
```

### 3. API Consumption
```
Frontend/Services → inventory service → Business Logic → Response
```

## Usage Examples

### Frontend Integration (Admin Dashboard)
```typescript
// Get dashboard data
const dashboard = await fetch('/api/v1/analytics/dashboard')
const data = await dashboard.json()

// Search products
const products = await fetch('/api/v1/inventory/search?q=หมูสามชั้น')
const results = await products.json()

// Check low stock
const lowStock = await fetch('/api/v1/inventory/stock/low')
const alerts = await lowStock.json()
```

### Order Service Integration
```go
// Check stock availability before creating order
resp, err := http.Get("http://inventory:8082/api/v1/inventory/products/123/stock")
if err != nil {
    return err
}

var stockData struct {
    Success bool `json:"success"`
    Data    struct {
        StockLevels []StockLevel `json:"stock_levels"`
    } `json:"data"`
}

if err := json.NewDecoder(resp.Body).Decode(&stockData); err != nil {
    return err
}

// Check if sufficient stock available
for _, stock := range stockData.Data.StockLevels {
    if stock.StoreID == orderStoreID && stock.QuantityOnHand >= requiredQty {
        // Proceed with order
        break
    }
}
```

### Chat AI Integration
```go
// Chat AI can query inventory for user questions
func handleInventoryQuery(query string) (string, error) {
    // Example: "เหลือหมูสามชั้นเท่าไหร่"
    resp, err := http.Get(fmt.Sprintf(
        "http://inventory:8082/api/v1/inventory/search?q=%s", 
        url.QueryEscape(query),
    ))
    
    // Process response and generate natural language answer
    // "ขณะนี้มีหมูสามชั้นคงเหลือ 12.5 กิโลกรัม ที่สาขา 1"
}
```

### WebSocket Real-time Updates
```javascript
// Connect to real-time inventory updates
const ws = new WebSocket('ws://localhost:8082/ws/inventory')

ws.onmessage = function(event) {
    const update = JSON.parse(event.data)
    if (update.type === 'stock_update') {
        updateDashboard(update.data)
    }
}
```

## Service Dependencies

### Required Services
- **Redis**: Cache layer (populated by integrations/loyverse)
- **PostgreSQL**: Business metadata and analytics
- **Kafka**: Real-time event processing

### Optional Dependencies
- **Order Service**: For order-inventory integration
- **Chat Service**: For AI-powered inventory queries

## Monitoring & Observability

### Health Checks
```bash
# Basic health
curl http://localhost:8082/health

# Comprehensive readiness check
curl http://localhost:8082/ready
```

### Metrics (Admin)
```bash
# System statistics
curl -H "X-Admin-Token: saan-dev-admin-2024-secure" \
     http://localhost:8082/api/v1/admin/stats
```

## Performance Considerations

### Caching Strategy
- **Redis**: Primary data cache (updated by loyverse integration)
- **Application**: In-memory caching for frequently accessed data
- **API**: Response caching with appropriate TTL

### Database Optimization
- **Indexes**: Strategic indexing for search and analytics queries
- **Connection Pooling**: Efficient database connection management
- **Query Optimization**: Optimized SQL queries for analytics

### Scalability
- **Horizontal Scaling**: Stateless design allows multiple instances
- **Load Balancing**: Can be load balanced across multiple containers
- **Event Processing**: Kafka ensures reliable event processing

This service completes the missing piece in the SaaN ecosystem, providing powerful inventory management and business intelligence capabilities while maintaining clean separation of concerns with the Loyverse integration layer.
