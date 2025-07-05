# 💳 Payment Service Implementation Plan

## 🎯 Overview

Payment Service for SAAN system with enhanced Loyverse integration, multi-store management, and delivery context tracking for clear financial separation and driver commission management.

## 📋 Implementation Scope

### **Core Features**
- Enhanced Loyverse POS integration with multi-store support
- Automatic and manual store selection for receipts
- Delivery driver context tracking and commission management
- Rich receipt notes with driver and delivery information
- Store-based financial separation and analytics

### **Integration Points**
- **Order Service**: Order details and item information
- **Customer Service**: Customer data for Loyverse receipts
- **Shipping Service**: Delivery context and driver information
- **Loyverse POS**: Multi-store receipt creation and management

## 🏗️ Architecture

### **Service Structure (Clean Architecture)**
```
payment-service/
├── cmd/
│   └── main.go
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── payment_transaction.go
│   │   │   ├── loyverse_store.go
│   │   │   └── payment_delivery_context.go
│   │   ├── repository/
│   │   │   ├── payment_repository.go
│   │   │   ├── loyverse_store_repository.go
│   │   │   └── payment_delivery_context_repository.go
│   │   └── service/
│   ├── application/
│   │   ├── usecase/
│   │   │   ├── payment_usecase.go
│   │   │   ├── store_selection_usecase.go
│   │   │   ├── loyverse_integration_usecase.go
│   │   │   └── delivery_context_usecase.go
│   │   └── dto/
│   ├── infrastructure/
│   │   ├── database/
│   │   ├── cache/
│   │   ├── events/
│   │   ├── repository/
│   │   ├── external/
│   │   │   ├── loyverse_client.go
│   │   │   ├── order_service_client.go
│   │   │   ├── customer_service_client.go
│   │   │   └── shipping_service_client.go
│   │   └── config/
│   └── transport/
│       └── http/
│           ├── handler/
│           │   ├── payment_handler.go
│           │   ├── store_handler.go
│           │   └── loyverse_handler.go
│           └── middleware/
└── migrations/
```

## 💾 Database Schema

### **Core Payment Tables**
```sql
-- Payment transactions (main table)
CREATE TABLE payment_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    customer_id UUID NOT NULL,
    
    -- Payment details
    payment_method VARCHAR(50) NOT NULL,     -- 'cash', 'bank_transfer', 'cod_cash', 'cod_transfer'
    payment_channel VARCHAR(50) NOT NULL,   -- 'loyverse_pos', 'saan_app', 'saan_chat', 'delivery'
    payment_timing VARCHAR(50) NOT NULL,    -- 'prepaid', 'cod'
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'THB',
    
    -- Payment status
    status VARCHAR(50) DEFAULT 'pending',   -- 'pending', 'completed', 'failed', 'refunded'
    paid_at TIMESTAMP,
    
    -- Loyverse integration
    loyverse_receipt_id VARCHAR(100),
    loyverse_payment_type VARCHAR(50),
    assigned_store_id VARCHAR(100),
    
    -- Audit fields
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_payment_transactions_order (order_id),
    INDEX idx_payment_transactions_customer (customer_id),
    INDEX idx_payment_transactions_status (status),
    INDEX idx_payment_transactions_store (assigned_store_id),
    INDEX idx_payment_transactions_created (created_at)
);
```

### **Multi-Store Management Tables**
```sql
-- Loyverse store configuration
CREATE TABLE loyverse_stores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id VARCHAR(100) UNIQUE NOT NULL,
    store_name VARCHAR(200) NOT NULL,
    store_type VARCHAR(50) NOT NULL,          -- 'main', 'delivery', 'warehouse'
    is_active BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    
    -- Store capabilities
    accepts_cash BOOLEAN DEFAULT true,
    accepts_transfer BOOLEAN DEFAULT true,
    accepts_cod BOOLEAN DEFAULT false,
    
    -- Delivery configuration
    delivery_driver_phone VARCHAR(20),
    delivery_route VARCHAR(50),
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_loyverse_stores_active (is_active, store_type),
    INDEX idx_loyverse_stores_delivery (delivery_driver_phone)
);

-- Payment delivery context for COD and delivery tracking
CREATE TABLE payment_delivery_context (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_transaction_id UUID REFERENCES payment_transactions(id) ON DELETE CASCADE,
    
    -- Delivery context
    delivery_driver_name VARCHAR(100),
    delivery_driver_phone VARCHAR(20),
    delivery_route VARCHAR(50),
    delivery_app VARCHAR(50),               -- 'saan_delivery', 'grab', 'lineman'
    
    -- Store assignment
    assigned_store_id VARCHAR(100) REFERENCES loyverse_stores(store_id),
    store_assignment_reason TEXT,           -- 'driver_route', 'manual_selection', 'default'
    
    -- Receipt context
    receipt_note TEXT,
    
    created_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_payment_delivery_context_payment (payment_transaction_id),
    INDEX idx_payment_delivery_context_driver (delivery_driver_phone),
    INDEX idx_payment_delivery_context_store (assigned_store_id)
);
```

## 🔧 Key Components

