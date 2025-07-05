# Customer Service API Documentation

## 🚀 **Base Information**
- **Service Name**: Customer Service  
- **Port**: 8110
- **Base URL**: `http://customer:8110`
- **Module Name**: `customer`
- **Health Check**: `GET /health`

---

## 📝 **API Endpoints**

### **Customer Management**

#### Get Customer Details
```http
GET /api/v1/customers/{customer_id}
```

**Response:**
```json
{
  "id": "customer_123",
  "name": "John Doe",
  "email": "john@example.com", 
  "phone": "0812345678",
  "vip_level": "gold",
  "vip_discount": 5.0,
  "vip_valid_until": "2024-12-31T23:59:59Z",
  "addresses": [
    {
      "id": "addr_456",
      "type": "home",
      "address": "123 Main Street",
      "district": "Downtown",
      "province": "Bangkok",
      "postal_code": "10100",
      "phone": "0812345678",
      "is_default": true
    }
  ],
  "preferences": {
    "language": "th",
    "currency": "THB",
    "notifications": {
      "email": true,
      "sms": false,
      "push": true
    }
  },
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

#### Get Customer VIP Status
```http
GET /api/v1/customers/{customer_id}/vip-status
```

**Response:**
```json
{
  "customer_id": "customer_123",
  "vip_level": "gold",
  "discount_percentage": 5.0,
  "benefits": [
    "Free delivery",
    "Priority support", 
    "Exclusive products",
    "Early access to sales"
  ],
  "points": 1250,
  "tier_requirements": {
    "current_spending": 45000.00,
    "next_tier": "platinum",
    "spending_needed": 5000.00
  },
  "valid_until": "2024-12-31T23:59:59Z"
}
```

#### Create Customer
```http
POST /api/v1/customers
Content-Type: application/json

{
  "name": "Jane Smith",
  "email": "jane@example.com",
  "phone": "0987654321",
  "address": {
    "address": "456 Second Street",
    "district": "Uptown", 
    "province": "Bangkok",
    "postal_code": "10200"
  },
  "preferences": {
    "language": "en",
    "notifications": {
      "email": true,
      "sms": true
    }
  }
}
```

#### Update Customer
```http
PATCH /api/v1/customers/{customer_id}
Content-Type: application/json

{
  "name": "John Smith",
  "phone": "0812345679",
  "preferences": {
    "language": "en"
  }
}
```

### **Address Management**

#### Add Customer Address
```http
POST /api/v1/customers/{customer_id}/addresses
Content-Type: application/json

{
  "type": "work",
  "address": "789 Business District",
  "district": "Central",
  "province": "Bangkok", 
  "postal_code": "10500",
  "phone": "0812345678"
}
```

#### Update Address
```http
PATCH /api/v1/customers/{customer_id}/addresses/{address_id}
Content-Type: application/json

{
  "address": "Updated address",
  "is_default": true
}
```

### **VIP Management**

#### Upgrade VIP Level
```http
POST /api/v1/customers/{customer_id}/vip/upgrade
Content-Type: application/json

{
  "new_level": "platinum",
  "reason": "spending_threshold",
  "points_used": 0
}
```

#### Add VIP Points
```http
POST /api/v1/customers/{customer_id}/points
Content-Type: application/json

{
  "points": 100,
  "reason": "order_completion",
  "order_id": "order_123"
}
```

---

## 🔗 **Integration Points**

### **Inbound Calls (Services that call Customer Service)**

#### Order Service (8081)
```go
// Customer validation during order creation
GET /api/v1/customers/{id}
GET /api/v1/customers/{id}/vip-status

// Used for: Order validation, VIP pricing, address selection
```

#### Product Service (8083)  
```go
// VIP pricing calculation
GET /api/v1/customers/{id}/vip-status

// Used for: Dynamic pricing based on VIP level
```

#### Payment Service (8087)
```go
// Customer information for payment processing
GET /api/v1/customers/{id}
GET /api/v1/customers/{id}/addresses/{id}

