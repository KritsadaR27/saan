# Finance Service API Documentation

## 🚀 **Base Information**
- **Service Name**: Finance Service  
- **Port**: 8088
- **Base URL**: `http://finance:8088`
- **Module Name**: `finance`
- **Health Check**: `GET /health`

---

## 📝 **API Endpoints**

### **Daily Cash Flow Management**

#### Create Daily Cash Flow
```http
POST /api/v1/finance/cash-flows
Content-Type: application/json

{
  "business_date": "2024-01-15",
  "location_id": "location_123",
  "location_type": "branch|vehicle",
  "location_name": "Branch Central",
  "loyverse_receipts_total": 15000.00,
  "delivery_orders_total": 8500.00,
  "other_income": 500.00,
  "cash_on_hand": 2000.00,
  "cash_reason": "Float money for tomorrow",
  "notes": "Busy day, high sales"
}
```

**Response:**
```json
{
  "id": "cash_flow_456",
  "business_date": "2024-01-15",
  "location_id": "location_123",
  "location_type": "branch",
  "location_name": "Branch Central",
  "total_sales": 24000.00,
  "loyverse_receipts_total": 15000.00,
  "delivery_orders_total": 8500.00,
  "other_income": 500.00,
  "total_expenses": 0.00,
  "cash_on_hand": 2000.00,
  "transfer_to_company": 0.00,
  "status": "draft",
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### Get Cash Flow Details
```http
GET /api/v1/finance/cash-flows/{cash_flow_id}
```

#### Update Cash Flow
```http
PUT /api/v1/finance/cash-flows/{cash_flow_id}
Content-Type: application/json

{
  "supplier_transfers": 5000.00,
  "operational_expenses": 1500.00,
  "fixed_expenses": 2000.00,
  "transfer_to_company": 15000.00,
  "transfer_proof_url": "https://cdn.saan.co/slips/transfer_001.jpg"
}
```

#### Submit Cash Flow for Approval
```http
POST /api/v1/finance/cash-flows/{cash_flow_id}/submit
```

#### Approve Cash Flow
```http
POST /api/v1/finance/cash-flows/{cash_flow_id}/approve
Content-Type: application/json

{
  "notes": "Approved - all documents verified"
}
```

#### List Cash Flows
```http
GET /api/v1/finance/cash-flows?location_id={id}&status={status}&date_from={date}&date_to={date}&limit={n}&offset={n}
```

---

### **Expense Management**

#### Record Expense
```http
POST /api/v1/finance/expenses
Content-Type: multipart/form-data

{
  "cash_flow_id": "cash_flow_456",
  "expense_category": "supplier|operational|fixed|personal",
  "expense_subcategory": "fuel|salary|rent|utilities|food",
  "description": "Fuel for delivery truck",
  "amount": 1500.00,
  "expense_date": "2024-01-15",
  "vendor_name": "PTT Station",
  "receipt_number": "INV-001234",
  "payment_method": "cash|transfer|card",
  "receipt_file": [receipt image file]
}
```

**Response:**
```json
{
  "id": "expense_789",
  "cash_flow_id": "cash_flow_456",
  "expense_category": "operational",
  "expense_subcategory": "fuel",
  "description": "Fuel for delivery truck",
  "amount": 1500.00,
  "expense_date": "2024-01-15",
  "receipt_url": "https://cdn.saan.co/receipts/expense_789.jpg",
  "vendor_name": "PTT Station",
  "status": "pending",
  "created_at": "2024-01-15T14:30:00Z"
}
```

#### Approve Expense
```http
POST /api/v1/finance/expenses/{expense_id}/approve
Content-Type: application/json

{
  "notes": "Receipt verified, amount correct"
}
```

#### Reject Expense
```http
POST /api/v1/finance/expenses/{expense_id}/reject
Content-Type: application/json

{
  "reason": "Receipt unclear, need better image"
}
```

#### List Expenses
```http
GET /api/v1/finance/expenses?cash_flow_id={id}&category={category}&status={status}&limit={n}&offset={n}
```

---

### **Transfer Management**

#### Plan Transfer
```http
POST /api/v1/finance/transfers/plan
Content-Type: application/json

