# Product Service API Documentation

## 🚀 **Base Information**
- **Service Name**: Product Service  
- **Port**: 8083
- **Base URL**: `http://product:8083`
- **Module Name**: `product`
- **Health Check**: `GET /health`

---

## 📝 **API Endpoints**

### **Product Catalog**

#### Get Product Details
```http
GET /api/v1/products/{product_id}
```

**Response:**
```json
{
  "id": "product_123",
  "name": "Premium Coffee Beans",
  "description": "Arabica coffee beans from Thailand",
  "sku": "COFFEE-001",
  "base_price": 250.00,
  "category": "beverages",
  "brand": "SAAN Coffee",
  "weight": 500,
  "unit": "grams",
  "tags": ["premium", "arabica", "thai"],
  "images": [
    "https://cdn.saan.co/products/coffee-001-1.jpg",
    "https://cdn.saan.co/products/coffee-001-2.jpg"
  ],
  "variants": [
    {
      "id": "var_roast_light",
      "name": "Light Roast",
      "price_modifier": 0.00
    },
    {
      "id": "var_roast_dark", 
      "name": "Dark Roast",
      "price_modifier": 25.00
    }
  ],
  "availability": {
    "available": true,
    "stock_level": "high",
    "estimated_days": 1
  },
  "tier_pricing": [
    {
      "min_quantity": 1,
      "max_quantity": 4,
      "name": "Retail",
      "price": 250.00
    },
    {
      "min_quantity": 5,
      "max_quantity": 19,
      "name": "Bulk 5+",
      "price": 230.00
    },
    {
      "min_quantity": 20,
      "max_quantity": null,
      "name": "Wholesale 20+",
      "price": 200.00
    }
  ],
  "loyverse_id": "loyverse_product_456",
  "created_at": "2024-01-10T08:00:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

#### Calculate Product Pricing
```http
GET /api/v1/products/{product_id}/pricing?quantity=10&customer_id=123&variant_id=var_roast_dark
```

**Response:**
```json
{
  "product_id": "product_123",
  "variant_id": "var_roast_dark",
  "quantity": 10,
  "customer_id": "customer_123",
  "pricing": {
    "base_price": 250.00,
    "variant_modifier": 25.00,
    "unit_price": 275.00,
    "tier_price": 250.00,
    "tier_name": "Bulk 5+",
    "vip_discount": 12.50,
    "final_unit_price": 237.50,
    "total_price": 2375.00
  },
  "customer_vip": {
    "level": "gold",
    "discount_percentage": 5.0
  },
  "calculated_at": "2024-01-15T10:35:00Z"
}
```

#### Check Product Availability
```http
GET /api/v1/products/{product_id}/availability?quantity=5
```

**Response:**
```json
{
  "product_id": "product_123",
  "available": true,
  "requested_quantity": 5,
  "available_quantity": 150,
  "stock_level": "high",
  "estimated_availability": "immediate",
  "estimated_days": 0,
  "variants": [
    {
      "variant_id": "var_roast_light",
      "available": true,
      "quantity": 75
    },
    {
      "variant_id": "var_roast_dark",
      "available": true, 
      "quantity": 75
    }
  ]
}
```

#### Search Products
```http
GET /api/v1/products/search?q=coffee&category=beverages&limit=20&offset=0
```

**Response:**
```json
{
  "products": [
    {
      "id": "product_123",
      "name": "Premium Coffee Beans",
      "price": 250.00,
      "image": "https://cdn.saan.co/products/coffee-001-1.jpg",
      "available": true,
      "rating": 4.8
    }
  ],
  "total": 45,
  "limit": 20,
  "offset": 0,
  "facets": {
    "categories": [
      {"name": "beverages", "count": 25},
      {"name": "food", "count": 20}
    ],
    "brands": [
      {"name": "SAAN Coffee", "count": 15},
      {"name": "Local Roasters", "count": 10}
    ]
  }
}
```

### **Product Management (Admin)**

#### Create Product
```http
POST /api/v1/products
Content-Type: application/json
Authorization: Bearer {admin_token}

{
  "name": "New Product",
  "description": "Product description",
  "sku": "PROD-002", 
  "base_price": 150.00,
  "category": "food",
  "brand": "SAAN",
  "weight": 300,
  "unit": "grams"
}
```

#### Update Product
```http
PATCH /api/v1/products/{product_id}
Content-Type: application/json
Authorization: Bearer {admin_token}

{
  "name": "Updated Product Name",
  "base_price": 275.00
}
```

#### Sync with Loyverse
```http
POST /api/v1/products/sync/loyverse
Authorization: Bearer {admin_token}

