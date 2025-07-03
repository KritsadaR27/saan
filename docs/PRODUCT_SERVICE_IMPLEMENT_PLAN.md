## 📚 **สรุปความเข้าใจจากเอกสารทั้งหมด**

### **1. MASTER_DATA_PROTECTION_PATTERN.md**
- **Core Pattern**: แยก fields เป็น SourceFields (จาก External) และ AdminFields (Admin เพิ่ม)
- **การ sync**: อัพเดตเฉพาะ SourceFields ไม่แตะ AdminFields
- **RelatedTables**: ตารางที่เกี่ยวข้องห้ามแตะ

### **2. PRODUCT_MANAGEMENT.md** 
- **Requirements**: Master Data Protection, Quantity-based pricing, VIP system
- **Database Schema**: Products table ที่แยก Loyverse fields และ Admin fields
- **API Structure**: Complete CRUD + Pricing + VIP access control

### **3. SNAPSHOT_STRATEGY.md**
- **ไม่ snapshot**: Product master data changes (เพราะเป็น master data)
- **ควร snapshot**: Business events เช่น pricing changes, availability changes

### **4. PROJECT_RULES.md**
- **Product Service**: port 8083, container `product`, service `product`
- **Communication**: Direct Call สำหรับ master data operations
- **Loyverse sync → Product Service**: Direct Call pattern

### **5. SAAN_FLOW.MD**
- **Architecture**: แสดง Product Service (8083) เป็น microservice แยก
- **Integration**: Order Service, Inventory Service, Chat AI ต้องการ Product data
- **Current Reality**: ไม่มี Product Service จริง, อยู่ใน Loyverse Integration

## ⚠️ **ความขัดแย้งและปัญหาที่พบ**

### **🚨 Critical Issues:**

1. **Service ไม่มีอยู่จริง**:
   - PROJECT_RULES.md: Product Service (8083) 
   - Reality: ไม่มี `services/product/` directory
   - Docker-compose: ไม่มี product service

2. **Architecture Confusion**:
   - เอกสารบอก Product Service แยก
   - โค้ดจริง: Product data อยู่ใน Loyverse Integration
   - Services อื่นยัง reference `http://product:8083`

3. **Data Flow Inconsistency**:
   ```
   Documents say: Loyverse → Product Service → Other Services
   Reality:       Loyverse Integration (8100) → Redis/Kafka → Other Services
   ```

## 🎯 **แผนการสร้าง Product Service แบบสมบูรณ์**

### **Phase 1: Architecture Decision & Setup**

#### **1.1 ตัดสินใจ Architecture Pattern**
เลือก **Option 1: Create Separate Product Service** เพราะ:
- ✅ สอดคล้องกับ PROJECT_RULES.md และ SAAN_FLOW.MD
- ✅ รองรับ PRODUCT_REQUIREMENTS.md ได้เต็มที่
- ✅ Clear separation of concerns
- ✅ Master Data Protection Pattern ใช้ได้เต็มที่

#### **1.2 Data Flow Design**
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Loyverse API    │───▶│ Loyverse        │───▶│ Product Service │
│                 │    │ Integration     │    │ (8083)          │
│                 │    │ (8100)          │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                 │                      │
                                 ▼                      ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │ Redis Cache     │    │ PostgreSQL      │
                       │                 │    │ products table  │
                       └─────────────────┘    └─────────────────┘
                                                      │
                       ┌─────────────────────────────┴─────────────────┐
                       │                                               │
                       ▼                    ▼                          ▼
            ┌─────────────────┐  ┌─────────────────┐      ┌─────────────────┐
            │ Order Service   │  │ Inventory       │      │ Chat AI         │
            │ (8081)          │  │ Service (8082)  │      │ (8090)          │
            └─────────────────┘  └─────────────────┘      └─────────────────┘
