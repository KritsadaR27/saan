# 🔧 SAAN Order Service - Refactoring Plan

## 📋 Current Status Analysis

จากการวิเคราะห์ Order Service ในระบบ SAAN พบว่าโครงสร้างปัจจุบันมีความสอดคล้องกับ Clean Architecture แล้วในระดับหนึ่ง แต่ยังมีจุดที่ควรปรับปรุงให้ตรงกับมาตรฐาน SAAN และ PROJECT_RULES.md

## 🎯 สถานะปัจจุบันของ Order Service

### ✅ ส่วนที่ดีแล้ว
- Clean Architecture layers แยกชัดเจน
- Domain-driven design ถูกต้อง
- Repository pattern implementation
- Event-driven architecture with outbox pattern
- Service-to-service communication ใช้ service names (ตาม PROJECT_RULES.md)
- RBAC middleware พร้อม JWT authentication
- Comprehensive test coverage

### ⚠️ ส่วนที่ต้องปรับปรุง
- `application` layer ควรมีแต่ `usecase` ตามมาตรฐาน
- Database schema ต้องเพิ่ม snapshot support
- ยังไม่มี Redis integration สำหรับ caching
- Event publishing ยังใช้ mock (ต้องใช้ Kafka จริง)

## 📊 การออกแบบ Database Schema

### Core Tables (มีอยู่แล้ว - ใช้ได้)

```sql
-- Orders table (✅ Complete)
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL,
    code VARCHAR(50) UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    source VARCHAR(20) DEFAULT 'online',
    paid_status VARCHAR(20) DEFAULT 'unpaid',
    total_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    discount DECIMAL(10,2) DEFAULT 0,
    shipping_fee DECIMAL(10,2) DEFAULT 0,
    tax DECIMAL(10,2) DEFAULT 0,
    tax_enabled BOOLEAN DEFAULT true,
    shipping_address TEXT NOT NULL,
    billing_address TEXT NOT NULL,
    payment_method VARCHAR(50),
    promo_code VARCHAR(50),
    notes TEXT,
    confirmed_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    cancelled_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Order Items table (✅ Complete with stock override)
CREATE TABLE order_items (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10,2) NOT NULL CHECK (unit_price >= 0),
    total_price DECIMAL(10,2) NOT NULL CHECK (total_price >= 0),
    is_override BOOLEAN NOT NULL DEFAULT FALSE,
    override_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Audit Log table (✅ Complete)
CREATE TABLE order_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    user_id VARCHAR(255),
    action VARCHAR(50) NOT NULL,
    details JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_audit_order_id FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);

-- Events Outbox table (✅ Complete)
CREATE TABLE order_events_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    sent_at TIMESTAMP WITH TIME ZONE,
    retry_count INTEGER DEFAULT 0,
    CONSTRAINT fk_events_order_id FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);
```

### 🆕 Tables ที่ต้องเพิ่ม (ตาม SNAPSHOT_STRATEGY.md)

```sql
-- Order Snapshots table (ใหม่ - สำหรับ Order snapshot strategy)
CREATE TABLE order_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    snapshot_type VARCHAR(50) NOT NULL, -- 'created', 'confirmed', 'shipped', 'completed', 'cancelled'
    snapshot_data JSONB NOT NULL,
    previous_snapshot_id UUID REFERENCES order_snapshots(id),
    created_by_user_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Quick access fields (denormalized)
    order_status VARCHAR(50),
    total_amount DECIMAL(12,2),
    item_count INT,
    customer_id UUID,
    delivery_method VARCHAR(50),
    
    CONSTRAINT fk_snapshots_order_id FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);

-- Indexes for snapshots
CREATE INDEX idx_order_snapshots ON order_snapshots(order_id, created_at);
CREATE INDEX idx_snapshot_type ON order_snapshots(snapshot_type, created_at);
CREATE INDEX idx_customer_snapshots ON order_snapshots(customer_id, created_at);
```

## 🏗️ Folder Structure Refactoring

### 📁 ในเปลี่ยน `application` ให้มีแต่ `usecase`

```
services/order/
├── cmd/
│   └── main.go
├── internal/
│   ├── domain/                    # ✅ Keep as is
│   │   ├── order.go
│   │   ├── repository.go
│   │   ├── audit.go
│   │   └── errors.go
│   ├── application/                   # 🔄  
│   │   ├── order_usecase.go       # 🔄 RENAMED from order_service.go
│   │   ├── chat_order_usecase.go  # 🔄 RENAMED from chat_order_service.go
│   │   ├── stats_usecase.go       # 🔄 RENAMED from stats_service.go
│   │   ├── snapshot_usecase.go    # 🆕 NEW - snapshot operations
│   │   ├── dto/                   # ✅ Keep as is
│   │   │   ├── order_dto.go
│   │   │   └── stats_dto.go
│   │   └── template/              # ✅ Keep as is
│   │       ├── message_template.go
│   │       └── template_selector.go
│   ├── infrastructure/            # ✅ Keep as is
│   │   ├── client/
│   │   ├── config/
│   │   ├── db/
│   │   ├── event/
│   │   ├── repository/
│   │   └── cache/                 # 🆕 NEW - Redis integration
│   │       └── redis.go
│   └── transport/                 # ✅ Keep as is
│       └── http/
├── migrations/                    # ✅ Keep as is
├── Dockerfile                     # ✅ Keep as is
├── Makefile                       # ✅ Keep as is
└── go.mod                         # ✅ Keep as is
```

