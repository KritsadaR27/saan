# 🎯 Complete SaaN Services Flow Diagrams

## 🏗 **System Architecture Overview**
```
🏪 Admin Dashboard (3010) ←→ 📱 Web App (3008) ←→ 🤖 Chatbot (8090)
                    ↓                ↓                     ↓
🛒 Order Service (8081) ← Central Orchestrator → 🔗 Webhook Listener (8091)
    ↓    ↓    ↓    ↓    ↓    ↓    ↓    ↓    ↓    ↓    ↓    ↓    ↓
[Product] [Customer] [Payment] [Inventory] [Shipping] [Finance] [AI] [Analytics] [Procurement] [Notification] [Reporting] [User]
   8083     8110      8087      8082       8086      8085    8097    8098        8099         8092          8089       8088
    ↓         ↓        ↓         ↓          ↓         ↓       ↓       ↓           ↓            ↓             ↓          ↓
PostgreSQL ←→ Redis ←→ Kafka ←→ Loyverse Integration (8091) ←→ Static CDN (8101) ←→ API Gateway (8080)
   5532      6379     9092
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
      "full_address": "123/45 ซอยสุขุมวิท 71 หัวหมาก บางกะปิ กรุงเทพฯ 10240",
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

### 📋 **Order History & Customer Journey**
```
Customer Order History Flow:
Admin/Customer requests history → GET /api/customers/{id}/orders
├── 1. Get orders with pagination
├── 2. Include order details & items
├── 3. Calculate order patterns
├── 4. Get delivery success rates
├── 5. Analyze payment preferences
├── 6. Generate insights
└── 7. Return comprehensive history

Database Queries:
-- Customer purchase patterns
SELECT 
  DATE_TRUNC('month', created_at) as month,
  COUNT(*) as orders,
  SUM(total_amount) as total_spent,
  AVG(total_amount) as avg_order_value
FROM orders 
WHERE customer_id = 'cust_123' 
GROUP BY month 
ORDER BY month DESC;

-- Top purchased products
SELECT 
  p.name,
  SUM(oi.quantity) as total_quantity,
  SUM(oi.total_price) as total_spent
FROM order_items oi
JOIN products p ON oi.product_id = p.id
JOIN orders o ON oi.order_id = o.id
WHERE o.customer_id = 'cust_123'
GROUP BY p.id, p.name
ORDER BY total_quantity DESC;
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
    "district": "บางกะปิ", 
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

-- No address parsing needed, all pre-populated from thai_addresses table
```

### 🚛 **Enhanced Route Planning & Multi-Carrier Management**
```
Multi-Delivery Method Planning Flow:
Daily at 10 PM (before delivery date) → Shipping Service

For Self-Delivery Orders:
├── 1. Get all pending COD orders for next day (self-delivery only)
├── 2. Group orders by delivery_route from addresses
├── 3. Optimize delivery sequence within each route
├── 4. Consider traffic patterns & delivery time windows
├── 5. Assign to available vehicles and drivers
├── 6. Calculate cash allocation needed per vehicle
├── 7. Generate delivery manifests with GPS coordinates
├── 8. Send routes to drivers via LINE/App
├── 9. Create vehicle tracking records
└── 10. Prepare COD cash collection reports

For Third-Party Carriers (Configurable Schedule):
├── 1. Check carrier pickup schedules:
│   ├── Inter Express: Mon, Wed, Fri (10 AM pickup)
│   ├── Lalamove: Daily (on-demand)
│   ├── Grab: Daily (on-demand)
│   ├── Flash: Mon-Sat (2 PM pickup)
│   ├── Nim Express: Tue, Thu, Sat (11 AM pickup)
├── 2. Group orders by carrier and pickup schedule
├── 3. Generate carrier shipping labels
├── 4. Prepare package manifest for each carrier
├── 5. Schedule carrier pickup requests
├── 6. Update order tracking with carrier info
├── 7. Send shipping notifications to customers
└── 8. Create carrier reconciliation records