{
  "planned_date": "2024-01-16",
  "from_location_id": "location_123",
  "from_location_type": "branch",
  "to_location_id": "company_main",
  "transfer_type": "company_return|supplier_payment|inter_location",
  "description": "Daily sales transfer to company",
  "planned_amount": 15000.00,
  "priority_level": "high|medium|low"
}
```

#### Execute Transfer
```http
POST /api/v1/finance/transfers/{transfer_id}/execute
Content-Type: multipart/form-data

{
  "executed_amount": 15000.00,
  "executed_date": "2024-01-16",
  "transfer_slip": [transfer slip image]
}
```

#### Verify Bank Transfer
```http
POST /api/v1/finance/transfers/verify
Content-Type: multipart/form-data

{
  "transfer_date": "2024-01-16",
  "amount": 15000.00,
  "from_account": "Kasikorn ***1234",
  "to_account": "SCB ***5678",
  "transfer_reference": "TXN123456789",
  "slip_image": [bank slip image],
  "related_cash_flow_id": "cash_flow_456"
}
```

---

### **Profit First Management**

#### Configure Profit First
```http
POST /api/v1/finance/profit-first/config
Content-Type: application/json

{
  "location_id": "location_123",
  "location_type": "branch",
  "profit_percentage": 5.00,
  "owner_pay_percentage": 10.00,
  "tax_percentage": 15.00,
  "operating_percentage": 70.00,
  "allocation_frequency": "daily",
  "auto_allocate": true,
  "minimum_allocation_amount": 1000.00,
  "effective_from": "2024-01-01"
}
```

#### Execute Profit Allocation
```http
POST /api/v1/finance/profit-first/allocate
Content-Type: application/json

{
  "business_date": "2024-01-15",
  "location_id": "location_123",
  "force_allocation": false
}
```

#### Get Allocation Details
```http
GET /api/v1/finance/profit-first/allocations/{allocation_id}
```

**Response:**
```json
{
  "id": "allocation_123",
  "business_date": "2024-01-15",
  "location_id": "location_123",
  "total_revenue": 24000.00,
  "allocatable_revenue": 15500.00,
  "profit_amount": 775.00,
  "owner_pay_amount": 1550.00,
  "tax_amount": 2325.00,
  "operating_amount": 10850.00,
  "profit_allocated": true,
  "owner_pay_allocated": false,
  "tax_allocated": false,
  "operating_allocated": false,
  "allocated_at": "2024-01-15T18:00:00Z"
}
```

---

### **Loyverse Integration**

#### Process Loyverse Receipts
```http
POST /api/v1/finance/loyverse/sync
Content-Type: application/json

{
  "date": "2024-01-15",
  "store_ids": ["store_1", "store_2"]
}
```

#### Manual Receipt Allocation
```http
PUT /api/v1/finance/loyverse/receipts/{receipt_id}/allocate
Content-Type: application/json

