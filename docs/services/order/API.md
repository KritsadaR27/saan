# Order Service API Documentation

## 🚀 **Base Information**
- **Service Name**: Order Service  
- **Port**: 8081
- **Base URL**: `http://order:8081`
- **Module Name**: `order`
- **Health Check**: `GET /health`

---

## 📝 **API Endpoints**

### **Order Management**

#### Create Order
```http
POST /api/v1/orders
Content-Type: application/json

{
  "customer_id": "string",
  "items": [
    {
      "product_id": "string",
      "quantity": number,
      "variant_id": "string (optional)"
    }
  ],
  "delivery_address": {
    "address": "string",
    "district": "string", 
    "province": "string",
    "postal_code": "string",
    "phone": "string"
  },
  "delivery_type": "pickup|delivery",
  "payment_method": "cash|card|transfer|wallet",
  "notes": "string (optional)"
}
```

**Response:**
```json
{
  "id": "order_123",
  "status": "pending",
  "customer_id": "customer_456",
  "items": [...],
  "subtotal": 500.00,
  "delivery_fee": 35.00,
  "total": 535.00,
  "delivery_estimate": "30-45 minutes",
  "payment_link": "https://payment.saan.co/pay/order_123",
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### Get Order
```http
GET /api/v1/orders/{order_id}
```

#### Update Order Status
```http
PATCH /api/v1/orders/{order_id}/status
Content-Type: application/json

{
  "status": "confirmed|preparing|ready|delivering|completed|cancelled",
  "reason": "string (required for cancelled)"
}
```

#### List Orders
```http
GET /api/v1/orders?customer_id={id}&status={status}&limit={n}&offset={n}
```

---

## 🔗 **Integration Points**

### **Outbound Calls (Services this service calls)**

#### Product Service
```go
// Product validation and pricing
GET http://product:8083/api/v1/products/{id}
GET http://product:8083/api/v1/products/{id}/pricing?quantity=10&customer_id=123
GET http://product:8083/api/v1/products/{id}/availability
```

#### Customer Service  
```go
// Customer validation and VIP status
GET http://customer:8110/api/v1/customers/{id}
GET http://customer:8110/api/v1/customers/{id}/vip-status
```

#### Inventory Service
```go
// Stock availability check (no reservation)
POST http://inventory:8082/api/v1/availability/check
{
  "items": [{"product_id": "123", "quantity": 5}]
}
```

#### Payment Service
```go
// Payment processing
POST http://payment:8087/api/v1/payments
{
  "order_id": "order_123",
  "amount": 535.00,
  "method": "card",
  "customer_id": "customer_456"
}
```

#### Shipping Service
```go
// Delivery options and cost
POST http://shipping:8086/api/v1/delivery/options
{
  "origin": "store_location",
  "destination": {...},
  "items": [...],
  "weight": 2.5
}
```

---

## 📤 **Events Published**

### **order.created**
```json
{
  "event_type": "order.created",
  "order_id": "order_123",
  "customer_id": "customer_456", 
  "items": [...],
  "total": 535.00,
  "created_at": "2024-01-15T10:30:00Z"
}
```

### **order.confirmed**
```json
{
  "event_type": "order.confirmed",
  "order_id": "order_123",
  "payment_method": "card",
  "delivery_estimate": "30-45 minutes",
  "confirmed_at": "2024-01-15T10:35:00Z"
}
```

### **order.completed**
```json
{
  "event_type": "order.completed", 
  "order_id": "order_123",
  "completion_time": "2024-01-15T11:20:00Z",
  "final_amount": 535.00
}
```

---

## 📥 **Events Consumed**

### **payment.confirmed**
- **Action**: Update order status to "confirmed"
- **Trigger**: Start preparation workflow

### **shipping.delivered**  
- **Action**: Update order status to "completed"
- **Trigger**: Complete order workflow

### **inventory.stock_updated**
- **Action**: Update product availability cache
- **Trigger**: Re-validate pending orders

---

## 🔧 **Configuration**

### **Environment Variables**
```bash
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=saan
DB_PASSWORD=saan_password
DB_NAME=saan_db
DB_SSLMODE=disable

# Server
SERVER_PORT=8081
GO_ENV=development

# External Services
CHAT_SERVICE_URL=http://chatbot:8090
INVENTORY_SERVICE_URL=http://inventory:8082
CUSTOMER_SERVICE_URL=http://customer:8110
PRODUCT_SERVICE_URL=http://product:8083
PAYMENT_SERVICE_URL=http://payment:8087
SHIPPING_SERVICE_URL=http://shipping:8086

# Redis Cache
REDIS_URL=redis://redis:6379

# Kafka Events
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC_ORDER_EVENTS=order-events
```

---

## 🚨 **Error Codes**

| Code | Message | Description |
|------|---------|-------------|
| 400 | Invalid request | Missing or invalid parameters |
| 404 | Order not found | Order ID doesn't exist |
| 409 | Order already confirmed | Cannot modify confirmed order |
| 422 | Product unavailable | One or more items out of stock |
| 500 | Service unavailable | Internal server error |
| 503 | External service error | Dependent service unavailable |

---

## 📊 **Status Flow**

```
pending → confirmed → preparing → ready → delivering → completed
   ↓         ↓          ↓         ↓         ↓
cancelled cancelled  cancelled cancelled cancelled
```

---

## 🧪 **Testing Examples**

### **Create Order (cURL)**
```bash
curl -X POST http://localhost:8081/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "customer_123",
    "items": [
      {
        "product_id": "product_456", 
        "quantity": 2
      }
    ],
    "delivery_address": {
      "address": "123 Main St",
      "district": "Downtown",
      "province": "Bangkok", 
      "postal_code": "10100",
      "phone": "0812345678"
    },
    "delivery_type": "delivery",
    "payment_method": "card"
  }'
```

### **Check Order Status**
```bash
curl http://localhost:8081/api/v1/orders/order_123
```

---

> 🎯 **Order Service is the central orchestrator for all purchase workflows**