```

### **Phase 2: Database Design & Migration**

#### **2.1 Enhanced Products Table**
```sql
-- สร้าง products table ตาม PRODUCT_MANAGEMENT.md
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- ✅ Loyverse-controlled fields (sync updates these)
    external_id VARCHAR(100) UNIQUE,        -- Loyverse ID
    source_system VARCHAR(50) DEFAULT 'loyverse',
    name VARCHAR(200) NOT NULL,
    description TEXT,
    sku VARCHAR(100),
    barcode VARCHAR(100),
    category_id UUID,
    supplier_id UUID,
    cost_price DECIMAL(10,2),
    selling_price DECIMAL(10,2),
    status VARCHAR(20) DEFAULT 'active',
    last_sync_from_loyverse TIMESTAMP,
    
    -- 🔒 Admin-controlled fields (sync never touches these)
    display_name VARCHAR(200),
    internal_category VARCHAR(100),
    internal_notes TEXT,
    is_featured BOOLEAN DEFAULT false,
    profit_margin_target DECIMAL(5,2),
    sales_tags JSONB,
    
    -- Product Specifications
    weight_grams DECIMAL(8,2),
    units_per_pack INT DEFAULT 1,
    unit_type VARCHAR(20) DEFAULT 'piece',
    
    -- Advanced Availability Control
    is_admin_active BOOLEAN DEFAULT true,
    inactive_reason VARCHAR(200),
    inactive_until TIMESTAMP,
    auto_reactivate BOOLEAN DEFAULT false,
    inactive_schedule JSONB,
    
    -- VIP Access Control
    vip_only BOOLEAN DEFAULT false,
    min_vip_level VARCHAR(20),
    vip_early_access BOOLEAN DEFAULT false,
    early_access_until TIMESTAMP,
    
    -- System fields
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

#### **2.2 Pricing Tables**
```sql
-- Product pricing tiers สำหรับ quantity-based pricing
CREATE TABLE product_pricing_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id),
    min_quantity INT NOT NULL,
    max_quantity INT,
    price DECIMAL(10,2) NOT NULL,
    tier_name VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    valid_from DATE DEFAULT CURRENT_DATE,
    valid_until DATE
);

-- Customer group pricing
CREATE TABLE customer_group_pricing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id),
    customer_group VARCHAR(50) NOT NULL,
    base_price DECIMAL(10,2),
    discount_percentage DECIMAL(5,2),
    is_active BOOLEAN DEFAULT true
);

-- VIP pricing benefits
CREATE TABLE vip_pricing_benefits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vip_level VARCHAR(20) NOT NULL,
    global_discount_percentage DECIMAL(5,2),
    quantity_multiplier DECIMAL(5,2) DEFAULT 1.0,
    is_active BOOLEAN DEFAULT true
);
```

### **Phase 3: Product Service Implementation**

#### **3.1 Clean Architecture Structure**
```
services/product/
├── cmd/
│   └── main.go                    # Entry point
├── internal/
│   ├── domain/                    # Domain entities & interfaces
│   │   ├── product.go            # Product entity
│   │   ├── pricing.go            # Pricing domain logic
│   │   ├── availability.go       # Availability rules
│   │   └── repository.go         # Repository interfaces
│   ├── application/               # Use cases
│   │   ├── product_service.go    # Core product operations
│   │   ├── pricing_service.go    # Pricing calculations
│   │   ├── sync_service.go       # Loyverse sync logic
│   │   └── availability_service.go # Availability control
│   ├── infrastructure/            # External adapters
│   │   ├── database/
│   │   │   ├── postgres.go       # PostgreSQL connection
│   │   │   └── repository.go     # Repository implementation
│   │   ├── cache/
│   │   │   └── redis.go          # Redis caching
│   │   ├── events/
│   │   │   └── kafka.go          # Event publishing
│   │   └── loyverse/
│   │       └── sync_client.go    # Loyverse sync client
│   └── transport/                 # Input adapters
│       └── http/
│           ├── product_handler.go # Product APIs
│           ├── pricing_handler.go # Pricing APIs
│           ├── sync_handler.go    # Sync APIs
│           └── routes.go          # Route definitions
├── migrations/                    # Database migrations
├── Dockerfile                     # Container definition
└── go.mod                        # Go dependencies
```

#### **3.2 Master Data Protection Implementation**
```go
// Field Policy Definition
type ProductFieldPolicy struct {
    // ✅ Loyverse-controlled (sync updates these)
    SourceFields []string
    
    // 🔒 Admin-controlled (sync never touches)
    AdminFields []string
    
    // 🔒 Related tables (sync never touches)
    RelatedTables []string
}

// Smart Upsert with Field Protection
func (s *ProductService) UpsertFromLoyverse(ctx context.Context, req *LoyverseSyncRequest) error {
    existing, err := s.repo.GetByExternalID(ctx, req.LoyverseID)
    if err != nil && !errors.Is(err, sql.ErrNoRows) {
        return err
    }
    
    if existing == nil {
        return s.createFromLoyverse(ctx, req)
    }
    
    // อัพเดตเฉพาะ SourceFields เท่านั้น
    return s.updateLoyverseFields(ctx, existing.ID, req)
}
```