## 🔄 Service Communication Patterns

### 📞 Direct Call (HTTP/gRPC) - ✅ มีแล้ว

```go
// ✅ ใช้งานถูกต้องแล้ว - ตาม PROJECT_RULES.md
inventoryClient := client.NewHTTPInventoryClient("http://inventory-service:8082")
customerClient := client.NewHTTPCustomerClient("http://user-service:8088") 
notificationClient := client.NewHTTPNotificationClient("http://notification-service:8092")
```

**Use Cases:**
- Stock checks (ไม่ reserve stock)
- Customer validation
- Product information lookup
- Real-time data ที่ต้องการ immediate response

### 📨 Event-Driven (Kafka) - 🔄 ต้องปรับ

```go
// 🔄 ปัจจุบันใช้ MockEventPublisher ต้องเปลี่ยนเป็น Kafka จริง
type KafkaEventPublisher struct {
    client kafka.Writer
}

func (p *KafkaEventPublisher) PublishEvent(ctx context.Context, event *domain.OrderEvent) error {
    message := kafka.Message{
        Topic: "order-events",
        Key:   []byte(event.OrderID.String()),
        Value: eventJSON,
    }
    return p.client.WriteMessages(ctx, message)
}
```

**Events ที่ต้อง Publish:**
- `order.created` → [Customer, Inventory, Analytics]
- `order.confirmed` → [Finance, Inventory, Customer, Analytics]
- `order.completed` → [Finance, Inventory, Customer, Analytics]
- `order.cancelled` → [Finance, Inventory, Customer]
- `payment.confirmed` → [Order, Finance, Notification]

### 🗄️ Redis Cache - 🆕 ต้องเพิ่ม

```go
// 🆕 เพิ่ม Redis client ใหม่
type RedisCache struct {
    client *redis.Client
}

// Cache patterns ตาม PROJECT_RULES.md
func (c *RedisCache) CacheOrder(orderID uuid.UUID, order *domain.Order) error {
    key := fmt.Sprintf("order:%s", orderID)
    data, _ := json.Marshal(order)
    return c.client.Set(key, data, 1*time.Hour).Err()
}

func (c *RedisCache) GetCachedOrder(orderID uuid.UUID) (*domain.Order, error) {
    key := fmt.Sprintf("order:%s", orderID)
    data, err := c.client.Get(key).Result()
    if err != nil {
        return nil, err
    }
    var order domain.Order
    json.Unmarshal([]byte(data), &order)
    return &order, nil
}
```

**Redis Usage Patterns:**
- `order:draft:{order_id}` → Order being built (no stock reservation)
- `order:pricing:{order_id}` → Calculated pricing cache
- `checkout:validation:{customer_id}` → Final stock check before order
- `user:session:{session_id}` → User sessions for order context

## 📸 Snapshot Integration

### 🆕 Snapshot Usecase

```go
// 🆕 internal/usecase/snapshot_usecase.go
type SnapshotUsecase struct {
    snapshotRepo domain.SnapshotRepository
    orderRepo    domain.OrderRepository
    logger       logger.Logger
}

func (uc *SnapshotUsecase) CreateOrderSnapshot(ctx context.Context, orderID uuid.UUID, snapshotType string) error {
    // Get current order state
    order, err := uc.orderRepo.GetByID(ctx, orderID)
    if err != nil {
        return err
    }
    
    // Create snapshot
    snapshot := &domain.OrderSnapshot{
        ID:           uuid.New(),
        OrderID:      orderID,
        SnapshotType: snapshotType,
        SnapshotData: order.ToSnapshotData(),
        OrderStatus:  string(order.Status),
        TotalAmount:  order.TotalAmount,
        ItemCount:    len(order.Items),
        CustomerID:   order.CustomerID,
        CreatedAt:    time.Now(),
    }
    
    return uc.snapshotRepo.Create(ctx, snapshot)
}
```

### 📊 Snapshot Triggers

