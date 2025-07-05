# 🔧 SAAN PROJECT RULES

## ❌ DO NOT:

* Run services directly with `go run` or `npm run dev`
* Install dependencies on your host machine
* Use `localhost` in your code or API calls
* Create or modify Dockerfiles without team review
* **Use inconsistent service naming between docker-compose and code**
* **Hardcode service URLs without environment variables**
* **Mix different naming conventions in same project**

## ✅ ALWAYS USE:

* `docker-compose up` to run the project
* `docker-compose logs -f [service]` to view service logs
* `docker exec -it [container_name] sh` to enter a container
* Service names (not `localhost`) for internal URLs
* **Consistent naming: docker service = container name = hostname**
* **Environment variables for all service URLs**
* **Standardized health check endpoints (/health)**

---

## 🧩 SERVICES & PORTS

| Service                | Description                          | Port | Docker Container Name | Service Name |
| ---------------------- | ------------------------------------ | ---- | --------------------- | ------------ |
| Web App                | Customer web frontend (Next.js)      | 3008 | web                   | web          |
| Admin Dashboard        | Internal admin panel                 | 3010 | admin                 | admin        |
| Chat Service           | AI / Rule-based reply engine         | 8090 | chat                  | chat         |
| Order Service          | Manages all order logic              | 8081 | order                 | order        |
| Inventory Service      | Manages stock and warehouse          | 8082 | inventory             | inventory    |
| Product Service        | Catalog / SKU management             | 8083 | product               | product      |
| Customer Service       | Customer profile & loyalty system    | 8110 | customer              | customer     |
| Shipping Service       | Delivery & routing logic             | 8086 | shipping              | shipping     |
| Payment Service        | Payment verification & QR            | 8087 | payment               | payment      |
| Finance Service        | Profit / accounting                  | 8085 | finance               | finance      |
| Loyverse Integration   | Loyverse API data sync connector     | 8100 | loyverse-integration  | loyverse     |
| Loyverse Webhook       | Handles Loyverse POS webhooks        | 8093 | loyverse-webhook      | loyverse-webhook |
| Chat Webhook           | FB/LINE message webhooks             | 8094 | chat-webhook          | chat-webhook |
| Delivery Webhook       | Grab/LineMan status webhooks         | 8095 | delivery-webhook      | delivery-webhook |
| Payment Webhook        | Payment gateway webhooks             | 8096 | payment-webhook       | payment-webhook |
| PostgreSQL Database    | Shared relational database           | 5432 | postgres              | postgres     |
| Redis Cache            | Cache layer for fast data access    | 6379 | redis                 | redis        |
| Kafka (Message Bus)    | Event queue system                   | 9092 | kafka                 | kafka        |

---

## 🏗️ ARCHITECTURE PATTERNS

### 📞 **Direct Call (HTTP/gRPC)**

**✅ ใช้เมื่อ:**
- Master data operations (CRUD)
- ต้องการ immediate response
- Transactional operations
- ข้อมูลที่ไม่มี side effects

**🎯 Use Cases:**
```
Product Service ←→ Inventory Service (stock checks, no reservations)
Customer Service ←→ Order Service  
Payment Service ←→ Order Service
User Service ←→ Auth validation
Loyverse sync → Product/Customer/Categories/Discount/Employee/Supplier/Payment_Type/Store/Inventory_level Services
Cart operations → Check availability (no stock reservation)
```

**📝 Examples:**
```go
// ✅ Direct call examples
http://product:8083/api/products/{id}
http://customer:8110/api/customers/{id}
http://inventory:8082/api/stock/check
```

**❌ ไม่ควรใช้เมื่อ:**
- มี downstream effects หลายระบบ
- ไม่ต้องการ immediate response
- Long-running processes

---

### 📨 **Event-Driven (Kafka)**

**✅ ใช้เมื่อ:**
- Business events ที่สำคัญ
- มี multiple consumers
- ต้องการ audit trail
- Async processing

**🎯 Use Cases:**
```
Order Events:
├── order.confirmed → [Customer, Inventory, Analytics]    // หลัง payment confirmed
├── order.completed → [Finance, Inventory, Customer, Analytics]  // หลังส่งของเสร็จ
├── order.cancelled → [Finance, Inventory, Customer]      // เมื่อยกเลิก order
├── payment.confirmed → [Order, Finance, Notification]
└── delivery.completed → [Order, Customer, Analytics]

Inventory Events:
├── stock.low → [Procurement, Analytics, Notification]
├── stock.updated → [Analytics, AI recommendations]

Customer Events:
├── customer.tier_upgraded → [Order pricing, Analytics]
├── customer.vip_achieved → [Notification, Marketing]
```

**📝 Examples:**
```go
// ✅ Event examples
Topic: "order-events"
Topic: "inventory-events"  
Topic: "customer-events"
Topic: "payment-events"
```