### **1. Store Selection Logic**
- **Automatic Selection**: Based on payment method, channel, and delivery context
- **Manual Override**: Allow admin to force specific store selection
- **Driver Assignment**: Auto-assign store based on driver phone/route
- **Fallback Logic**: Default store when no specific match found

### **2. Receipt Context Generation**
- **Driver Information**: Name, phone, route, delivery app
- **Payment Context**: Channel (app, chat, POS), method, timing
- **Store Context**: Store name and type for delivery stores
- **Custom Notes**: Manual notes for special cases

### **3. Loyverse Integration**
- **Multi-Store Support**: Create receipts in different Loyverse stores
- **Rich Receipt Notes**: Contextual information for financial tracking
- **Customer Sync**: Ensure customer exists in Loyverse before receipt creation
- **Error Handling**: Robust error handling for API failures

### **4. Financial Separation**
- **Store-Based Reporting**: Revenue tracking per store/route
- **Driver Commission**: Track sales per driver for commission calculation
- **Reconciliation Support**: Clear audit trail for accounting

## 📊 API Endpoints

### **Payment Management**
```
POST   /api/v1/payments                      # Create payment transaction
GET    /api/v1/payments/{id}                 # Get payment details
PUT    /api/v1/payments/{id}/status          # Update payment status
GET    /api/v1/payments/order/{order_id}     # Get payments by order
```

### **Store Management**
```
GET    /api/v1/stores/available              # Get available stores
GET    /api/v1/stores/{store_id}             # Get store details
POST   /api/v1/stores/selection-preview      # Preview store selection
PUT    /api/v1/stores/{store_id}/config      # Update store config (Admin)
```

### **Loyverse Integration**
```
POST   /api/v1/loyverse/receipts             # Create receipt (automatic store)
POST   /api/v1/loyverse/receipts/custom      # Create custom receipt with store selection
PUT    /api/v1/loyverse/receipts/{id}/regenerate  # Regenerate receipt
GET    /api/v1/loyverse/receipts/by-store/{store_id}  # Get receipts by store
```

### **Delivery Context**
```
PUT    /api/v1/payments/{id}/delivery-context    # Update delivery context
GET    /api/v1/payments/{id}/receipt-preview     # Preview receipt note
POST   /api/v1/payments/{id}/assign-driver       # Assign driver to payment
```

### **Analytics**
```
GET    /api/v1/analytics/stores               # Store performance analytics
GET    /api/v1/analytics/drivers              # Driver performance analytics
GET    /api/v1/analytics/stores/{id}/revenue  # Store-specific revenue
```

## 🎯 Implementation Phases

### **Phase 1: Core Payment Service (Week 1-2)**
- ✅ Basic payment transaction management
- ✅ Database schema and migrations
- ✅ Domain entities and repositories
- ✅ Basic Loyverse integration

### **Phase 2: Multi-Store Management (Week 3-4)**
- 🚧 Store configuration and management
- 🚧 Store selection logic implementation
- 🚧 Enhanced Loyverse integration with store selection
- 🚧 API endpoints for store management

### **Phase 3: Delivery Context & Analytics (Week 5-6)**
- 📋 Delivery context tracking
- 📋 Driver assignment and route management
- 📋 Rich receipt note generation
- 📋 Store and driver analytics

### **Phase 4: Advanced Features (Week 7-8)**
- 📋 Custom receipt creation
- 📋 Receipt regeneration
- 📋 Financial reporting and reconciliation
- 📋 Commission calculation support

## 🔗 Integration Requirements

### **External Services**
- **Order Service**: `GET /api/v1/orders/{id}` - Order details and items
- **Customer Service**: `GET /api/v1/customers/{id}` - Customer information
- **Shipping Service**: `GET /api/v1/deliveries/{id}` - Delivery context and driver info
- **Loyverse API**: POS integration for receipt creation

### **Event Publishing**
- `payment.created` - Payment transaction created
- `payment.completed` - Payment successfully processed
- `payment.failed` - Payment processing failed
- `loyverse.receipt.created` - Receipt created in Loyverse
- `store.assigned` - Store assigned to payment

### **Event Subscriptions**
- `order.confirmed` - Create payment transaction
- `delivery.assigned` - Update delivery context
- `delivery.completed` - Process COD payment

## 📈 Success Metrics

### **Business Metrics**
- **Store Revenue Tracking**: Revenue per store/route
- **Driver Performance**: Sales and commission per driver
- **Payment Success Rate**: % of successful payments
- **Loyverse Integration**: % of payments with receipts

### **Technical Metrics**
- **API Response Time**: < 200ms for payment operations
- **Store Selection Accuracy**: > 95% automatic selection success
- **Loyverse API Success**: > 99% receipt creation success
- **Error Rate**: < 1% payment processing errors

## 🚀 Deployment Strategy

### **Infrastructure Requirements**
- **Database**: PostgreSQL with read replicas
- **Cache**: Redis for store configurations and session data
- **Message Queue**: Kafka for event-driven communication
- **External APIs**: Loyverse API access and rate limiting

### **Monitoring & Alerting**
- Payment processing failures
- Loyverse API errors
- Store selection failures
- High error rates or latency

---

> 💳 **Enhanced Payment Service with Multi-Store Management for clear financial separation and driver commission tracking!**