```go
// Integration ใน order_usecase.go
func (uc *OrderUsecase) UpdateOrderStatus(ctx context.Context, id uuid.UUID, req *dto.UpdateOrderStatusRequest) (*dto.OrderResponse, error) {
    // Update order
    order, err := uc.updateOrder(ctx, id, req)
    if err != nil {
        return nil, err
    }
    
    // Create snapshot on status change
    snapshotType := uc.determineSnapshotType(req.Status)
    if snapshotType != "" {
        err = uc.snapshotUsecase.CreateOrderSnapshot(ctx, id, snapshotType)
        if err != nil {
            uc.logger.Error("Failed to create snapshot", "error", err)
            // Don't fail order update for snapshot failure
        }
    }
    
    return dto.ToOrderResponse(order), nil
}
```

## 🔧 Implementation Steps

### Phase 1: Structure Refactoring

1. **Rename `application` → `usecase`**
   ```bash
   cd services/order/internal
   mv application usecase
   
   # Update all import paths
   find . -name "*.go" -exec sed -i 's/application/usecase/g' {} \;
   ```

2. **Rename service files**
   ```bash
   cd usecase/
   mv order_service.go order_usecase.go
   mv chat_order_service.go chat_order_usecase.go
   mv stats_service.go stats_usecase.go
   ```

3. **Update interface names**
   ```go
   // From: OrderService
   // To:   OrderUsecase
   type OrderUsecase struct {
       orderRepo      domain.OrderRepository
       orderItemRepo  domain.OrderItemRepository
       // ... other dependencies
   }
   
   func NewOrderUsecase(...) *OrderUsecase {
       return &OrderUsecase{...}
   }
   ```

### Phase 2: Add Redis Integration

1. **Add Redis client**
   ```go
   // internal/infrastructure/cache/redis.go
   type RedisClient struct {
       client *redis.Client
   }
   
   func NewRedisClient(addr string) *RedisClient {
       rdb := redis.NewClient(&redis.Options{
           Addr: addr, // redis:6379 ตาม PROJECT_RULES.md
       })
       return &RedisClient{client: rdb}
   }
   ```

2. **Update main.go**
   ```go
   // cmd/main.go - Add Redis initialization
   redisClient := cache.NewRedisClient("redis:6379")
   defer redisClient.Close()
   ```

### Phase 3: Add Snapshot Support

1. **Create snapshot migration**
   ```sql
   -- migrations/004_add_snapshots.sql
   CREATE TABLE order_snapshots (
       -- ... as defined above
   );
   ```

2. **Add snapshot domain**
   ```go
   // internal/domain/snapshot.go
   type OrderSnapshot struct {
       ID           uuid.UUID
       OrderID      uuid.UUID
       SnapshotType string
       SnapshotData map[string]interface{}
       // ... other fields
   }
   ```

3. **Add snapshot repository**
   ```go
   // internal/domain/repository.go
   type SnapshotRepository interface {
       Create(ctx context.Context, snapshot *OrderSnapshot) error
       GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*OrderSnapshot, error)
   }
   ```

### Phase 4: Replace Mock Kafka

1. **Install Kafka client**
   ```bash
   go get github.com/segmentio/kafka-go
   ```

2. **Implement real Kafka publisher**
   ```go
   // internal/infrastructure/event/kafka_publisher.go
   type KafkaPublisher struct {
       writer *kafka.Writer
   }
   
   func NewKafkaPublisher() *KafkaPublisher {
       return &KafkaPublisher{
           writer: &kafka.Writer{
               Addr:     kafka.TCP("kafka:9092"), // ตาม PROJECT_RULES.md
               Topic:    "order-events",
               Balancer: &kafka.LeastBytes{},
           },
       }
   }
   ```

### Phase 5: Update Tests

1. **Update test imports**
   ```go
   // All test files
   import (
       "github.com/saan/order-service/internal/usecase" // Changed from application
   )
   ```

2. **Update mock service names**
   ```go
   // Change all references from *Service to *Usecase
   orderUsecase := NewOrderUsecase(...)
   ```

## 🔍 File & Folder Responsibilities

### 📁 `/cmd/`
- **main.go**: Entry point, dependency injection, service startup
- **debug.go**: Development debugging utilities

### 📁 `/internal/domain/`
- **order.go**: Order entity, business rules, status transitions
- **repository.go**: Repository interfaces (no implementations)
- **audit.go**: Audit log entities and event definitions
- **errors.go**: Domain-specific errors
- **snapshot.go**: 🆕 Snapshot entities

