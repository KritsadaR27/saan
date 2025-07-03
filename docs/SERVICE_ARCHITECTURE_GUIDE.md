# 🏗️ SAAN Service Architecture Guide

## 📋 **Overview**

แนวทางการออกแบบ microservices สำหรับ SAAN system ตาม Clean Architecture principles โดยทุก service ใช้โครงสร้างเดียวกัน

---

## 🎯 **Standard Service Structure**

### **📁 Universal Directory Structure**
```
services/{service-name}/
├── cmd/
│   └── main.go                   # Entry point
├── internal/
│   ├── domain/                   # 📦 Core Business Logic
│   │   ├── entity/              # Business entities
│   │   └── repository/          # Repository interfaces ONLY
│   ├── application/             # 📋 Use Cases & Business Logic
│   │   ├── {entity}_usecase.go # Business use cases
│   │   └── {feature}_usecase.go# Feature-specific logic
│   ├── infrastructure/          # 🔧 External Dependencies
│   │   ├── config/             # Configuration
│   │   ├── database/           # Database implementation
│   │   ├── cache/              # Redis implementation  
│   │   ├── events/             # Kafka implementation
│   │   ├── loyverse/           # Loyverse integration (if needed)
│   │   └── external/           # Other external APIs
│   └── transport/              # 🌐 Input/Output Adapters
│       └── http/
│           ├── handler/        # HTTP handlers
│           ├── middleware/     # HTTP middleware
│           └── routes.go       # Route definitions
├── migrations/                  # Database migrations
├── Dockerfile                  # Container definition
├── go.mod                      # Go dependencies
└── go.sum                      # Dependency checksums
```

---

## 🧩 **Service Types & Variations**

### **Type 1: Core Business Services**
**Services:** Order, Customer, Product, Inventory, Shipping, Payment, Finance

**Features:**
- ✅ Complete domain entities
- ✅ Complex business logic
- ✅ Database persistence
- ✅ Loyverse integration (some)
- ✅ Event publishing/subscribing

**Example Structure (Product Service):**
```
services/product/
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── product.go           # Product entity
│   │   │   ├── pricing.go           # Pricing entity
│   │   │   └── availability.go      # Availability rules
│   │   └── repository/
│   │       ├── product.go           # Product repo interface
│   │       └── pricing.go           # Pricing repo interface
│   ├── application/
│   │   ├── product_usecase.go       # Product business logic
│   │   ├── pricing_usecase.go       # Pricing calculations
│   │   ├── availability_usecase.go  # Availability control
│   │   └── sync_usecase.go          # Loyverse sync logic
│   ├── infrastructure/
│   │   ├── database/
│   │   │   └── repository.go        # All repo implementations
│   │   ├── loyverse/
│   │   │   ├── client.go            # Loyverse API client
│   │   │   └── sync.go              # Sync implementation
│   │   └── cache/
│   │       └── redis.go             # Product caching
│   └── transport/
│       └── http/
│           ├── handler/
│           │   ├── product.go       # Product CRUD APIs
│           │   ├── pricing.go       # Pricing APIs
│           │   └── sync.go          # Sync APIs
│           └── middleware/
```

### **Type 2: Integration Services**
**Services:** AI, Analytics, Notification, Reporting, Procurement

**Features:**
- ✅ Lightweight entities
- ✅ Data processing logic
- ✅ External API integration
- ✅ Event consumption heavy
- ✅ Minimal database storage

**Example Structure (Analytics Service):**
```
services/analytics/
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── metrics.go           # Metrics entities
│   │   │   └── report.go            # Report entities
│   │   └── repository/
│   │       └── analytics.go         # Analytics repo interface
│   ├── application/
│   │   ├── metrics_usecase.go       # Metrics calculation
│   │   ├── reporting_usecase.go     # Report generation
│   │   └── dashboard_usecase.go     # Dashboard logic
│   ├── infrastructure/
│   │   ├── database/
│   │   │   └── repository.go        # Analytics data storage
│   │   ├── events/
│   │   │   ├── consumer.go          # Kafka event consumer
│   │   │   └── processor.go         # Event processing
│   │   └── external/
│   │       └── ai_client.go         # AI service integration
│   └── transport/
│       └── http/
│           ├── handler/
│           │   ├── metrics.go       # Metrics APIs
│           │   ├── reports.go       # Report APIs
│           │   └── dashboard.go     # Dashboard APIs
```

