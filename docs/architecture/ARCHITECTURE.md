# 🎯 Complete SAAN Services Flow Diagrams

## 🏗 **System Architecture Overview**
```
🏪 Admin Dashboard (3010) ←→ 📱 Web App (3008) ←→ 💬 Chat (8090)
                    ↓                ↓                     ↓
🛒 Order Service (8081) ← Central Orchestrator → 🔗 Webhook Listener (8091)
    ↓    ↓    ↓    ↓    ↓    ↓    ↓    ↓    ↓    ↓    ↓    ↓    ↓
[Product] [Customer] [Payment] [Inventory] [Shipping] [Finance] [AI] [Analytics] [Procurement] [Notification] [Reporting] [User]
   8083     8110      8087      8082       8086      8085    8097    8098        8099         8092          8089       8088
    ↓         ↓        ↓         ↓          ↓         ↓       ↓       ↓           ↓            ↓             ↓          ↓
PostgreSQL ←→ Redis ←→ Kafka ←→ Loyverse Integration (8100) ←→ Static CDN (8101) ←→ API Gateway (8080)
   5432      6379     9092
```

---

## 💬 **Chat Service (8090) Flow**

### 🤖 **AI-Powered Conversation Management**
```
Customer Message → Chat Service
├── 1. Message received from webhook (LINE/Facebook)
├── 2. Identify customer and conversation context
├── 3. Process message intent using AI/NLP
├── 4. Check for order-related requests
├── 5. Generate appropriate response
├── 6. Update conversation state → Redis
├── 7. Send response via appropriate channel
└── 8. Log interaction for analytics

Chat State Management:
GET /api/chat/sessions/{user_id}
├── Load conversation history from Redis
├── Get customer context from Customer Service
├── Check for active orders from Order Service
├── Analyze conversation patterns
├── Generate contextual AI responses
└── Update session state

Database Schema:
CREATE TABLE chat_sessions (
    id UUID PRIMARY KEY,
    user_id VARCHAR(100), -- LINE/Facebook user ID
    customer_id UUID REFERENCES customers(id),
    platform VARCHAR(20), -- 'line', 'facebook', 'whatsapp'
    session_start TIMESTAMP DEFAULT NOW(),
    last_activity TIMESTAMP DEFAULT NOW(),
    conversation_context JSONB,
    is_active BOOLEAN DEFAULT true,
    
    INDEX idx_chat_user (user_id, platform),
    INDEX idx_chat_customer (customer_id),
    INDEX idx_chat_active (is_active, last_activity)
);

CREATE TABLE chat_messages (
    id UUID PRIMARY KEY,
    session_id UUID REFERENCES chat_sessions(id),
    message_type VARCHAR(20), -- 'text', 'image', 'sticker', 'location'
    content TEXT,
    metadata JSONB, -- Platform-specific data
    sender_type VARCHAR(10), -- 'customer', 'bot', 'agent'
    intent VARCHAR(50), -- 'order_inquiry', 'product_question', 'complaint'
    created_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_messages_session (session_id, created_at),
    INDEX idx_messages_intent (intent)
);
```

### 🔄 **Chat-to-Order Conversion Flow**
```
Order Intent Detected → Chat Service
├── 1. Analyze message for product requests
├── 2. Extract product names/quantities
├── 3. Search products via Product Service
├── 4. Present product options to customer
├── 5. Confirm selections and quantities
├── 6. Calculate pricing and delivery
├── 7. Create draft order via Order Service
├── 8. Guide customer through checkout
├── 9. Send order confirmation
└── 10. Hand off to Order Service

Chat Response Examples:
{
  "message_type": "product_suggestions",
  "content": "เจอสินค้าที่คุณต้องการแล้วค่ะ:",
  "suggestions": [
    {
      "product_id": "prod_123",
      "name": "โค้ก 325ml",
      "price": 15.00,
      "stock": 100,
      "image": "https://cdn.saan.com/products/coke-325ml.jpg"
    }
  ],
  "quick_replies": [
    {"text": "เพิ่มลงตะกร้า", "action": "add_to_cart"},
    {"text": "ดูเพิ่มเติม", "action": "view_details"},
    {"text": "ข้าม", "action": "skip"}
  ]
}
```

---

## 👤 **Customer Service (8110) Flow**