{
  "store_id": "loyverse_store_123",
  "product_ids": ["loyverse_prod_456", "loyverse_prod_789"]
}
```

---

## 🔗 **Integration Points**

### **Outbound Calls (Services this service calls)**

#### Customer Service (8110)
```go
// VIP level validation for pricing
GET http://customer:8110/api/v1/customers/{id}/vip-status

Response:
{
  "customer_id": "customer_123",
  "vip_level": "gold",
  "discount_percentage": 5.0,
  "valid_until": "2024-12-31T23:59:59Z"
}
```

#### Loyverse Integration (8091)
```go
// Product data synchronization
GET http://loyverse-integration:8091/api/v1/products/{loyverse_id}
POST http://loyverse-integration:8091/api/v1/products/sync

// Webhook from Loyverse
POST /api/v1/webhooks/loyverse/product-updated
{
  "loyverse_id": "loyverse_prod_456",
  "name": "Updated Product Name",
  "price": 280.00,
  "updated_at": "2024-01-15T10:30:00Z"
}
```

---

## 📤 **Events Published**

### **product.updated**
```json
{
  "event_type": "product.updated",
  "product_id": "product_123",
  "changes": {
    "price": {"old": 250.00, "new": 275.00},
    "availability": {"old": true, "new": false}
  },
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### **product.pricing.changed**
```json
{
  "event_type": "product.pricing.changed",
  "product_id": "product_123", 
  "tier_pricing": [...],
  "affected_customers": ["vip", "wholesale"],
  "effective_date": "2024-01-16T00:00:00Z"
}
```

### **product.availability.changed**
```json
{
  "event_type": "product.availability.changed",
  "product_id": "product_123",
  "available": false,
  "reason": "out_of_stock",
  "estimated_restock": "2024-01-20T00:00:00Z"
}
```

---

## 📥 **Events Consumed**

### **loyverse.product.synced**
- **Action**: Update product data from Loyverse
- **Trigger**: Sync product information, preserve custom fields

### **inventory.stock.updated**
- **Action**: Update product availability status
- **Trigger**: Refresh availability cache

### **customer.vip.upgraded**
- **Action**: Clear pricing cache for customer
- **Trigger**: Recalculate pricing for pending orders

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
PORT=8083
GO_ENV=development

# External Services
CUSTOMER_SERVICE_URL=http://customer:8110
LOYVERSE_INTEGRATION_URL=http://loyverse-integration:8091

# Redis Cache
REDIS_URL=redis://redis:6379

# Loyverse Integration
LOYVERSE_API_KEY=your_api_key
LOYVERSE_WEBHOOK_SECRET=your_webhook_secret

# Image Storage
CDN_BASE_URL=https://cdn.saan.co
UPLOAD_MAX_SIZE=10MB

# Search
ELASTICSEARCH_URL=http://elasticsearch:9200
```

---

## 🚨 **Error Codes**

| Code | Message | Description |
|------|---------|-------------|
| 400 | Invalid request | Missing or invalid parameters |
| 404 | Product not found | Product ID doesn't exist |
| 409 | SKU already exists | Duplicate SKU in catalog |
| 422 | Invalid pricing | Negative price or invalid tier |
| 429 | Rate limit exceeded | Too many requests |
| 500 | Service unavailable | Internal server error |
| 503 | External service error | Loyverse API unavailable |

---

## 💾 **Caching Strategy**

### **Redis Cache Keys**
```redis
# Product details (1 hour TTL)
product:{product_id} → Full product data

# Product pricing (30 min TTL)
product:pricing:{product_id}:{quantity}:{vip_level} → Calculated pricing

# Product availability (5 min TTL) 
product:availability:{product_id} → Stock status

# Search results (15 min TTL)
product:search:{query_hash} → Search results

# VIP customer data (30 min TTL)
product:customer:vip:{customer_id} → VIP level and discount
```

---

## 🧪 **Testing Examples**

### **Get Product (cURL)**
```bash
curl http://localhost:8083/api/v1/products/product_123
```

### **Calculate Pricing (cURL)**
```bash
curl "http://localhost:8083/api/v1/products/product_123/pricing?quantity=10&customer_id=customer_456"
```

### **Search Products (cURL)**
```bash
curl "http://localhost:8083/api/v1/products/search?q=coffee&category=beverages"
```

### **Create Product (Admin)**
```bash
curl -X POST http://localhost:8083/api/v1/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer admin_token_here" \
  -d '{
    "name": "Test Product",
    "sku": "TEST-001",
    "base_price": 100.00,
    "category": "test"
  }'
```

---

## 📊 **Performance Metrics**

### **Key Metrics**
- Product lookup response time: <100ms
- Pricing calculation time: <50ms
- Search response time: <200ms
- Cache hit rate: >90%
- Loyverse sync success rate: >95%

---

> 🛍️ **Product Service provides the foundation for all e-commerce operations in SAAN**