Database Schema Updates:
CREATE TABLE delivery_carriers (
    id UUID PRIMARY KEY,
    carrier_name VARCHAR(100), -- 'inter_express', 'lalamove', etc.
    display_name VARCHAR(100), -- 'Inter Express', 'Lalamove'
    carrier_type VARCHAR(50),  -- 'scheduled', 'on_demand'
    pickup_schedule JSONB,     -- {"mon": "10:00", "wed": "10:00", "fri": "10:00"}
    pickup_address TEXT,       -- Where carrier picks up from
    contact_info JSONB,        -- Phone, email, API endpoints
    pricing_rules JSONB,       -- Distance-based pricing
    is_active BOOLEAN DEFAULT true,
    cutoff_time TIME,          -- Last order time for same-day pickup
    
    -- Integration settings
    api_endpoint VARCHAR(200),
    api_key VARCHAR(200),
    tracking_url_template VARCHAR(300)
);

CREATE TABLE delivery_carrier_schedules (
    id UUID PRIMARY KEY,
    carrier_id UUID REFERENCES delivery_carriers(id),
    schedule_date DATE,
    pickup_time TIME,
    cutoff_time TIME,          -- Orders must be ready by this time
    estimated_capacity INT,    -- Max packages per pickup
    current_load INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'open', -- 'open', 'full', 'picked_up'
    notes TEXT
);

-- Enhanced delivery tasks for multi-carrier
ALTER TABLE delivery_tasks ADD COLUMN carrier_id UUID REFERENCES delivery_carriers(id);
ALTER TABLE delivery_tasks ADD COLUMN carrier_tracking_number VARCHAR(100);
ALTER TABLE delivery_tasks ADD COLUMN pickup_scheduled_time TIMESTAMP;
ALTER TABLE delivery_tasks ADD COLUMN carrier_pickup_status VARCHAR(50);
```

### 📦 **Integrated Order Processing Flow**
```
Complete Order-to-Delivery Flow:
Order Payment Confirmed → Shipping Service
├── 1. Get customer default address
├── 2. Determine delivery method:
│   ├── Self-delivery provinces → Queue for 10 PM route planning
│   ├── Other provinces → Check carrier schedules
├── 3. For Self-Delivery:
│   ├── Add to route planning queue
│   ├── Estimate delivery date (next day)
│   ├── Prepare for COD collection
├── 4. For Third-Party Carriers:
│   ├── Check next available pickup schedule
│   ├── Calculate delivery timeline
│   ├── Generate shipping label
│   ├── Add to carrier pickup manifest
├── 5. Create delivery task record
├── 6. Update order with delivery info
├── 7. Send customer delivery notification
├── 8. Schedule tracking updates
└── 9. Integrate with Finance Service for payment tracking

Carrier-Specific Processing:

A. Scheduled Carriers (Inter Express, Flash, Nim Express):
├── Check next pickup date from schedule
├── Validate order cutoff time
├── Add to pickup manifest
├── Generate carrier-specific labels
├── Schedule pickup request
└── Provide tracking number to customer

B. On-Demand Carriers (Lalamove, Grab):
├── Calculate immediate availability
├── Get real-time pricing
├── Book delivery slot via carrier API
├── Generate booking reference
├── Track real-time status
└── Handle dynamic pricing updates

C. Self-Delivery:
├── Queue for route optimization (10 PM)
├── Assign vehicle and driver
├── Prepare COD cash requirements
├── Generate delivery manifest
├── Enable GPS tracking
└── Schedule driver notifications

Database Integration:
-- Link orders to delivery planning
CREATE TABLE order_delivery_planning (
    id UUID PRIMARY KEY,
    order_id UUID REFERENCES orders(id),
    delivery_method VARCHAR(50),
    carrier_id UUID REFERENCES delivery_carriers(id),
    planned_delivery_date DATE,
    delivery_route VARCHAR(50),
    vehicle_id UUID,
    carrier_schedule_id UUID,
    planning_status VARCHAR(50), -- 'queued', 'planned', 'dispatched'
    created_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_delivery_date_method (planned_delivery_date, delivery_method)
);