{
  "allocated_to_location_id": "location_123",
  "allocated_to_location_type": "vehicle",
  "allocation_reason": "Manual review - delivery note indicates vehicle 3"
}
```

#### Get Unallocated Receipts
```http
GET /api/v1/finance/loyverse/receipts/unallocated?date={date}&limit={n}&offset={n}
```

---

### **Financial Reporting**

#### Generate Daily Report
```http
GET /api/v1/finance/reports/daily?location_id={id}&date={date}
```

**Response:**
```json
{
  "date": "2024-01-15",
  "location_id": "location_123",
  "location_name": "Branch Central",
  "location_type": "branch",
  "revenue": {
    "total_sales": 24000.00,
    "loyverse_receipts": 15000.00,
    "delivery_orders": 8500.00,
    "other_income": 500.00,
    "revenue_breakdown": {
      "walk_in_customers": 60.0,
      "delivery_orders": 35.0,
      "other": 5.0
    }
  },
  "expenses": {
    "supplier_transfers": 5000.00,
    "operational_expenses": 1500.00,
    "fixed_expenses": 2000.00,
    "personal_expenses": 500.00,
    "total_expenses": 9000.00,
    "expense_details": [
      {
        "category": "operational",
        "subcategory": "fuel",
        "amount": 1500.00,
        "count": 3
      }
    ]
  },
  "cash_flow": {
    "cash_on_hand": 2000.00,
    "transfer_to_company": 15000.00,
    "net_cash_flow": 15000.00
  },
  "profit_first": {
    "allocatable_revenue": 15000.00,
    "profit_amount": 750.00,
    "owner_pay_amount": 1500.00,
    "tax_amount": 2250.00,
    "operating_amount": 10500.00
  },
  "status": "completed",
  "generated_at": "2024-01-16T08:00:00Z"
}
```

#### Generate Monthly Report
```http
GET /api/v1/finance/reports/monthly?location_id={id}&year={year}&month={month}
```

#### Executive Dashboard
```http
GET /api/v1/finance/dashboard?date={date}
```

#### Export to Excel
```http
GET /api/v1/finance/reports/export/excel?type=daily&location_id={id}&date={date}
```

---

## 🔗 **Integration Points**

### **Outbound Calls (Services this service calls)**

#### Order Service (8081)
```go
// Get delivery orders for revenue allocation
GET http://order:8081/api/v1/orders?delivery_date={date}&vehicle_id={id}
GET http://order:8081/api/v1/orders/{id}
```

#### Customer Service (8110)
```go
// Get location information
GET http://customer:8110/api/v1/locations/{id}
GET http://customer:8110/api/v1/locations
```

#### Payment Service (8087)
```go
// Get payment confirmation details
GET http://payment:8087/api/v1/payments/{id}
GET http://payment:8087/api/v1/payments/order/{order_id}
```

#### Loyverse Integration (8091)
```go
// Sync receipts from Loyverse
GET http://loyverse:8091/api/v1/receipts?date={date}&store_id={id}
GET http://loyverse:8091/api/v1/receipts/{id}
```

---

## 📤 **Events Published**

### **cash_flow.created**
```json
{
  "event_type": "cash_flow.created",
  "cash_flow_id": "cash_flow_456",
  "location_id": "location_123",
  "location_type": "branch",
  "business_date": "2024-01-15",
  "total_sales": 24000.00,
  "created_at": "2024-01-15T10:30:00Z"
}
```

### **profit_allocation.completed**
```json
{
  "event_type": "profit_allocation.completed",
  "allocation_id": "allocation_123",
  "location_id": "location_123",
  "business_date": "2024-01-15",
  "allocatable_revenue": 15000.00,
  "profit_amount": 750.00,
  "allocated_at": "2024-01-15T18:00:00Z"
}
```

### **expense.approved**
```json
{
  "event_type": "expense.approved",
  "expense_id": "expense_789",
  "cash_flow_id": "cash_flow_456",
  "category": "operational",
  "amount": 1500.00,
  "approved_by": "user_123",
  "approved_at": "2024-01-15T15:00:00Z"
}
```

### **financial.alert**
```json
{
  "event_type": "financial.alert",
  "alert_type": "high_cash_on_hand",
  "severity": "medium",
  "location_id": "location_123",
  "amount": 5000.00,
  "message": "Cash on hand exceeds recommended threshold",
  "action_required": true,
  "timestamp": "2024-01-15T19:00:00Z"
}
```

---

## 📥 **Events Consumed**

### **order.completed**
- **Action**: Record delivery revenue automatically
- **Trigger**: Update delivery orders total in cash flow

### **payment.confirmed**
- **Action**: Update cash flow revenue
- **Trigger**: Refresh financial calculations

### **loyverse.receipt_created**
- **Action**: Process receipt allocation
- **Trigger**: Update location revenue

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
SERVER_PORT=8088
GO_ENV=development

# External Services
ORDER_SERVICE_URL=http://order:8081
CUSTOMER_SERVICE_URL=http://customer:8110
PAYMENT_SERVICE_URL=http://payment:8087
LOYVERSE_INTEGRATION_URL=http://loyverse:8091
NOTIFICATION_SERVICE_URL=http://notification:8092

# Redis Cache
REDIS_URL=redis://redis:6379

# Kafka Events
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC_FINANCE_EVENTS=finance-events

# File Storage
S3_BUCKET=saan-finance-receipts
S3_REGION=ap-southeast-1
S3_ACCESS_KEY=${S3_ACCESS_KEY}
S3_SECRET_KEY=${S3_SECRET_KEY}

# Profit First Defaults
DEFAULT_PROFIT_PERCENTAGE=5.0
DEFAULT_OWNER_PAY_PERCENTAGE=10.0
DEFAULT_TAX_PERCENTAGE=15.0
MINIMUM_ALLOCATION_AMOUNT=1000.0

# Banking Integration
BANK_API_URL=${BANK_API_URL}
BANK_API_KEY=${BANK_API_KEY}
```

