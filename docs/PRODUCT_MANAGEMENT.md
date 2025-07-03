# 🛍️ Product Management Requirements

## 📋 Overview

ระบบจัดการสินค้าแบบครบครัน รองรับการ sync จาก Loyverse, ข้อมูลเพิ่มเติมจาก Admin, และระบบ quantity-based pricing

## 🎯 Core Requirements

### 1. **Master Data Protection Pattern**
- ✅ Sync ข้อมูลพื้นฐานจาก Loyverse (ไม่แตะข้อมูล Admin)
- ✅ Admin เพิ่มข้อมูลเองได้โดยไม่หายระหว่าง sync
- ✅ แยกชัด field ไหนมาจาก source ไหน

### 2. **Product Specifications**
- ✅ น้ำหนัก, จำนวนต่อแพ็ค, หน่วย
- ✅ Product images และ gallery
- ✅ Internal categorization
- ✅ Marketing features (featured products, tags)

### 3. **Advanced Availability Control**
- ✅ Admin override (เปิด/ปิดสินค้า)
- ✅ Schedule-based availability (ปิดตามเวลา/วันที่)
- ✅ Temporary unavailability
- ✅ Automatic reactivation

### 4. **Quantity-based Pricing**
- ✅ Multiple pricing tiers ตาม quantity
- ✅ Customer group pricing
- ✅ Bulk order discounts
- ✅ Time-based pricing validity

### 5. **VIP Customer System**
- ✅ VIP-only products และ early access
- ✅ VIP pricing benefits (global discounts, quantity multipliers)
- ✅ Minimum VIP level requirements
- ✅ VIP point earning system

---

## 📊 Database Schema

### 1. **Enhanced Products Table**

```sql
-- Main products table (enhanced existing schema)
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
    brand VARCHAR(100),
    model VARCHAR(100),
    cost_price DECIMAL(10,2),               -- Base price from Loyverse
    selling_price DECIMAL(10,2),            -- Default selling price
    status VARCHAR(20) DEFAULT 'active',    -- Loyverse status
    last_sync_from_loyverse TIMESTAMP,
    
    -- 🔒 Admin-controlled fields (sync never touches these)
    display_name VARCHAR(200),              -- "โค้ก 325ml (โปรโมชั่น)"
    internal_category VARCHAR(100),         -- Internal categorization
    internal_notes TEXT,
    
    -- Marketing & Sales
    is_featured BOOLEAN DEFAULT false,
    profit_margin_target DECIMAL(5,2),
    sales_tags JSONB,                       -- ["popular", "cold_drink"]
    
    -- Product Specifications
    weight_grams DECIMAL(8,2),              -- น้ำหนักเป็นกรัม
    units_per_pack INT DEFAULT 1,           -- จำนวนต่อแพ็ค
    unit_type VARCHAR(20) DEFAULT 'piece',  -- piece, bottle, can, box, kg
    
    -- Advanced Availability Control
    is_admin_active BOOLEAN DEFAULT true,   -- Admin override
    inactive_reason VARCHAR(200),           -- เหตุผลที่ปิด
    inactive_until TIMESTAMP,               -- ปิดจนถึงวันที่
    auto_reactivate BOOLEAN DEFAULT false,  -- เปิดอัตโนมัติ
    inactive_schedule JSONB,                -- Schedule-based availability
    
    -- VIP Access Control
    vip_only BOOLEAN DEFAULT false,         -- VIP-only product
    min_vip_level VARCHAR(20),              -- 'gold', 'platinum', 'diamond'
    vip_early_access BOOLEAN DEFAULT false, -- Early access for VIP
    early_access_until TIMESTAMP,          -- Early access end time
    
    -- Existing fields (keep as-is)
    weight DECIMAL(8,2),                    -- Original weight field
    dimensions VARCHAR(100),
    wholesale_price DECIMAL(10,2),
    markup_percentage DECIMAL(5,2),
    unit_of_measure VARCHAR(20),
    min_stock_level INT DEFAULT 0,
    max_stock_level INT DEFAULT 1000,
    reorder_point INT DEFAULT 10,
    safety_stock INT DEFAULT 5,
    is_serialized BOOLEAN DEFAULT false,
    requires_expiry_tracking BOOLEAN DEFAULT false,
    primary_image_url TEXT,
    gallery_images JSONB,
    tags JSONB,
    
    -- System fields
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Indexes
    INDEX idx_products_external_id (external_id),
    INDEX idx_products_sku (sku),
    INDEX idx_products_category (category_id),
    INDEX idx_products_featured (is_featured),
    INDEX idx_products_admin_active (is_admin_active),
    INDEX idx_products_sync_time (last_sync_from_loyverse)
);
```