**❌ ไม่ควรใช้เมื่อ:**
- Simple CRUD operations
- Master data sync
- ต้องการ immediate response
- ไม่มี downstream consumers

---

### 🗄️ **Redis Cache**

**✅ ใช้เมื่อ:**
- Hot data caching
- Session management
- Real-time counters
- Temporary calculations
- Chat state management

**🎯 Use Cases:**

#### **Session & Authentication:**
```redis
user:session:{session_id} → JWT data, permissions
api:rate_limit:{user_id} → rate limiting counters
auth:blacklist:{token} → invalidated tokens
```

#### **Product & Inventory Cache:**
```redis
product:hot:{product_id} → frequently accessed products
inventory:levels:{product_id} → current stock levels
pricing:calculation:{customer_id} → pricing cache
product:featured → featured products list
```

#### **Chat & Real-time:**
```redis
chat:session:{user_id} → conversation state
chat:history:{user_id} → recent messages
chat:suggestions:{user_id} → AI suggestions
websocket:connections → active connections
```

#### **Order Processing:**
```redis
cart:session:{session_id} → shopping cart state (no stock reservation)
order:draft:{order_id} → order being built
order:pricing:{order_id} → calculated pricing
checkout:validation:{customer_id} → final stock check before order
```

#### **Analytics & Metrics:**
```redis
metrics:daily:orders:{date} → real-time order counts
metrics:daily:revenue:{date} → running totals
analytics:trends:{category} → trending products
dashboard:stats:{date} → dashboard data
```

#### **Notification Queue:**
```redis
notifications:pending:{user_id} → pending notifications
email:queue → email sending queue
sms:queue → SMS sending queue
```

**📝 Examples:**
```go
// ✅ Redis usage examples
redis.Set("product:123", productData, 1*time.Hour)
redis.Get("user:session:abc123")
redis.Incr("metrics:daily:orders:2025-07-03")
redis.HSet("cart:user123", "product456", quantity)
```

**❌ ไม่ควรใช้เมื่อ:**
- Master data storage (ใช้ PostgreSQL)
- Complex relationships
- ข้อมูลที่ต้อง ACID compliance
- Long-term data storage

**📸 Snapshot Usage:**
```redis
# ❌ DON'T snapshot temporary data
cart:user123 → Don't snapshot every cart change
product:views → Don't snapshot view counts

# ✅ DO snapshot critical business events
order:snapshot:order123 → Order state changes
inventory:transaction:tx456 → Stock movements
daily:inventory:2025-07-03 → Daily stock levels
```

---

## 🎯 **Key Principles**

### **📞 Direct Call Principle:**
**"Immediate response needed = Direct Call"**
- Cart operations (user expects instant feedback)
- Stock checks (real-time availability, no reservations)
- Checkout validation (final stock verification)
- Payment processing (synchronous validation required)

### **📨 Event-Driven Principle:**  
**"Business impact across services = Event"**
- Order state changes (confirmed/completed/cancelled)
- Payment confirmations (affects multiple systems)
- Inventory level changes (triggers procurement/analytics)

### **🗄️ Cache Principle:**
**"Frequently accessed or temporary data = Redis"**
- Shopping carts (temporary, session-based)
- User sessions (temporary authentication)
- Hot product data (frequently accessed)
- Real-time metrics (counters, analytics)

### **📸 Snapshot Principle:**
**"Critical business state changes = Snapshot"**
- Order lifecycle snapshots (created/confirmed/completed)
- Inventory transactions (deducted/restocked/adjusted)
- Daily/monthly inventory levels (for reporting)
- Financial transactions (for compliance)

## 🔄 COMMUNICATION DECISION MATRIX
|----------|---------|---------|
| Scenario | Pattern | Example |
|----------|---------|---------|
| **Add to Cart** | Direct Call | `POST http://order:8081/api/cart/add` |
| **Stock Check** | Direct Call | `GET http://inventory:8082/api/stock/{id}` |
| **Create Order** | Direct Call | `POST http://order:8081/api/orders` |
| **Order Confirmed** | Event-Driven | `Publish: order.confirmed` |
| **Order Completed** | Event-Driven | `Publish: order.completed` |
| **Stock Updated** | Event-Driven | `Publish: inventory.updated` |
| **Cache Product** | Redis | `SET product:123 {data}` |
| **User Session** | Redis | `SET session:abc {jwt}` |
| **Sync from Loyverse** | Direct Call | `POST http://product:8083/sync/loyverse` |
| **Sync Categories** | Direct Call | `POST http://product:8083/sync/categories` |
| **Sync Customers** | Direct Call | `POST http://customer:8110/sync/loyverse` |
| **Sync Inventory** | Direct Call | `POST http://inventory:8082/sync/levels` |
| **Payment Confirmed** | Event-Driven | `Publish: payment.confirmed` |
| **Chat State** | Redis | `SET chat:user123 {state}` |
| **Real-time Metrics** | Redis | `INCR orders:today` |

---