// Used for: Payment validation, billing information
```

#### Shipping Service (8086)
```go
// Delivery address information
GET /api/v1/customers/{id}/addresses
GET /api/v1/customers/{id}/addresses/{id}

// Used for: Delivery cost calculation, address validation
```

#### Chat Service (8090)
```go
// Customer profile for AI context
GET /api/v1/customers/{id}
GET /api/v1/customers/{id}/vip-status

// Used for: Personalized chat responses, order history context
```

---

## 📤 **Events Published**

### **customer.created**
```json
{
  "event_type": "customer.created",
  "customer_id": "customer_123", 
  "name": "John Doe",
  "email": "john@example.com",
  "vip_level": "standard",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### **customer.vip.upgraded**
```json
{
  "event_type": "customer.vip.upgraded",
  "customer_id": "customer_123",
  "old_level": "gold",
  "new_level": "platinum", 
  "new_discount": 10.0,
  "upgraded_at": "2024-01-15T10:30:00Z"
}
```

### **customer.address.updated**
```json
{
  "event_type": "customer.address.updated",
  "customer_id": "customer_123",
  "address_id": "addr_456",
  "is_default": true,
  "updated_at": "2024-01-15T10:30:00Z"
}
```

---

## 📥 **Events Consumed**

### **order.completed**
- **Action**: Add VIP points based on order value
- **Trigger**: Check for VIP level upgrade eligibility

### **payment.completed**
- **Action**: Update customer spending history
- **Trigger**: VIP tier progression calculation

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
PORT=8110
GO_ENV=development

# Redis Cache
REDIS_URL=redis://redis:6379

# Kafka Events
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC_CUSTOMER_EVENTS=customer-events

# VIP System
VIP_POINTS_PER_THB=1
VIP_GOLD_THRESHOLD=25000
VIP_PLATINUM_THRESHOLD=50000
VIP_DIAMOND_THRESHOLD=100000

# Notifications
SMS_PROVIDER=twilio
EMAIL_PROVIDER=sendgrid
```

---

## 🚨 **Error Codes**

| Code | Message | Description |
|------|---------|-------------|
| 400 | Invalid request | Missing or invalid parameters |
| 404 | Customer not found | Customer ID doesn't exist |
| 409 | Email already exists | Duplicate email registration |
| 422 | Invalid VIP operation | Cannot upgrade/downgrade VIP |
| 429 | Rate limit exceeded | Too many requests |
| 500 | Service unavailable | Internal server error |

---

## 💾 **Caching Strategy**

### **Redis Cache Keys**
```redis
# Customer details (30 min TTL)
customer:{customer_id} → Full customer data

# VIP status (1 hour TTL)
customer:vip:{customer_id} → VIP level and benefits

# Customer addresses (1 hour TTL)  
customer:addresses:{customer_id} → All customer addresses

# Points and spending (15 min TTL)
customer:points:{customer_id} → Current points and spending
```

---

## 🧪 **Testing Examples**

### **Get Customer (cURL)**
```bash
curl http://localhost:8110/api/v1/customers/customer_123
```

### **Check VIP Status (cURL)**  
```bash
curl http://localhost:8110/api/v1/customers/customer_123/vip-status
```

### **Create Customer (cURL)**
```bash
curl -X POST http://localhost:8110/api/v1/customers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Customer",
    "email": "test@example.com", 
    "phone": "0812345678",
    "address": {
      "address": "123 Test Street",
      "district": "Test District",
      "province": "Bangkok",
      "postal_code": "10100"
    }
  }'
```

---

## 📊 **VIP Tier System**

### **Tier Levels**
```
Standard (0-24,999 THB)     → 0% discount
Gold (25,000-49,999 THB)    → 5% discount  
Platinum (50,000-99,999 THB) → 10% discount
Diamond (100,000+ THB)      → 15% discount
```

### **Points System**
```
1 THB spent = 1 point earned
100 points = 1 THB discount (in special promotions)
Points expire after 1 year
```

---

> 👤 **Customer Service manages all customer-related data and VIP tier progression**