### 2. **Quantity-based Pricing Tables**

```sql
-- Product pricing tiers
CREATE TABLE product_pricing_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id) ON DELETE CASCADE,
    
    -- Quantity Range
    min_quantity INT NOT NULL,               -- 1, 10, 20
    max_quantity INT,                        -- 9, 19, NULL (unlimited)
    
    -- Pricing
    price DECIMAL(10,2) NOT NULL,            -- 325, 320, 310
    discount_percentage DECIMAL(5,2),        -- Alternative: % discount
    discount_amount DECIMAL(8,2),            -- Alternative: fixed discount
    
    -- Metadata
    tier_name VARCHAR(100),                  -- "Single", "Bulk 10", "Wholesale 20+"
    tier_description TEXT,
    
    -- Validity
    is_active BOOLEAN DEFAULT true,
    valid_from DATE DEFAULT CURRENT_DATE,
    valid_until DATE,
    
    -- Audit
    created_by_user_id UUID,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT valid_quantity_range CHECK (min_quantity > 0),
    CONSTRAINT valid_max_quantity CHECK (max_quantity IS NULL OR max_quantity >= min_quantity),
    CONSTRAINT unique_quantity_range UNIQUE (product_id, min_quantity),
    
    INDEX idx_pricing_tiers_product (product_id),
    INDEX idx_pricing_tiers_quantity (product_id, min_quantity)
);

-- VIP pricing benefits
CREATE TABLE vip_pricing_benefits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vip_level VARCHAR(20) NOT NULL,          -- 'gold', 'platinum', 'diamond'
    
    -- Pricing Benefits
    global_discount_percentage DECIMAL(5,2), -- 5%, 10%, 15%
    free_delivery_threshold DECIMAL(10,2),   -- Free delivery above X amount
    always_free_delivery BOOLEAN DEFAULT false,
    
    -- Quantity Benefits
    quantity_multiplier DECIMAL(5,2),        -- 1.2x (buy 10 get 12 pricing)
    bulk_tier_reduction INT,                  -- Access bulk pricing at lower qty
    
    -- Special Access
    early_access_hours INT,                   -- Hours before public for new products
    exclusive_products BOOLEAN DEFAULT false,
    priority_support BOOLEAN DEFAULT false,
    
    -- Point System
    points_multiplier DECIMAL(5,2) DEFAULT 1.0, -- Points earning multiplier
    welcome_bonus_points INT,                 -- Points when achieving tier
    monthly_bonus_points INT,                 -- Monthly bonus points
    birthday_bonus_points INT,               -- Birthday bonus
    
    -- Validity
    is_active BOOLEAN DEFAULT true,
    effective_from DATE DEFAULT CURRENT_DATE,
    effective_until DATE,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE (vip_level),
    INDEX idx_vip_benefits_level (vip_level, is_active)
);
```

---

## 🔄 Loyverse Sync Implementation

### 1. **Field Policy Definition**