### **Type 3: Webhook Services**
**Services:** Chat Webhook, Loyverse Webhook, Delivery Webhook, Payment Webhook

**Features:**
- ✅ Minimal domain logic
- ✅ Webhook validation & forwarding
- ✅ No database (usually)
- ✅ Event publishing only
- ✅ Lightweight structure

**Example Structure (Chat Webhook):**
```
services/chat-webhook/
├── internal/
│   ├── domain/
│   │   └── entity/
│   │       └── webhook.go           # Webhook event entity
│   ├── application/
│   │   ├── webhook_usecase.go       # Webhook processing logic
│   │   └── forwarder_usecase.go     # Message forwarding logic
│   ├── infrastructure/
│   │   ├── external/
│   │   │   └── chat_client.go       # Chat service client
│   │   └── events/
│   │       └── publisher.go         # Event publishing
│   └── transport/
│       └── http/
│           ├── handler/
│           │   ├── line.go          # LINE webhook handler
│           │   └── facebook.go      # Facebook webhook handler
│           └── middleware/
│               └── validation.go    # Webhook validation
```

### **Type 4: Support Services**
**Services:** User, CDN

**Features:**
- ✅ Simple CRUD operations
- ✅ Authentication/Authorization
- ✅ File management
- ✅ Standard database operations

**Example Structure (User Service):**
```
services/user/
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── user.go              # User entity
│   │   │   └── permission.go        # Permission entity
│   │   └── repository/
│   │       └── user.go              # User repo interface
│   ├── application/
│   │   ├── auth_usecase.go          # Authentication logic
│   │   ├── user_usecase.go          # User management
│   │   └── permission_usecase.go    # Permission management
│   ├── infrastructure/
│   │   ├── database/
│   │   │   └── repository.go        # User repo implementation
│   │   ├── auth/
│   │   │   ├── jwt.go               # JWT handling
│   │   │   └── hash.go              # Password hashing
│   │   └── cache/
│   │       └── session.go           # Session caching
│   └── transport/
│       └── http/
│           ├── handler/
│           │   ├── auth.go          # Auth endpoints
│           │   └── user.go          # User CRUD endpoints
│           └── middleware/
│               └── auth.go          # Auth middleware
```

---

## 🔧 **Implementation Guidelines**

### **1. 📦 Domain Layer Rules**
```go
// ✅ DO: Pure business logic, no external dependencies
type Product struct {
    ID           string
    Name         string
    Price        decimal.Decimal
    IsAvailable  bool
}

func (p *Product) ApplyDiscount(percentage float64) error {
    if percentage < 0 || percentage > 100 {
        return errors.New("invalid discount percentage")
    }
    // Business logic here
    return nil
}

// ❌ DON'T: No database, HTTP, or external service dependencies
// import "database/sql"  // ❌
// import "net/http"      // ❌
```

### **2. 📋 Application Layer Rules**
```go
// ✅ DO: Orchestrate domain logic, use repository interfaces
type ProductUsecase struct {
    productRepo domain.ProductRepository  // Interface only
    cache       cache.Cache               // Interface only
    eventBus    events.Publisher          // Interface only
}

func (uc *ProductUsecase) CreateProduct(ctx context.Context, req CreateProductRequest) error {
    // 1. Validation
    if err := req.Validate(); err != nil {
        return err
    }
    
    // 2. Business logic
    product := domain.NewProduct(req.Name, req.Price)
    
    // 3. Persistence
    if err := uc.productRepo.Create(ctx, product); err != nil {
        return err
    }
    
    // 4. Side effects
    uc.eventBus.Publish("product.created", ProductCreatedEvent{ProductID: product.ID})
    
    return nil
}

// ❌ DON'T: No direct database or HTTP calls
```

### **3. 🔧 Infrastructure Layer Rules**
```go
// ✅ DO: Implement domain interfaces, handle external systems
type productRepository struct {
    db *sql.DB
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
    query := `INSERT INTO products (id, name, price) VALUES ($1, $2, $3)`
    _, err := r.db.ExecContext(ctx, query, product.ID, product.Name, product.Price)
    return err
}

// ✅ DO: Loyverse integration in infrastructure
type LoyverseSyncService struct {
    client      *LoyverseClient
    productRepo domain.ProductRepository
}

func (s *LoyverseSyncService) SyncProducts(ctx context.Context) error {
    // External API call + domain repository usage
}
```