---

## 🚨 **Error Codes**

| Code | Message | Description |
|------|---------|-------------|
| 400 | Invalid request | Missing or invalid parameters |
| 404 | Cash flow not found | Cash flow ID doesn't exist |
| 409 | Cash flow already exists | Duplicate cash flow for date/location |
| 422 | Invalid allocation percentages | Profit First percentages don't sum to 100% |
| 423 | Cash flow locked | Cannot modify submitted/approved cash flow |
| 429 | Rate limit exceeded | Too many requests |
| 500 | Service unavailable | Internal server error |
| 503 | External service error | Dependent service unavailable |

---

## 💾 **Caching Strategy**

### **Redis Cache Keys**
```redis
# Real-time metrics (5-15 min TTL)
finance:daily_sales:{location_id}:{date} → Current day sales total
finance:cash_position:{location_id} → Current cash on hand
finance:pending_expenses:{location_id} → Sum of pending expenses
finance:allocation_status:{location_id}:{date} → Profit allocation progress

# Calculation cache (1-4 hours TTL)
finance:profit_calculation:{location_id}:{date} → Daily profit calculations
finance:expense_totals:{location_id}:{month} → Monthly expense summaries
finance:revenue_breakdown:{location_id}:{date} → Revenue source analysis

# Report cache (1-24 hours TTL)
finance:daily_report:{location_id}:{date} → Complete daily reports
finance:monthly_summary:{location_id}:{month} → Monthly summaries
finance:dashboard_data:{user_id} → User-specific dashboard
```

---

## 🧪 **Testing Examples**

### **Create Cash Flow (cURL)**
```bash
curl -X POST http://localhost:8088/api/v1/finance/cash-flows \
  -H "Content-Type: application/json" \
  -d '{
    "business_date": "2024-01-15",
    "location_id": "location_123",
    "location_type": "branch",
    "location_name": "Branch Central",
    "loyverse_receipts_total": 15000.00,
    "delivery_orders_total": 8500.00,
    "cash_on_hand": 2000.00,
    "cash_reason": "Float money"
  }'
```

### **Record Expense (cURL)**
```bash
curl -X POST http://localhost:8088/api/v1/finance/expenses \
  -H "Content-Type: application/json" \
  -d '{
    "cash_flow_id": "cash_flow_456",
    "expense_category": "operational",
    "expense_subcategory": "fuel",
    "description": "Fuel for delivery truck",
    "amount": 1500.00,
    "expense_date": "2024-01-15",
    "vendor_name": "PTT Station",
    "payment_method": "cash"
  }'
```

### **Get Daily Report (cURL)**
```bash
curl "http://localhost:8088/api/v1/finance/reports/daily?location_id=location_123&date=2024-01-15"
```

---

## 📊 **Status Flow**

### **Cash Flow Status**
```
draft → submitted → approved → completed
  ↓         ↓         ↓
rejected  rejected  rejected
```

### **Expense Status**
```
pending → approved → paid
    ↓
 rejected
```

### **Profit Allocation Status**
```
calculated → profit_allocated → owner_pay_allocated → tax_allocated → completed
```

---

## 📈 **Performance Metrics**

### **Key Metrics**
- Daily report generation: <2 seconds
- Expense processing: <500ms
- Profit allocation: <1 second
- Cache hit rate: >85%
- Loyverse sync success: >95%

---

> 💰 **Finance Service provides comprehensive cash flow management, automated profit allocation, and real-time financial monitoring for SAAN operations**