```go
type ProductFieldPolicy struct {
    // ✅ Loyverse-controlled (sync updates these)
    SourceFields []string = []string{
        "external_id",              // Loyverse ID
        "source_system",            // "loyverse"
        "name",                     // Product name
        "description",              // Product description
        "sku",                      // SKU code
        "barcode",                  // Barcode
        "category_id",              // Category from Loyverse
        "cost_price",               // Cost price
        "selling_price",            // Base selling price
        "status",                   // active/inactive
        "last_sync_from_loyverse",  // Sync timestamp
    }
    
    // 🔒 Admin-controlled (sync never touches)
    AdminFields []string = []string{
        // Enhanced display
        "display_name",
        "internal_category", 
        "internal_notes",
        
        // Marketing fields
        "is_featured",
        "profit_margin_target",
        "sales_tags",
        "weight_grams",
        "units_per_pack",
        "unit_type",
        
        // Availability control
        "is_admin_active",
        "inactive_reason",
        "inactive_until",
        "auto_reactivate",
        "inactive_schedule",
        
        // VIP access control
        "vip_only",
        "min_vip_level", 
        "vip_early_access",
        "early_access_until",
        
        // Existing admin fields
        "min_stock_level",
        "max_stock_level",
        "reorder_point",
        "safety_stock",
        "markup_percentage",
        "wholesale_price",
        "weight",
        "dimensions",
        "unit_of_measure",
        "is_serialized",
        "requires_expiry_tracking",
        "primary_image_url",
        "gallery_images",
        "tags",
    }
}
```

### 2. **Sync Service Implementation**

```go
// Product Service - Smart Upsert
func (s *ProductService) UpsertFromLoyverse(ctx context.Context, req *ProductSyncRequest) error {
    existing, err := s.db.GetProductByExternalID(ctx, req.LoyverseID)
    if err != nil && !errors.Is(err, sql.ErrNoRows) {
        return err
    }
    
    if existing == nil {
        // Create new product with basic Loyverse data
        return s.createProductFromLoyverse(ctx, req)
    }
    
    // Update only Loyverse-controlled fields
    return s.updateProductLoyverseFields(ctx, existing.ID, req)
}

func (s *ProductService) updateProductLoyverseFields(ctx context.Context, productID uuid.UUID, req *ProductSyncRequest) error {
    query := `
        UPDATE products SET
            name = $2,                      -- ✅ Update from Loyverse
            description = $3,               -- ✅ Update from Loyverse
            sku = $4,                       -- ✅ Update from Loyverse
            barcode = $5,                   -- ✅ Update from Loyverse
            cost_price = $6,                -- ✅ Update from Loyverse
            selling_price = $7,             -- ✅ Update from Loyverse
            status = $8,                    -- ✅ Update from Loyverse
            last_sync_from_loyverse = NOW(),
            updated_at = NOW()
        WHERE id = $1
        -- ❌ Never touch: display_name, is_featured, weight_grams, 
        -- ❌ Never touch: pricing_tiers, admin settings, etc.
    `
    
    _, err := s.db.ExecContext(ctx, query,
        productID, req.Name, req.Description, req.SKU, req.Barcode,
        req.CostPrice, req.SellingPrice, req.Status,
    )
    
    return err
}
```

---

## 🎯 API Endpoints

### 1. **Product Management**

```go
// GET /api/products - List products with pricing
// GET /api/products/{id} - Get product details
// POST /api/products - Create product (admin)
// PUT /api/products/{id} - Update product (admin fields only)
// DELETE /api/products/{id} - Soft delete product

// GET /api/products/{id}/availability - Check current availability
// POST /api/products/{id}/availability - Update availability settings
```

### 2. **Pricing Management**

```go
// GET /api/products/{id}/pricing?quantity=10&customer_group=wholesale&vip_level=gold
// GET /api/products/{id}/pricing-tiers
// POST /api/products/{id}/pricing-tiers
// PUT /api/products/{id}/pricing-tiers/{tier_id}
// DELETE /api/products/{id}/pricing-tiers/{tier_id}

// VIP-specific endpoints
// GET /api/vip/{level}/benefits
// POST /api/vip/{level}/benefits
// GET /api/products/{id}/vip-access?customer_id={id}
```

### 3. **Sync Management**

```go
// POST /api/sync/loyverse/products - Manual sync trigger
// GET /api/sync/loyverse/status - Sync status
// GET /api/sync/loyverse/logs - Sync history
```