-- Carrier pickup manifests
CREATE TABLE carrier_pickup_manifests (
    id UUID PRIMARY KEY,
    carrier_id UUID,
    pickup_date DATE,
    pickup_time TIME,
    total_packages INT,
    total_weight DECIMAL(8,2),
    manifest_status VARCHAR(50), -- 'preparing', 'ready', 'picked_up'
    pickup_reference VARCHAR(100), -- Carrier's pickup reference
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE carrier_manifest_items (
    id UUID PRIMARY KEY,
    manifest_id UUID REFERENCES carrier_pickup_manifests(id),
    order_id UUID,
    package_weight DECIMAL(6,2),
    package_dimensions VARCHAR(50), -- "20x15x10 cm"
    delivery_address TEXT,
    recipient_phone VARCHAR(20),
    cod_amount DECIMAL(10,2),
    carrier_tracking_number VARCHAR(100)
);
```

### ⏰ **Dynamic Scheduling System**
```
Carrier Schedule Management:
Admin configures carrier schedules → PUT /api/shipping/carriers/{id}/schedule
{
  "carrier_id": "inter_express",
  "weekly_schedule": {
    "monday": {"pickup_time": "10:00", "cutoff_time": "09:00", "capacity": 50},
    "wednesday": {"pickup_time": "10:00", "cutoff_time": "09:00", "capacity": 50}, 
    "friday": {"pickup_time": "10:00", "cutoff_time": "09:00", "capacity": 50}
  },
  "special_schedules": [
    {
      "date": "2025-07-15",
      "pickup_time": "14:00", 
      "cutoff_time": "13:00",
      "capacity": 30,
      "note": "Holiday adjusted schedule"
    }
  ]
}

Real-time Order Assignment:
Order Confirmed (3 PM Tuesday) → Shipping Service
├── Check carrier schedules for delivery address province
├── Inter Express: Next pickup is Wednesday 10 AM
├── Check cutoff time: 9 AM Wednesday (still open)
├── Check capacity: 15/50 packages (available)
├── Assign to Wednesday Inter Express pickup
├── Generate shipping label
├── Add to pickup manifest
├── Notify customer: "จัดส่งวันพุธที่ 3 กรกฎาคม"
└── Schedule pickup confirmation

Automatic Carrier Selection:
If primary carrier full/unavailable:
├── 1. Check alternative carriers for same route
├── 2. Compare delivery times and costs
├── 3. Select best available option
├── 4. Notify customer of delivery method change
├── 5. Update order delivery details
└── 6. Log carrier selection reasoning
```

### 📱 **Integrated Tracking System**
```
Unified Delivery Tracking:
Customer checks delivery → GET /api/orders/{id}/tracking
{
  "order_id": "order_123",
  "delivery_method": "inter_express",
  "carrier_name": "Inter Express",
  "tracking_number": "IE2025070301234",
  "delivery_status": "in_transit",
  "estimated_delivery": "2025-07-04 16:00",
  "tracking_history": [
    {
      "timestamp": "2025-07-03 10:15",
      "status": "picked_up", 
      "location": "ร้านลุงรวยหลัก",
      "description": "สินค้าถูกรับจากร้านค้าแล้ว"
    },
    {
      "timestamp": "2025-07-03 14:30",
      "status": "in_transit",
      "location": "Hub กรุงเทพ", 
      "description": "สินค้าอยู่ระหว่างการขนส่ง"
    }
  ],
  "self_delivery_details": null, // Only for self-delivery orders
  "carrier_contact": {
    "phone": "02-123-4567",
    "website": "https://interexpress.co.th/tracking"
  }
}

Self-Delivery Tracking:
{
  "delivery_method": "self_delivery",
  "vehicle_info": {
    "vehicle_id": "truck_01",
    "driver_name": "คุณสมชาย",
    "driver_phone": "+66812345678",
    "license_plate": "1กก-1234"
  },
  "route_info": {
    "route_name": "เส้น A - บางกะปิ",
    "delivery_sequence": 3,
    "estimated_arrival": "2025-07-04 14:30"
  },
  "live_tracking": {
    "current_location": "13.7563,100.5018",
    "distance_to_destination": "2.5 km",
    "eta_updated": "2025-07-04 13:45"
  }
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

Database Schema:
CREATE TABLE profit_allocation_rules (
    id UUID PRIMARY KEY,
    branch_id UUID, -- NULL for global rules
    vehicle_id UUID, -- NULL for global rules
    profit_percentage DECIMAL(5,2) DEFAULT 30.00,
    owner_pay_percentage DECIMAL(5,2) DEFAULT 20.00,
    tax_percentage DECIMAL(5,2) DEFAULT 15.00,
    effective_from DATE,
    effective_to DATE,
    is_active BOOLEAN DEFAULT true,
    updated_by_user_id UUID
);

CREATE TABLE daily_cash_summary (
    id UUID PRIMARY KEY,
    business_date DATE,
    branch_id UUID,
    vehicle_id UUID,
    opening_cash DECIMAL(12,2),
    total_sales DECIMAL(12,2),
    cod_collections DECIMAL(12,2),
    
    -- Profit First allocations
    profit_allocation DECIMAL(12,2),
    owner_pay_allocation DECIMAL(12,2), 
    tax_allocation DECIMAL(12,2),
    available_for_expenses DECIMAL(12,2),
    
    -- Manual entries
    manual_expenses DECIMAL(12,2),
    supplier_transfers DECIMAL(12,2),
    other_transfers DECIMAL(12,2),
    
    closing_cash DECIMAL(12,2),
    reconciled BOOLEAN DEFAULT false,
    reconciled_by_user_id UUID,
    reconciled_at TIMESTAMP
);
```

### 🏪 **Multi-Branch Multi-Transfer Management**
```
Branch End-of-Day Process:
6 PM Daily → Branch Manager/Staff enters expenses
├── 1. System shows calculated revenue & allocations
├── 2. Staff enters actual expenses:
│   ├── Fuel costs
│   ├── Staff meals
│   ├── Utilities
│   ├── Maintenance
│   └── Other operational costs
├── 3. System calculates available cash for transfers
├── 4. Manager/Head Office authorizes transfers:
│   ├── Supplier A payment: 15,000 THB
│   ├── Supplier B payment: 8,500 THB  
│   ├── Rent payment: 12,000 THB
│   ├── Central office transfer: 25,000 THB
│   └── Emergency reserve: 5,000 THB
├── 5. Execute authorized bank transfers
├── 6. Update cash flow records
├── 7. Generate reconciliation reports
└── 8. Prepare next day opening balance

Vehicle End-of-Day Process:
Driver returns (8 PM) → Vehicle cash reconciliation
├── 1. Driver reports cash collected & expenses
├── 2. System compares with expected collections
├── 3. Record actual expenses:
│   ├── Fuel: 800 THB
│   ├── Tolls: 120 THB
│   ├── Meals: 200 THB
│   ├── Vehicle maintenance: 500 THB
├── 4. Calculate net cash to transfer
├── 5. Manager authorizes vehicle transfers:
│   ├── Supplier payment: 5,000 THB
│   ├── Central deposit: 15,000 THB
│   ├── Next day cash: 2,000 THB
├── 6. Process transfers
├── 7. Update vehicle cash tracking
└── 8. Prepare next day allocation

Multiple Transfer Execution:
CREATE TABLE cash_transfer_batches (
    id UUID PRIMARY KEY,
    source_branch_id UUID,
    source_vehicle_id UUID,
    batch_date DATE,
    total_amount DECIMAL(12,2),
    authorized_by_user_id UUID,
    status VARCHAR(20), -- 'pending', 'processing', 'completed'
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE cash_transfers (
    id UUID PRIMARY KEY,
    batch_id UUID REFERENCES cash_transfer_batches(id),
    transfer_type VARCHAR(50), -- 'supplier_payment', 'expense_payment', 'central_transfer'
    recipient_id UUID, -- supplier_id or account_id
    recipient_name VARCHAR(200),
    amount DECIMAL(12,2),
    reference_number VARCHAR(100),
    bank_account VARCHAR(100),
    purpose TEXT,
    status VARCHAR(20),
    executed_at TIMESTAMP
);
```

### 📊 **Real-time Cash Flow Dashboard**
```
Multi-Entity Cash Management:
Real-time Cash Status → GET /api/finance/cash-status
{
  "summary": {
    "total_system_cash": 125000.00,
    "available_for_transfer": 45000.00,
    "pending_transfers": 15000.00,
    "emergency_reserves": 10000.00
  },
  "branches": [
    {
      "branch_id": "store_main",
      "name": "ร้านลุงรวยหลัก", 
      "current_cash": 35000.00,
      "today_sales": 12000.00,
      "today_expenses": 2500.00,
      "pending_transfers": [
        {"supplier": "ซัพพลายเออร์ A", "amount": 8000, "status": "authorized"},
        {"type": "central_transfer", "amount": 15000, "status": "pending"}
      ]
    }
  ],
  "vehicles": [
    {
      "vehicle_id": "truck_01",
      "name": "รถส่งของ A",
      "driver": "คุณสมชาย",
      "current_cash": 8500.00,
      "today_collections": 6200.00,
      "today_expenses": 800.00,
      "pending_transfers": [
        {"supplier": "ซัพพลายเออร์ B", "amount": 3000, "status": "pending"}
      ]
    }
  ]
}
```

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

-- Order audit trail
CREATE TABLE order_audit_log (
    id UUID PRIMARY KEY,
    order_id UUID,
    snapshot_id UUID REFERENCES order_snapshots(id),
    action VARCHAR(100),       -- 'order_created', 'item_added', 'payment_confirmed'
    changed_fields JSONB,      -- What specifically changed
    old_values JSONB,         -- Previous values
    new_values JSONB,         -- New values
    user_id UUID,             -- Who made the change
    ip_address VARCHAR(45),   -- Where the change came from
    user_agent TEXT,          -- How the change was made
    created_at TIMESTAMP DEFAULT NOW()
);
```

### 📸 **Snapshot Creation Process**
```
Automatic Snapshot Triggers:

1. Order Creation Snapshot:
POST /api/orders/create → Order Service
├── Create initial order record
├── Generate snapshot with type 'created'
├── Store complete order state: customer, addresses, initial items
├── Log audit entry: 'order_created'
└── Cache snapshot → Redis

2. Item Change Snapshot:  
POST /api/orders/{id}/items → Order Service
├── Compare current items with previous state
├── Create snapshot with type 'item_change'
├── Store: added_items, removed_items, quantity_changes
├── Update order totals
├── Log detailed item changes
└── Update cached order state

3. Payment Confirmation Snapshot:
Payment Confirmed → Order Service
├── Create snapshot with type 'payment'
├── Store: payment_method, amount, payment_details
├── Lock order for modifications
├── Log payment confirmation details
└── Trigger downstream processes

4. Shipping Assignment Snapshot:
Shipping Service assigns delivery → Order Service  
├── Create snapshot with type 'shipping'
├── Store: delivery_method, carrier, vehicle, estimated_delivery
├── Log shipping assignment
└── Update customer notifications

5. Order Completion Snapshot:
Loyverse Receipt Created → Order Service
├── Create snapshot with type 'completed'
├── Store: loyverse_receipt_id, completion_time, final_amounts
├── Lock order permanently
├── Archive order data
└── Trigger analytics processing

Snapshot Data Structure:
{
  "snapshot_id": "snap_123",
  "order_id": "order_456", 
  "snapshot_type": "item_change",
  "timestamp": "2025-07-02T14:30:00Z",
  "triggered_by": {
    "user_id": "staff_789",
    "action": "added_item",
    "source": "admin_dashboard"
  },
  "order_state": {
    "status": "draft",
    "customer": {
      "id": "cust_123",
      "name": "นายสมชาย ใจดี",
      "tier": "silver"
    },
    "items": [
      {
        "product_id": "prod_789",
        "variant_id": "var_456", 
        "name": "โค้ก 325ml",
        "quantity": 2,
        "unit_price": 15.00,
        "total_price": 30.00,
        "added_at": "2025-07-02T14:30:00Z"
      }
    ],
    "pricing": {
      "subtotal": 30.00,
      "discount": 1.50,
      "delivery_fee": 50.00,
      "total": 78.50
    },
    "delivery": {
      "method": "self_delivery",
      "address_id": "addr_789",
      "route": "route_a"
    }
  },
  "changes": {
    "action": "item_added",
    "item_details": {
      "product_id": "prod_789",
      "quantity": 2,
      "reason": "customer_request"
    }
  }
}
```

### 🔍 **Order History & Analytics**
```
Order History Retrieval:
GET /api/orders/{id}/history → Order Service
├── 1. Get all snapshots for order (chronological)
├── 2. Build timeline of changes
├── 3. Calculate duration between status changes
├── 4. Identify modification patterns
├── 5. Generate change summary
└── 6. Return comprehensive history

Response Example:
{
  "order_id": "order_456",
  "current_status": "completed",
  "total_duration": "2 hours 45 minutes",
  "timeline": [
    {
      "timestamp": "2025-07-02T12:00:00Z",
      "action": "Order Created",
      "user": "staff_789",
      "details": "Initial order with 0 items",
      "duration_since_previous": null
    },
    {
      "timestamp": "2025-07-02T12:15:00Z", 
      "action": "Items Added",
      "user": "staff_789",
      "details": "Added 2x โค้ก 325ml",
      "duration_since_previous": "15 minutes"
    },
    {
      "timestamp": "2025-07-02T12:30:00Z",
      "action": "Payment Confirmed", 
      "user": "system",
      "details": "Bank transfer confirmed - 78.50 THB",
      "duration_since_previous": "15 minutes"
    },
    {
      "timestamp": "2025-07-02T14:45:00Z",
      "action": "Order Completed",
      "user": "system", 
      "details": "Loyverse receipt created",
      "duration_since_previous": "2 hours 15 minutes"
    }
  ],
  "statistics": {
    "total_modifications": 3,
    "average_time_per_stage": "41 minutes",
    "staff_efficiency_score": 8.5
  }
}

Order Analytics Queries:
-- Order completion time analysis
SELECT 
  DATE(created_at) as date,
  AVG(EXTRACT(EPOCH FROM (completed_at - created_at))/60) as avg_completion_minutes,
  COUNT(*) as total_orders
FROM order_snapshots 
WHERE snapshot_type = 'completed'
GROUP BY DATE(created_at);

-- Staff performance analysis  
SELECT 
  created_by_user_id,
  COUNT(*) as orders_processed,
  AVG(item_count) as avg_items_per_order,
  AVG(total_amount) as avg_order_value
FROM order_snapshots 
WHERE snapshot_type = 'created'
  AND created_at >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY created_by_user_id;

-- Customer ordering patterns
SELECT 
  customer_id,
  COUNT(DISTINCT DATE(created_at)) as active_days,
  COUNT(*) as total_orders,
  AVG(total_amount) as avg_order_value,
  MAX(created_at) as last_order
FROM order_snapshots 
WHERE snapshot_type = 'completed'
GROUP BY customer_id
HAVING COUNT(*) > 1;
```

### 🔄 **Snapshot-Based Recovery**
```
Order Recovery System:
If order processing fails → Order Service recovery
├── 1. Get latest valid snapshot
├── 2. Compare with current order state
├── 3. Identify inconsistencies
├── 4. Restore to last known good state
├── 5. Log recovery action
├── 6. Notify admin of recovery
└── 7. Resume processing from recovered state

Order Rollback Capability:
Admin requests rollback → POST /api/orders/{id}/rollback
├── 1. Validate rollback permissions
├── 2. Get target snapshot (specific point in time)
├── 3. Check if rollback is safe (no payments completed)
├── 4. Create rollback snapshot
├── 5. Restore order to target state
├── 6. Invalidate newer snapshots
├── 7. Log rollback reason
└── 8. Notify stakeholders

Snapshot Cleanup:
Weekly cleanup job → Order Service
├── 1. Archive completed order snapshots > 90 days
├── 2. Compress old snapshot data
├── 3. Maintain audit trail for legal requirements
├── 4. Clean up Redis cached snapshots
└── 5. Generate cleanup report
```

### 📊 **Real-time Order Monitoring**
```
Order Dashboard Updates:
Real-time snapshot events → WebSocket broadcasts
├── 1. New order created → Update pending orders count
├── 2. Payment confirmed → Update revenue counters
├── 3. Order completed → Update completion metrics
├── 4. Order cancelled → Update cancellation stats
└── 5. System issues → Alert dashboard

Snapshot-based Alerts:
├── Order stuck in draft > 30 minutes → Alert staff
├── Payment pending > 2 hours → Alert finance team
├── Shipping not assigned > 1 hour → Alert logistics
├── Unusual modification patterns → Alert management
└── High cancellation rate → Alert business team

Redis Caching Strategy:
-- Active order snapshots (hot data)
order:snapshot:latest:{order_id} → Current snapshot (30min TTL)
order:snapshot:history:{order_id} → Recent snapshots (1hr TTL)

-- Order analytics (warm data)  
order:stats:daily:{date} → Daily order statistics (24hr TTL)
order:stats:staff:{user_id} → Staff performance (1hr TTL)
```

### 🧠 **Customer Behavior Analysis**
```
Customer Intelligence Pipeline:
Daily at 2 AM → AI Service processes customer data
├── 1. Analyze purchase patterns
├── 2. Calculate customer lifetime value
├── 3. Predict churn probability
├── 4. Generate product recommendations
├── 5. Identify cross-selling opportunities
├── 6. Segment customers by behavior
├── 7. Update recommendation models
└── 8. Generate insights for business team

Real-time Recommendation Engine:
Customer adds item to cart → AI Service
├── 1. Get customer purchase history
├── 2. Analyze product associations
├── 3. Apply collaborative filtering
├── 4. Consider seasonal trends
├── 5. Calculate recommendation scores
├── 6. Filter by inventory availability
├── 7. Return ranked recommendations
└── 8. Log recommendation performance

Database Schema:
CREATE TABLE customer_behavior_profiles (
    customer_id UUID PRIMARY KEY,
    purchase_frequency DECIMAL(5,2),
    average_order_value DECIMAL(10,2),
    preferred_categories JSONB,
    churn_risk_score DECIMAL(3,2),
    lifetime_value_prediction DECIMAL(12,2),
    last_analyzed TIMESTAMP
);

CREATE TABLE ai_recommendations (
    id UUID PRIMARY KEY,
    customer_id UUID,
    recommended_products JSONB,
    recommendation_type VARCHAR(50), -- 'cross_sell', 'upsell', 'reorder'
    confidence_score DECIMAL(3,2),
    generated_at TIMESTAMP,
    clicked BOOLEAN,
    purchased BOOLEAN
);
```

### 📈 **Demand Forecasting**
```
Inventory Forecasting Pipeline:
Weekly on Sunday → AI Service
├── 1. Analyze historical sales data
├── 2. Factor in seasonal patterns
├── 3. Consider external factors (weather, events)
├── 4. Calculate demand predictions
├── 5. Generate reorder recommendations
├── 6. Optimize stock levels
├── 7. Alert Procurement Service
└── 8. Update forecasting models

Smart Pricing Optimization:
Daily at 6 AM → AI Service
├── 1. Analyze competitor pricing
├── 2. Consider demand elasticity
├── 3. Factor in inventory levels
├── 4. Calculate optimal pricing
├── 5. Simulate profit scenarios
├── 6. Recommend price adjustments
├── 7. A/B test pricing strategies
└── 8. Monitor pricing performance
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

Database Schema:
CREATE TABLE daily_sales_summary (
    date DATE PRIMARY KEY,
    total_revenue DECIMAL(15,2),
    total_orders INT,
    average_order_value DECIMAL(10,2),
    new_customers INT,
    returning_customers INT,
    top_selling_products JSONB,
    profit_margin DECIMAL(5,2)
);

CREATE TABLE customer_analytics (
    customer_id UUID,
    analysis_date DATE,
    orders_count INT,
    total_spent DECIMAL(12,2),
    avg_order_value DECIMAL(10,2),
    days_since_last_order INT,
    customer_segment VARCHAR(50), -- 'vip', 'regular', 'at_risk'
    PRIMARY KEY (customer_id, analysis_date)
);
```

---

## 🛒 **Procurement Service (8099) Flow**

### 📋 **Purchase Order Management**
```
Automatic Reorder Flow:
AI Service detects low stock → Procurement Service
├── 1. Validate reorder recommendations
├── 2. Check supplier availability
├── 3. Compare supplier prices
├── 4. Calculate optimal order quantities
├── 5. Generate purchase orders
├── 6. Send PO to suppliers
├── 7. Schedule delivery
└── 8. Update procurement status

Supplier Management Flow:
New Supplier Registration → POST /api/suppliers
├── 1. Validate supplier information
├── 2. Set up payment terms
├── 3. Configure product catalog
├── 4. Establish pricing agreements
├── 5. Set up communication channels
├── 6. Create performance tracking
├── 7. Integrate with finance for payments
└── 8. Activate supplier account

Database Schema:
CREATE TABLE suppliers (
    id UUID PRIMARY KEY,
    name VARCHAR(200),
    contact_info JSONB,
    payment_terms VARCHAR(100), -- 'net_30', 'net_15', 'cod'
    performance_rating DECIMAL(3,2),
    preferred_supplier BOOLEAN,
    active BOOLEAN DEFAULT true
);

CREATE TABLE purchase_orders (
    id UUID PRIMARY KEY,
    supplier_id UUID,
    order_date DATE,
    expected_delivery DATE,
    total_amount DECIMAL(12,2),
    status po_status_enum,
    payment_status payment_status_enum,
    items JSONB -- Order line items
);
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

Bulk Marketing Campaign Flow:
Marketing Campaign → POST /api/notifications/campaign
├── 1. Define target customer segments
├── 2. Validate message content
├── 3. Schedule send time
├── 4. Prepare personalized messages
├── 5. Execute batch sending
├── 6. Monitor delivery rates
├── 7. Track engagement metrics
└── 8. Generate campaign reports

Database Schema:
CREATE TABLE notification_templates (
    id UUID PRIMARY KEY,
    template_name VARCHAR(100),
    channel VARCHAR(50), -- 'line', 'email', 'sms'
    content_template TEXT,
    variables JSONB -- Dynamic content variables
);

CREATE TABLE notification_logs (
    id UUID PRIMARY KEY,
    customer_id UUID,
    notification_type VARCHAR(50),
    channel VARCHAR(50),
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    opened_at TIMESTAMP,
    clicked BOOLEAN,
    status VARCHAR(20) -- 'sent', 'delivered', 'failed'
);
```

---

## 📋 **Reporting Service (8089) Flow**

### 📊 **Executive Report Generation**
```
Automated Report Generation:
Scheduled Reports → Reporting Service
├── 1. Query data from multiple services
├── 2. Aggregate business metrics
├── 3. Apply report templates
├── 4. Generate visualizations
├── 5. Create PDF/Excel reports
├── 6. Send to stakeholders
├── 7. Archive report versions
└── 8. Schedule next generation

Ad-hoc Report Requests:
Manager requests custom report → POST /api/reports/custom
├── 1. Validate report parameters
├── 2. Query relevant data sources
├── 3. Apply filters and grouping
├── 4. Generate dynamic visualization
├── 5. Export in requested format
├── 6. Provide download link
├── 7. Log report usage
└── 8. Cache for future requests

Report Types:
├── Daily Sales Report (8 AM daily)
├── Weekly Performance Report (Monday 9 AM)
├── Monthly Financial Report (1st of month)
├── Customer Analytics Report (Weekly)
├── Inventory Movement Report (Daily)
├── Staff Performance Report (Weekly)
└── Executive Dashboard (Real-time)
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

Permission Management Flow:
Admin updates staff permissions → PUT /api/users/{id}/permissions
├── 1. Validate admin privileges
├── 2. Update role assignments
├── 3. Refresh permission cache
├── 4. Log permission changes
├── 5. Notify affected user
├── 6. Update audit trail
└── 7. Force token refresh

Database Schema:
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(50) UNIQUE,
    email VARCHAR(100),
    phone VARCHAR(20),
    role VARCHAR(50), -- 'admin', 'manager', 'staff', 'driver'
    permissions JSONB,
    active BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE user_sessions (
    session_id VARCHAR(100) PRIMARY KEY,
    user_id UUID,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP,
    expires_at TIMESTAMP
);
```

---

## 🌐 **Static CDN Service (8101) Flow**

### 📁 **Asset Management**
```
Image Upload Flow:
Product Image Upload → POST /api/cdn/upload
├── 1. Validate file type & size
├── 2. Generate unique filename
├── 3. Optimize image (resize, compress)
├── 4. Upload to cloud storage
├── 5. Generate CDN URLs
├── 6. Update product record
├── 7. Invalidate cache
└── 8. Return new image URLs

Cache Management:
Static Asset Request → CDN Service
├── 1. Check cache headers
├── 2. Serve from cache if valid
├── 3. If expired, fetch from origin
├── 4. Apply compression
├── 5. Set cache headers
├── 6. Serve to client
└── 7. Log access statistics
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

Rate Limiting & Throttling:
├── Free tier: 100 requests/hour
├── Paid tier: 1000 requests/hour  
├── Admin: Unlimited
├── Burst allowance: 2x rate for 5 minutes
└── Block after limit exceeded
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
├── Loyverse Integration (8091) - External sync

Medium Priority (Operations):
├── Inventory Service (8082) - Stock management
├── Shipping Service (8086) - Delivery operations
├── Finance Service (8085) - Financial tracking

Low Priority (Intelligence):
├── AI Service (8097) - Recommendations & insights
├── Analytics Service (8098) - Business intelligence
├── Reporting Service (8089) - Report generation
```

---

> 🚀 **Complete SaaN ecosystem dengan 15+ microservices ที่ทำงานร่วมกันอย่างมีประสิทธิภาพ!**