### **Phase 4: API Design & Implementation**

#### **4.1 Product Management APIs**
```go
// Core Product APIs
GET    /api/v1/products                 # List products with filters
GET    /api/v1/products/{id}            # Get product details
POST   /api/v1/products                 # Create product (admin only)
PUT    /api/v1/products/{id}            # Update admin fields only
DELETE /api/v1/products/{id}            # Soft delete

// Availability Control
GET    /api/v1/products/{id}/availability    # Check availability
POST   /api/v1/products/{id}/availability    # Update availability settings
POST   /api/v1/products/{id}/schedule        # Set schedule-based availability

// Search & Discovery
GET    /api/v1/products/search?q={query}     # Search products
GET    /api/v1/products/featured             # Get featured products
GET    /api/v1/products/categories/{cat}     # Products by category
```

#### **4.2 Pricing APIs**
```go
// Pricing Calculation
GET    /api/v1/products/{id}/pricing?quantity=10&customer_group=wholesale&vip_level=gold
GET    /api/v1/products/{id}/pricing-tiers   # Get pricing tiers
POST   /api/v1/products/{id}/pricing-tiers   # Create pricing tier
PUT    /api/v1/products/{id}/pricing-tiers/{tier_id}  # Update tier
DELETE /api/v1/products/{id}/pricing-tiers/{tier_id}  # Delete tier

// Customer Group Pricing
GET    /api/v1/customer-groups/{group}/pricing       # Group pricing rules
POST   /api/v1/customer-groups/{group}/pricing       # Set group pricing

// VIP System
GET    /api/v1/vip/{level}/benefits                  # VIP benefits
POST   /api/v1/vip/{level}/benefits                  # Set VIP benefits
GET    /api/v1/products/{id}/vip-access?customer_id={id}  # Check VIP access
```

#### **4.3 Sync Management APIs**
```go
// Loyverse Integration
POST   /api/v1/sync/loyverse/products        # Manual sync trigger
GET    /api/v1/sync/loyverse/status          # Sync status
GET    /api/v1/sync/loyverse/logs            # Sync history
POST   /api/v1/sync/loyverse/products/{id}   # Sync specific product

// Field Protection
GET    /api/v1/sync/field-policy             # Get field policy
PUT    /api/v1/sync/field-policy             # Update field policy
```

### **Phase 5: Integration & Communication**

#### **5.1 Integration with Loyverse**
```go
// Loyverse Integration Service เปลี่ยนจาก:
loyverseService.PublishToKafka(productEvent)

// เป็น:
productService.SyncFromLoyverse(loyverseData)
```

#### **5.2 Integration with Other Services**
```go
// Order Service Integration
func (o *OrderService) ValidateOrderItems(items []OrderItem) error {
    for _, item := range items {
        // Check product availability
        available, reason := o.productClient.IsProductAvailable(item.ProductID)
        if !available {
            return fmt.Errorf("product unavailable: %s", reason)
        }
        
        // Get pricing
        pricing, err := o.productClient.GetProductPricing(PricingRequest{
            ProductID: item.ProductID,
            Quantity: item.Quantity,
            CustomerGroup: customer.Group,
            VIPLevel: customer.VIPLevel,
        })
        if err != nil {
            return err
        }
        
        item.UnitPrice = pricing.FinalPrice
    }
    return nil
}

// Chat AI Integration
func (c *ChatService) SearchProducts(query string) ([]Product, error) {
    return c.productClient.SearchProducts(SearchRequest{
        Query: query,
        CustomerID: c.customerID,
        IncludeVIPOnly: c.customer.IsVIP(),
    })
}
```

### **Phase 6: Advanced Features Implementation**

#### **6.1 Availability Control System**
```go
func (s *AvailabilityService) IsProductAvailable(productID uuid.UUID, customerID uuid.UUID) (bool, string) {
    product := s.GetProduct(productID)
    customer := s.GetCustomer(customerID)
    
    // 1. Check Loyverse status
    if product.Status != "active" {
        return false, "Product inactive in Loyverse"
    }
    
    // 2. Check admin override
    if !product.IsAdminActive {
        return false, product.InactiveReason
    }
    
    // 3. Check time-based inactive
    if product.InactiveUntil != nil && time.Now().Before(*product.InactiveUntil) {
        return false, "Temporarily unavailable"
    }
    
    // 4. Check VIP access
    if product.VIPOnly && !customer.IsVIP() {
        return false, "VIP members only"
    }
    
    // 5. Check schedule-based availability
    if product.InactiveSchedule != nil {
        return s.checkScheduleAvailability(product.InactiveSchedule)
    }
    
    return true, ""
}
```