---

## 🔍 Business Logic Examples

### 1. **Availability Check**

```go
func (s *ProductService) IsProductAvailable(productID uuid.UUID) (bool, string) {
    product := s.GetProduct(productID)
    
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
    
    // 4. Check schedule-based availability
    if product.InactiveSchedule != nil {
        return s.checkScheduleAvailability(product.InactiveSchedule)
    }
    
    return true, ""
}
```

### 2. **VIP Product Access Control**

```go
func (s *ProductService) CanCustomerAccessProduct(ctx context.Context, customerID, productID string) (bool, string) {
    product := s.GetProduct(ctx, productID)
    customer := s.customerService.GetCustomer(ctx, customerID)
    
    // Check basic availability
    if available, reason := s.IsProductAvailable(productID); !available {
        return false, reason
    }
    
    // Check VIP-only products
    if product.VIPOnly {
        if customer.VIPLevel == "" || customer.VIPLevel == "bronze" {
            return false, "This product is available for VIP members only"
        }
    }
    
    // Check minimum VIP level requirement
    if product.MinVIPLevel != "" {
        if !s.hasMinimumVIPLevel(customer.VIPLevel, product.MinVIPLevel) {
            return false, fmt.Sprintf("Requires %s level or higher", product.MinVIPLevel)
        }
    }
    
    // Check VIP early access
    if product.VIPEarlyAccess && product.EarlyAccessUntil != nil {
        if time.Now().Before(*product.EarlyAccessUntil) {
            if customer.VIPLevel == "" || customer.VIPLevel == "bronze" {
                return false, "Early access for VIP members only"
            }
        }
    }
    
    return true, ""
}
```

---

## 🚀 Integration with Chat Service

### **Chat-to-Product Discovery**

```go
// Chat Service integration for product discovery
func (s *ChatService) HandleProductInquiry(message string, customerID string) (*ChatResponse, error) {
    // 1. Extract product intent from message
    productQuery := s.nlp.ExtractProductQuery(message)
    
    // 2. Search products via Product Service
    products, err := s.productService.SearchProducts(productQuery, customerID)
    if err != nil {
        return nil, err
    }
    
    // 3. Filter by customer VIP access
    accessibleProducts := []Product{}
    for _, product := range products {
        if canAccess, _ := s.productService.CanCustomerAccessProduct(customerID, product.ID); canAccess {
            accessibleProducts = append(accessibleProducts, product)
        }
    }
    
    // 4. Generate chat response with product suggestions
    return &ChatResponse{
        Type: "product_suggestions",
        Products: accessibleProducts,
        QuickReplies: []QuickReply{
            {Text: "เพิ่มลงตะกร้า", Action: "add_to_cart"},
            {Text: "ดูเพิ่มเติม", Action: "view_details"},
        },
    }, nil
}
```

---

## ✅ Implementation Checklist

### Phase 1: Basic Enhancement
- [ ] Run database migration
- [ ] Implement field policy
- [ ] Update sync service
- [ ] Test Loyverse sync preservation

### Phase 2: Advanced Features
- [ ] Implement availability control
- [ ] Add pricing tiers system
- [ ] Create VIP access control system
- [ ] Implement VIP pricing benefits
- [ ] Create admin UI for enhanced fields

### Phase 3: Integration
- [ ] Integrate with Order Service
- [ ] Add pricing API endpoints
- [ ] Add VIP validation in order flow
- [ ] Integrate with Chat Service for product discovery
- [ ] Add monitoring & alerts

---

## 🚀 Benefits

| Feature | Benefit |
|---------|---------|
| **Master Data Protection** | Admin enhancements never lost during sync |
| **Flexible Pricing** | Support wholesale, retail, and custom pricing |
| **Advanced Availability** | Granular control over when products are sold |
| **VIP System** | Enhanced customer loyalty and revenue |
| **Chat Integration** | Seamless chat-to-order product discovery |
| **API-First Design** | Easy integration with all services |

**Complete product management system ready for enterprise use! 🎯**