### 🏠 **Customer Address Management**
```
Customer Registration with Address Flow:
Admin creates customer → POST /api/customers/create
├── 1. Validate customer basic info (name, phone)
├── 2. Admin types subdistrict name
├── 3. Get address suggestions → GET /api/addresses/suggest?q=หัวหมาก
│   └── Return: [
│       "1. หัวหมาก > บางกะปิ > กรุงเทพมหานคร (10240)",
│       "2. หัวหมาก > เมือง > ร้อยเอ็ด (45000)"
│     ]
├── 4. Admin selects address option
├── 5. Auto-populate: subdistrict, district, province, postal_code
├── 6. Admin completes: house_number, address_line1, location_name
├── 7. Determine delivery route based on province
├── 8. Set as default address
├── 9. Cache customer data → Redis
└── 10. Create customer profile

Database Schema:
CREATE TABLE customers (
    id UUID PRIMARY KEY,
    phone VARCHAR(20) UNIQUE,
    name VARCHAR(100),
    email VARCHAR(100),
    tier customer_tier_enum DEFAULT 'bronze',
    points_balance INT DEFAULT 0,
    total_spent DECIMAL(12,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE customer_addresses (
    id UUID PRIMARY KEY,
    customer_id UUID REFERENCES customers(id),
    location_name VARCHAR(100), -- "บ้าน", "ออฟฟิศ", "คลัง"
    house_number VARCHAR(20),
    address_line1 TEXT,
    subdistrict VARCHAR(100),
    district VARCHAR(100), 
    province VARCHAR(100),
    postal_code VARCHAR(10),
    delivery_route VARCHAR(50), -- Auto-assigned based on location
    is_default BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_customer_addresses (customer_id),
    INDEX idx_delivery_route (delivery_route)
);

-- Thai address lookup table
CREATE TABLE thai_addresses (
    id UUID PRIMARY KEY,
    subdistrict VARCHAR(100),
    district VARCHAR(100),
    province VARCHAR(100), 
    postal_code VARCHAR(10),
    is_self_delivery_area BOOLEAN, -- Based on 11 provinces
    delivery_route VARCHAR(50),
    INDEX idx_subdistrict_search (subdistrict),
    INDEX idx_province_delivery (province, is_self_delivery_area)
);
```

### 📊 **Customer Analytics & AI Insights**
```
Customer Profile Dashboard Flow:
Admin views customer → GET /api/customers/{id}/profile
├── 1. Get customer basic info
├── 2. Get all addresses with delivery preferences
├── 3. Calculate purchase history analytics
├── 4. Get AI customer insights
├── 5. Generate care recommendations
├── 6. Return comprehensive profile

Customer Analytics Response:
{
  "customer": {
    "id": "cust_123",
    "name": "นายสมชาย ใจดี",
    "phone": "+66812345678",
    "tier": "silver",
    "points_balance": 1250
  },
  "addresses": [
    {
      "id": "addr_456", 
      "location_name": "บ้าน",
      "full_address": "123/45 ซอยสุขุมวิท 71 หัวหมาก บังกะปิ กรุงเทพฯ 10240",
      "delivery_route": "route_a",
      "is_default": true
    }
  ],
  "purchase_analytics": {
    "total_orders": 25,
    "total_spent": 25000.00,
    "average_order_value": 1000.00,
    "last_purchase": "2025-06-15",
    "purchase_frequency": "2.1 times/month",
    "top_categories": ["beverages", "snacks", "household"],
    "favorite_products": [
      {"name": "โค้ก 325ml", "quantity": 48, "total_spent": 720},
      {"name": "มาม่า", "quantity": 24, "total_spent": 240}
    ]
  },
  "ai_insights": {
    "customer_segment": "loyal_regular",
    "churn_risk": 0.15, // Low risk
    "predicted_next_order": "2025-07-10",
    "recommended_products": ["เป็บซี่ 140g", "น้ำปลาทิพรส"],
    "upsell_opportunities": [
      {
        "product": "โค้ก 1.25L", 
        "reason": "Customer buys Coke 325ml frequently, bigger size = better value"
      }
    ],
    "care_recommendations": [
      "ลูกค้าซื้อสม่ำเสมอ แนะนำให้เสนอ bulk discount",
      "ชอบซื้อเครื่องดื่ม แนะนำ combo set ใหม่",
      "ไม่เคยลองขนมขบเคี้ยว premium แนะนำให้ลอง"
    ]
  }
}
```

---

## 🛒 **Order Service (8081) Flow - Enhanced with Snapshots**

### 📊 **Order Snapshot Management**
```
Order Snapshot Creation Flow:
Order Status Changes → Order Service creates snapshot
├── 1. Order Created (draft) → Initial snapshot
├── 2. Items Added/Removed → Item snapshot
├── 3. Payment Confirmed → Payment snapshot  
├── 4. Shipping Assigned → Delivery snapshot
├── 5. Order Completed → Final snapshot
├── 6. Order Cancelled → Cancellation snapshot
└── 7. Order Modified → Modification snapshot

Database Schema:
CREATE TABLE order_snapshots (
    id UUID PRIMARY KEY,
    order_id UUID REFERENCES orders(id),
    snapshot_type VARCHAR(50), -- 'created', 'item_change', 'payment', 'shipping', 'completed', 'cancelled'
    snapshot_data JSONB,       -- Complete order state at this moment
    previous_snapshot_id UUID REFERENCES order_snapshots(id),
    created_by_user_id UUID,   -- Who triggered this change
    created_at TIMESTAMP DEFAULT NOW(),
    
    -- Quick access fields (denormalized for performance)
    order_status VARCHAR(50),
    total_amount DECIMAL(12,2),
    item_count INT,
    customer_id UUID,
    delivery_method VARCHAR(50),
    
    INDEX idx_order_snapshots (order_id, created_at),
    INDEX idx_snapshot_type (snapshot_type, created_at),
    INDEX idx_customer_snapshots (customer_id, created_at)
);
```

---