### 📁 `/internal/usecase/` (เดิม `application/`)
- **order_usecase.go**: Core order business logic
- **chat_order_usecase.go**: Chat-based order operations
- **stats_usecase.go**: Order statistics and analytics
- **snapshot_usecase.go**: 🆕 Snapshot operations
- **dto/**: Data transfer objects for API
- **template/**: Message templates for chat

### 📁 `/internal/infrastructure/`
- **client/**: HTTP clients for external services
- **config/**: Configuration loading
- **db/**: Database connection
- **event/**: Event publishing (Kafka)
- **repository/**: Repository implementations
- **cache/**: 🆕 Redis caching layer

### 📁 `/internal/transport/`
- **http/**: HTTP handlers, routes, middleware
- **middleware/**: RBAC authentication middleware

## 🚀 API Endpoints Summary

### Core Order Operations
```
POST   /api/v1/orders                    # Create order
GET    /api/v1/orders                    # List orders
GET    /api/v1/orders/:id                # Get order
PUT    /api/v1/orders/:id                # Update order
DELETE /api/v1/orders/:id                # Delete order
PATCH  /api/v1/orders/:id/status         # Update status
GET    /api/v1/orders/status/:status     # Get by status
POST   /api/v1/orders/:id/confirm-with-override # Stock override
```

### Customer Orders
```
GET    /api/v1/customers/:id/orders      # Customer's orders
```

### Chat Operations
```
POST   /api/v1/chat/orders               # Chat order creation
POST   /api/v1/chat/orders/:id/confirm   # Confirm chat order
POST   /api/v1/chat/orders/:id/cancel    # Cancel chat order
POST   /api/v1/chat/orders/:id/summary   # Generate summary
```

### Admin Operations
```
POST   /api/v1/admin/orders              # Admin create order
POST   /api/v1/admin/orders/:id/link-chat # Link to chat
POST   /api/v1/admin/orders/bulk-status  # Bulk status update
GET    /api/v1/admin/orders/export       # Export orders
```

### Statistics
```
GET    /api/v1/stats/daily               # Daily stats
GET    /api/v1/stats/monthly             # Monthly stats
GET    /api/v1/stats/top-products        # Top products
GET    /api/v1/stats/customer/:id        # Customer stats
GET    /api/v1/stats/overview            # Overview dashboard
```

### Health & Monitoring
```
GET    /health                           # Health check
```

## 🔐 RBAC Roles & Permissions

### Roles
- **sales**: Create/view orders, view customers
- **manager**: Sales permissions + update/confirm/cancel orders, stock override
- **admin**: Full access to all operations + bulk updates + export
- **ai_assistant**: Chat operations + draft orders + view-only access

### Protected Endpoints
- Order operations: Require sales/manager/admin
- Stock override: Require `orders:override_stock` permission
- Admin APIs: Require admin role + specific permissions
- Chat APIs: Require ai_assistant/manager/admin
- Statistics: Require manager/admin

## 📊 Redis Cache Patterns

### Order Operations
```go
// Draft orders (no stock reservation)
redis.Set("order:draft:{order_id}", orderData, 30*time.Minute)

// Pricing cache
redis.Set("order:pricing:{order_id}", pricingData, 1*time.Hour)

// Final checkout validation
redis.Set("checkout:validation:{customer_id}", validationData, 5*time.Minute)
```

### User Sessions
```go
// JWT sessions
redis.Set("user:session:{session_id}", userData, 24*time.Hour)

// API rate limiting
redis.Incr("api:rate_limit:{user_id}")
redis.Expire("api:rate_limit:{user_id}", 1*time.Hour)
```

### Analytics & Metrics
```go
// Real-time metrics
redis.Incr("metrics:daily:orders:{date}")
redis.Incr("metrics:daily:revenue:{date}")

// Dashboard cache
redis.Set("dashboard:stats:{date}", dashboardData, 1*time.Hour)
```

## 🎯 Next Steps Priority

1. **High Priority**
   - [ ] Rename `application` → `usecase` (Breaking change)
   - [ ] Add Redis integration
   - [ ] Replace mock Kafka with real implementation

2. **Medium Priority**
   - [ ] Add snapshot support
   - [ ] Improve error handling
   - [ ] Add more comprehensive tests

3. **Low Priority**
   - [ ] Performance optimization
   - [ ] Advanced analytics
   - [ ] Additional admin features

## ✅ Conclusion

Order Service มีโครงสร้าง Clean Architecture ที่ดีอยู่แล้ว แต่ต้องปรับให้ตรงกับมาตรฐาน SAAN:

1. **เปลี่ยน `application` → `usecase`** เพื่อให้สอดคล้องกับ Clean Architecture standard
2. **เพิ่ม Redis caching** ตาม PROJECT_RULES.md patterns
3. **เพิ่ม Snapshot support** ตาม SNAPSHOT_STRATEGY.md
4. **ใช้ Kafka จริง** แทน Mock publisher
5. **ปรับปรุง test coverage** หลังการ refactoring

การปรับปรุงเหล่านี้จะทำให้ Order Service มีความสอดคล้องกับมาตรฐาน SAAN ทั้งในด้าน architecture, performance และ maintainability