### **4. 🌐 Transport Layer Rules**
```go
// ✅ DO: Handle HTTP concerns, delegate to application layer
type ProductHandler struct {
    productUsecase application.ProductUsecase
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
    var req CreateProductRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    if err := h.productUsecase.CreateProduct(r.Context(), req); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
}

// ❌ DON'T: No business logic in handlers
```

---

## 📊 **Service-Specific Variations**

### **Core Services (Product, Order, Customer, etc.)**
```
✅ Full domain entities
✅ Complex business logic
✅ Database persistence
✅ Loyverse integration (if applicable)
✅ Event publishing & subscribing
✅ Redis caching
✅ Complete CRUD APIs
```

### **Integration Services (Analytics, AI, etc.)**
```
✅ Lightweight entities
✅ Data processing focus
✅ Heavy event consumption
✅ External API integration
✅ Minimal database usage
✅ Reporting/Dashboard APIs
```

### **Webhook Services**
```
✅ Minimal domain logic
✅ Webhook validation
✅ Message forwarding
✅ Event publishing only
✅ No database (usually)
✅ Lightweight structure
```

### **Support Services (User, CDN)**
```
✅ Standard CRUD operations
✅ Authentication/Authorization
✅ File management
✅ Simple business logic
✅ Database operations
✅ Utility APIs
```

---

## 🚀 **Development Workflow**

### **1. Create New Service**
```bash
# Copy template structure
cp -r services/_template services/new-service

# Update go.mod
cd services/new-service
go mod init new-service

# Update service-specific code
# - Domain entities
# - Repository interfaces
# - Use cases
# - Infrastructure implementations
# - HTTP handlers
```

### **2. Standard Dependencies**
```go
// Common dependencies for all services
require (
    github.com/gorilla/mux v1.8.0          // HTTP router
    github.com/lib/pq v1.10.7              // PostgreSQL driver
    github.com/go-redis/redis/v8 v8.11.5   // Redis client
    github.com/segmentio/kafka-go v0.4.38  // Kafka client
    github.com/golang-migrate/migrate/v4   // Database migrations
    github.com/google/uuid v1.3.0          // UUID generation
    github.com/shopspring/decimal v1.3.1   // Decimal numbers
)
```

### **3. Docker Integration**
```dockerfile
# Standard Dockerfile for all services
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/main .

EXPOSE {SERVICE_PORT}
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:{SERVICE_PORT}/health || exit 1

CMD ["./main"]
```

---

## 📋 **Service Checklist**

### **Required for All Services:**
- [ ] Standard directory structure
- [ ] Domain entities with business logic
- [ ] Repository interfaces in domain
- [ ] Use cases in application layer
- [ ] Infrastructure implementations
- [ ] HTTP handlers with proper error handling
- [ ] Health check endpoint (`/health`)
- [ ] Database migrations (if applicable)
- [ ] Docker configuration
- [ ] Environment variable configuration
- [ ] Proper logging
- [ ] Unit tests for domain logic
- [ ] Integration tests for APIs

### **Additional for Core Services:**
- [ ] Loyverse integration (if applicable)
- [ ] Event publishing/subscribing
- [ ] Redis caching
- [ ] Master Data Protection (if syncing)
- [ ] Complex business logic
- [ ] Authentication middleware

### **Additional for Integration Services:**
- [ ] Event processing logic
- [ ] External API clients
- [ ] Data transformation logic
- [ ] Reporting capabilities

---

## 🎯 **Best Practices**

### **✅ DO's:**
- Use consistent naming conventions across all services
- Implement proper error handling and logging
- Follow Clean Architecture layers strictly
- Use interfaces for external dependencies
- Implement health checks and metrics
- Use environment variables for configuration
- Write unit tests for domain logic
- Use dependency injection
- Follow RESTful API conventions
- Implement proper validation

### **❌ DON'Ts:**
- Don't put business logic in handlers
- Don't access database directly from application layer
- Don't import infrastructure packages in domain
- Don't hardcode configuration values
- Don't skip error handling
- Don't mix different architectural patterns
- Don't create circular dependencies
- Don't ignore proper logging
- Don't skip database migrations
- Don't forget proper cleanup in tests

---

> 🏗️ **Consistent architecture across all SAAN services for maintainable, scalable microservices!**