## 🚚 **Shipping Service (8086) Flow**

### 📍 **Address-Based Delivery Assignment**
```
Smart Delivery Assignment Flow:
Order Creation → Shipping Service receives customer_address_id
├── 1. Get address details from Customer Service
├── 2. Check if province in self-delivery list (11 provinces)
├── 3. Get delivery route assignment from address
├── 4. Check customer delivery history/preferences
├── 5. Calculate delivery options & costs
├── 6. Return recommended delivery method with routes
└── 7. Cache delivery decision

Address Lookup Integration:
GET /api/shipping/delivery-options
{
  "customer_address_id": "addr_456",
  "address_details": {
    "province": "กรุงเทพมหานคร",
    "district": "บังกะปิ", 
    "subdistrict": "หัวหมาก",
    "delivery_route": "route_a"
  },
  "delivery_options": [
    {
      "method": "self_delivery",
      "route": "route_a",
      "vehicle_id": "truck_01", 
      "estimated_hours": 4,
      "delivery_fee": 50.00,
      "is_recommended": true,
      "reason": "Customer is in self-delivery area"
    }
  ]
}
```

---

## 💰 **Finance Service (8085) Flow - Enhanced**

### 💵 **Flexible Profit First Implementation**
```
Configurable Revenue Allocation Flow:
End of Day (6 PM) → Finance Service processes branch/vehicle sales
├── 1. Calculate total daily revenue per branch/vehicle
├── 2. Get flexible allocation percentages from configuration
├── 3. Apply Profit First allocations:
│   ├── X% → Profit Account
│   ├── Y% → Owner Pay Account  
│   ├── Z% → Tax Account
│   └── Remaining% → Available for expenses/transfers
├── 4. Wait for manual expense entries
├── 5. Process authorized transfers to suppliers/expenses
├── 6. Update cash flow records
├── 7. Generate end-of-day financial reports
└── 8. Alert management of cash positions
```

---

## 📊 **Analytics Service (8098) Flow**

### 📈 **Business Intelligence Processing**
```
Daily Analytics Pipeline:
Daily at 3 AM → Analytics Service
├── 1. Aggregate sales data from all sources
├── 2. Process customer analytics
├── 3. Calculate KPIs and metrics
├── 4. Generate trend analysis
├── 5. Update executive dashboards
├── 6. Prepare automated reports
├── 7. Detect anomalies
└── 8. Send insights to stakeholders

Real-time Dashboard Updates:
Transaction Completed → Analytics Service
├── 1. Update real-time sales counters
├── 2. Refresh product performance metrics
├── 3. Update customer acquisition stats
├── 4. Calculate hourly/daily targets
├── 5. Push updates to dashboard via WebSocket
├── 6. Trigger alerts if targets missed
└── 7. Log analytics events
```

---

## 🔔 **Notification Service (8092) Flow**

### 📱 **Multi-channel Communication**
```
Order Notification Flow:
Order Status Changed → Notification Service
├── 1. Determine notification type
├── 2. Get customer communication preferences
├── 3. Prepare message content
├── 4. Choose delivery channels:
│   ├── LINE Official Account
│   ├── Email  
│   ├── SMS
│   └── In-app notification
├── 5. Send notifications
├── 6. Track delivery status
├── 7. Handle failures/retries
└── 8. Log notification analytics
```

---

## 👥 **User Service (8088) Flow**

### 🔐 **Staff Authentication & Authorization**
```
Staff Login Flow:
Staff Login → POST /api/auth/login
├── 1. Validate credentials
├── 2. Check account status (active/suspended)
├── 3. Determine user role & permissions
├── 4. Generate JWT token
├── 5. Log login activity
├── 6. Update last login timestamp
├── 7. Cache session → Redis
└── 8. Return user profile + token
```

---

## 🔗 **API Gateway (8080) Flow**

### 🛡 **Request Routing & Security**
```
API Request Processing:
Client Request → API Gateway
├── 1. Validate API key/token
├── 2. Apply rate limiting
├── 3. Route to appropriate service
├── 4. Load balance requests
├── 5. Monitor response times
├── 6. Log request/response
├── 7. Apply security headers
└── 8. Return response to client
```

---

## 🎯 **Service Integration Summary**

### 🔄 **Primary Data Flow**
```
Customer Action → API Gateway → Core Service → Database
                              ↓
                          Kafka Event → Consumer Services
                              ↓ 
                      Cache Update → Real-time Updates
```

### 📊 **Service Dependencies**
```
High Priority (Core Business):
├── Order Service (8081) - Central orchestrator
├── Customer Service (8110) - Customer management
├── Payment Service (8087) - Payment processing
├── Loyverse Integration (8100) - External sync

Medium Priority (Operations):
├── Inventory Service (8082) - Stock management
├── Shipping Service (8086) - Delivery operations
├── Finance Service (8085) - Financial tracking

Low Priority (Intelligence):
├── Chat Service (8090) - AI conversations & chat-to-order
├── Analytics Service (8098) - Business intelligence
├── Reporting Service (8089) - Report generation
```

---

> 🚀 **Complete SAAN ecosystem with 15+ microservices working together efficiently!**