## 🔄 INTERNAL COMMUNICATION RULES

### **✅ Standard Service Communication:**
Use service names as hostnames for API calls and DB access:

```go
// ✅ Correct service URLs (matching docker-compose)
http://order:8081/api/orders
http://customer:8110/api/customers
http://inventory:8082/api/stock
http://payment:8087/api/payments
http://shipping:8086/api/delivery
http://finance:8085/api/revenue
http://loyverse:8100/api/sync

// Database connections
postgres://saan:saan_password@postgres:5432/saan_db
redis://redis:6379
kafka:9092
```

### **🏗️ Environment Variable Standards:**
```go
// Service URLs (always use env vars)
CUSTOMER_SERVICE_URL=http://customer:8110
ORDER_SERVICE_URL=http://order:8081
INVENTORY_SERVICE_URL=http://inventory:8082
PAYMENT_SERVICE_URL=http://payment:8087
SHIPPING_SERVICE_URL=http://shipping:8086
FINANCE_SERVICE_URL=http://finance:8085
CHAT_SERVICE_URL=http://chat:8090

// Database connections
DATABASE_URL=postgres://saan:saan_password@postgres:5432/saan_db?sslmode=disable
REDIS_ADDR=redis:6379
KAFKA_BROKERS=kafka:9092
```

### **🚨 Health Check Standards:**
All services must implement:
```go
GET /health → {"status": "ok", "service": "order", "version": "1.0.0"}
GET /metrics → Prometheus metrics
GET /ready → Readiness probe
```

---

## 📊 **Pattern Selection Guide**

### **🚀 High Priority (Real-time)**
```
Direct Call Pattern:
├── Order creation (draft/pending)
├── Cart operations (add/remove/update)
├── Stock availability checks (no reservations)
├── Payment processing  
├── Product lookups
├── Customer authentication
├── Final checkout validation
└── Master data sync (Products/Categories/Customers/Suppliers/etc. from Loyverse)
```

### **⚡ Medium Priority (Business Events)**
```
Event-Driven Pattern:
├── Order confirmation (หลัง payment)
├── Order completion (หลังส่งของ)
├── Payment confirmation
├── Inventory changes
├── Customer tier upgrades
└── Delivery updates
```

### **🔄 Support Layer (Performance)**
```
Redis Cache Pattern:
├── Session management
├── Hot data caching
├── Real-time counters
├── Chat state
└── Temporary calculations
```

---

## ⚙️ DEVELOPMENT TIPS

### **🐳 Docker Best Practices:**
* Use `.env.local` files for environment variables
* Keep `docker-compose.override.yml` for local dev tweaks
* **All services must have health checks**
* **Consistent port mapping: service runs on same port inside/outside container**
* **Use air or reflex in Go services for hot-reload**

### **📝 Code Standards:**
* **Service discovery via environment variables only**
* **Never hardcode service URLs in code**
* **Use consistent error handling across services**
* **Implement proper logging with structured format**
* **All endpoints must return consistent JSON response format**

### **🔧 Naming Conventions:**
```yaml
# docker-compose.yml structure
services:
  service-name:           # kebab-case
    container_name: service-name    # same as service name
    environment:
      - SERVICE_NAME_VAR=value      # UPPER_SNAKE_CASE with service prefix
```

### **🚀 Performance Rules:**
* **Enable HTTP/2 for service-to-service communication**
* **Use connection pooling for database connections**
* **Implement proper timeouts (5s request, 30s connection)**
* **Cache frequently accessed data in Redis with proper TTL**

---

## 🎯 **Anti-Patterns to Avoid**

### ❌ **Don't Do This:**
```go
// ❌ Using Kafka for simple lookups
kafka.Publish("get.product", productID)

// ❌ Using Direct calls for business events  
http.Post("http://finance:8085/revenue", orderData)
http.Post("http://analytics:8098/order", orderData) 
http.Post("http://customer:8110/points", orderData)

// ❌ Storing master data in Redis only
redis.Set("product:123", productData) // Without DB backup

// ❌ Using Redis as primary database
redis.Set("order:456", orderData) // No PostgreSQL backup

// ❌ Stock reservations in Redis (we don't do this)
redis.Set("stock:reserved:123", reservationData) // Not our approach
```

### ✅ **Do This Instead:**
```go
// ✅ Direct call for lookups
product := http.Get("http://product:8083/api/products/123")

// ✅ Event for business operations
kafka.Publish("order.completed", orderEvent)

// ✅ Cache with DB backup
db.Save(product)
redis.Set("product:123", product, 1*time.Hour) // Cache only

// ✅ Stock checks without reservations (our approach)
stockLevel := http.Get("http://inventory:8082/api/stock/123")
if stockLevel.Available >= requestedQty {
    // Proceed with cart update, validate again at checkout
}
```

---

> ✅ **Stick to these patterns and we guarantee a smooth Dev → Test → Prod transition with optimal performance!**