#### **6.2 Pricing Calculation Engine**
```go
func (s *PricingService) CalculatePrice(req PricingRequest) (*PricingResult, error) {
    // 1. Get base price from Loyverse
    basePrice := s.getBasePrice(req.ProductID)
    
    // 2. Find quantity-based pricing tier
    tier := s.findPricingTier(req.ProductID, req.Quantity)
    
    // 3. Apply customer group discount
    groupDiscount := s.getCustomerGroupDiscount(req.ProductID, req.CustomerGroup)
    
    // 4. Apply VIP benefits
    vipDiscount := 0.0
    if req.VIPLevel != "" {
        vipBenefits := s.getVIPBenefits(req.VIPLevel)
        
        // Global VIP discount
        vipDiscount = tier.Price * vipBenefits.GlobalDiscountPercentage / 100
        
        // VIP quantity multiplier for better tiers
        if vipBenefits.QuantityMultiplier > 1.0 {
            effectiveQuantity := int(float64(req.Quantity) * vipBenefits.QuantityMultiplier)
            betterTier := s.findPricingTier(req.ProductID, effectiveQuantity)
            if betterTier.Price < tier.Price {
                tier = betterTier
            }
        }
    }
    
    // 5. Calculate final price
    finalPrice := tier.Price - groupDiscount - vipDiscount
    
    return &PricingResult{
        BasePrice:    basePrice,
        TierPrice:    tier.Price,
        GroupDiscount: groupDiscount,
        VIPDiscount:  vipDiscount,
        FinalPrice:   finalPrice,
        TotalPrice:   finalPrice * float64(req.Quantity),
        Savings:      (basePrice - finalPrice) * float64(req.Quantity),
    }, nil
}
```

### **Phase 7: Docker & Deployment**

#### **7.1 Dockerfile**
```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget tzdata
ENV TZ=Asia/Bangkok

WORKDIR /root/
COPY --from=builder /app/main .

RUN adduser -D -s /bin/sh appuser
USER appuser

EXPOSE 8083

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8083/health || exit 1

CMD ["./main"]
```

#### **7.2 Docker Compose Integration**
```yaml
# เพิ่มใน docker-compose.yml
product:
  build:
    context: ./services/product
    dockerfile: Dockerfile
  container_name: product
  environment:
    - DB_HOST=postgres
    - DB_PORT=5432
    - DB_USER=saan
    - DB_PASSWORD=saan_password
    - DB_NAME=saan_db
    - REDIS_ADDR=redis:6379
    - KAFKA_BROKERS=kafka:9092
    - LOYVERSE_SERVICE_URL=http://loyverse:8100
  ports:
    - "8083:8083"
  depends_on:
    postgres:
      condition: service_healthy
    redis:
      condition: service_healthy
  networks:
    - saan-network
  healthcheck:
    test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8083/health"]
    interval: 30s
    timeout: 10s
    retries: 3
```

## 📋 **Implementation Checklist**

### **Phase 1: Foundation** 
- [ ] สร้าง services/product directory structure
- [ ] Setup Go project (go.mod, dependencies)
- [ ] สร้าง database migrations
- [ ] Implement basic domain entities

### **Phase 2: Core Features**
- [ ] Implement Master Data Protection pattern
- [ ] Create Product CRUD operations
- [ ] Setup Loyverse sync integration
- [ ] Implement basic pricing system

### **Phase 3: Advanced Features**
- [ ] Quantity-based pricing tiers
- [ ] VIP access control system
- [ ] Availability control system
- [ ] Admin override features

### **Phase 4: Integration**
- [ ] เพิ่ม product service ใน docker-compose
- [ ] Update other services to use product APIs
- [ ] Update Loyverse integration to sync via Product Service
- [ ] Test complete data flow

### **Phase 5: Production Ready**
- [ ] Add comprehensive testing
- [ ] Implement monitoring & logging
- [ ] Add rate limiting & security
- [ ] Performance optimization

---

**🎯 Product Service จะเป็น central hub สำหรับ product data ที่สมบูรณ์แบบ รองรับทุก requirement และสอดคล้องกับ architecture design ทั้งหมด!**

คุณต้องการให้เริ่มจาก Phase ไหนก่อน?
