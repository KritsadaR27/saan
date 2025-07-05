# 📸 SAAN Snapshot Strategy Guide

## 🎯 Overview

Snapshot strategy สำหรับ SAAN system เพื่อ audit trail, compliance, historical analysis และ business intelligence

## � Table of Contents
1. [Snapshot Categories](#-snapshot-categories)
2. [Database Schema & Implementation](#️-database-schema--implementation)
3. [Usage Examples & Analytics](#-usage-examples--analytics)
4. [Performance & Best Practices](#-performance--best-practices)

## �📊 Snapshot Categories

### ✅ **ควรทำ Snapshot**

#### 🛒 **Order Snapshots**
**Purpose:** Business compliance, audit trail, rollback capability

**Trigger Points:**
```
├── order.created → Initial snapshot (draft state)
├── order.confirmed → Payment confirmed snapshot  
├── order.shipped → Delivery assigned snapshot
├── order.completed → Final snapshot with receipt
├── order.cancelled → Cancellation snapshot
└── order.modified → Modification snapshot
```

**Data Structure:**
```json
{
  "snapshot_id": "snap_123",
  "order_id": "order_456", 
  "snapshot_type": "confirmed",
  "timestamp": "2025-07-03T14:30:00Z",
  "triggered_by": {
    "user_id": "staff_789",
    "action": "payment_confirmed",
    "source": "payment_service"
  },
  "order_state": {
    "status": "confirmed",
    "customer": {...},
    "items": [...],
    "pricing": {...},
    "delivery": {...}
  },
  "changes": {
    "previous_status": "pending",
    "new_status": "confirmed"
  }
}
```

---

#### 📦 **Inventory Transaction Snapshots**
**Purpose:** Cost calculation, audit trail, reconciliation

**Trigger Points:**
```
├── inventory.deducted → Order completion, sales
├── inventory.restocked → Supplier delivery, returns
├── inventory.adjusted → Manual adjustments, corrections
├── inventory.transferred → Inter-store transfers
└── inventory.damaged → Damage reports, write-offs
```

**Data Structure:**
```json
{
  "transaction_id": "inv_tx_789",
  "product_id": "prod_123",
  "transaction_type": "deducted",
  "quantity": -5,
  "reason": "order_456_completed",
  "reference_id": "order_456",
  "product_state": {
    "name": "โค้ก 325ml",
    "cost_price": 15.00,
    "selling_price": 20.00
  },
  "transaction_context": {
    "warehouse_location": "A-1-5",
    "batch_number": "BATCH001",
    "expiry_date": "2025-12-31"
  },
  "created_at": "2025-07-03T14:30:00Z"
}
```

---

#### 💬 **Chat Conversation Snapshots**
**Purpose:** Customer service quality, AI training, dispute resolution

**Trigger Points:**
```
├── chat.order_created → Customer completed order via chat
├── chat.escalated → Conversation escalated to human agent
├── chat.complaint → Customer complaint logged
├── chat.session_ended → End of customer conversation
└── chat.ai_learning → Significant AI interaction for training
```

**Data Structure:**
```json
{
  "snapshot_id": "chat_snap_456",
  "session_id": "chat_sess_123",
  "customer_id": "cust_789",
  "snapshot_type": "order_created",
  "timestamp": "2025-07-03T15:45:00Z",
  "conversation_context": {
    "platform": "line",
    "total_messages": 12,
    "conversation_duration": 420,
    "intent_progression": ["greeting", "product_inquiry", "order_creation", "payment_confirmation"],
    "ai_confidence_scores": [0.95, 0.87, 0.92, 0.98],
    "human_handoff": false
  },
  "business_outcome": {
    "order_created": true,
    "order_id": "order_456",
    "order_value": 285.50,
    "conversion_rate": 1.0,
    "customer_satisfaction": 5
  },
  "key_messages": [
    {
      "timestamp": "2025-07-03T15:30:00Z",
      "sender": "customer", 
      "content": "อยากสั่งโค้กกับมาม่า",
      "intent": "product_inquiry"
    },
    {
      "timestamp": "2025-07-03T15:44:00Z",
      "sender": "bot",
      "content": "ยืนยันการสั่งซื้อเรียบร้อยแล้วค่ะ รวม 285.50 บาท",
      "intent": "order_confirmation"
    }
  ]
}
```

---

#### 📅 **Daily Inventory Snapshots**
**Purpose:** Historical analysis, monthly reports, trend tracking

**Schedule:** Daily at 23:59
```json
{
  "snapshot_date": "2025-07-03",
  "product_id": "prod_123",
  "opening_stock": 100,
  "closing_stock": 95,
  "total_inbound": 20,
  "total_outbound": 25,
  "adjustments": 0,
  "average_cost": 15.50,
  "total_value": 1472.50,
  "snapshot_created_at": "2025-07-03T23:59:00Z"
}
```

---

#### 💰 **Financial Snapshots**
**Purpose:** Accounting compliance, profit calculation

**Trigger Points:**
```
├── daily_revenue → End of day financial summary
├── payment_confirmed → Payment transaction record
├── refund_processed → Refund transaction record
├── expense_recorded → Expense transaction record
└── cash_reconciliation → Daily cash reconciliation
```

#### 👤 **Customer Tier Snapshots**
**Purpose:** Customer lifecycle tracking, loyalty analysis

**Trigger Points:**
```
├── tier_upgraded → VIP tier change
├── points_redeemed → Points transaction
├── milestone_achieved → Customer milestone
└── annual_summary → Yearly customer summary
```

---

### ❌ **ไม่ควรทำ Snapshot**

| Category | Operations | เหตุผล |
|----------|------------|--------|
| **🛒 Cart** | `cart.item_added`, `cart.item_removed`, `cart.quantity_updated`, `cart.cleared` | Temporary data, ไม่มี business value หลัง checkout |
| **📊 Stock Levels** | `current_stock_changed`, `stock_availability_checked`, `price_calculated` | เปลี่ยนบ่อยมาก, ใช้ transactions คำนวณได้ |
| **🔍 User Activity** | `product_viewed`, `search_performed`, `category_browsed` | Analytics data, ไม่จำเป็นต้อง audit trail |
| **📱 Chat Messages** | `chat_message_sent`, `chat_typing_indicator`, `chat_read_receipt` | Too granular, snapshot significant conversations only |

---

## 🏗️ Database Schema & Implementation

### **Core Snapshot Tables**

### **Order Snapshots**
```sql
CREATE TABLE order_snapshots (
    id UUID PRIMARY KEY,
    order_id UUID REFERENCES orders(id),
    snapshot_type VARCHAR(50),
    snapshot_data JSONB,
    previous_snapshot_id UUID REFERENCES order_snapshots(id),
    created_by_user_id UUID,
    created_at TIMESTAMP DEFAULT NOW(),
    
    -- Quick access fields
    order_status VARCHAR(50),
    total_amount DECIMAL(12,2),
    item_count INT,
    customer_id UUID,
    
    INDEX idx_order_snapshots (order_id, created_at),
    INDEX idx_snapshot_type (snapshot_type, created_at)
);
```

### **Chat Conversation Snapshots**
```sql
CREATE TABLE chat_conversation_snapshots (
    id UUID PRIMARY KEY,
    session_id UUID REFERENCES chat_sessions(id),
    customer_id UUID REFERENCES customers(id),
    snapshot_type VARCHAR(50), -- 'order_created', 'escalated', 'session_ended'
    
    -- Conversation Context
    platform VARCHAR(20), -- 'line', 'facebook', 'whatsapp'
    total_messages INT,
    conversation_duration INT, -- seconds
    intent_progression JSONB,
    ai_confidence_scores JSONB,
    human_handoff BOOLEAN DEFAULT false,
    
    -- Business Outcome
    business_outcome JSONB,
    conversion_metrics JSONB,
    customer_satisfaction INT, -- 1-5 rating
    
    -- Key Messages
    key_messages JSONB,
    
    created_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_chat_snapshots_session (session_id, created_at),
    INDEX idx_chat_snapshots_customer (customer_id, created_at),
    INDEX idx_chat_snapshots_type (snapshot_type)
);
```

### **Inventory Transaction Snapshots**
```sql
CREATE TABLE inventory_transactions (
    id UUID PRIMARY KEY,
    product_id UUID REFERENCES products(id),
    transaction_type VARCHAR(50),
    quantity INT,
    reason VARCHAR(200),
    reference_id UUID,
    
    -- Snapshot data
    product_state JSONB,
    transaction_context JSONB,
    
    created_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_inventory_product_time (product_id, created_at),
    INDEX idx_inventory_type (transaction_type)
);
```

### **Daily Inventory Snapshots**
```sql
CREATE TABLE daily_inventory_snapshots (
    id UUID PRIMARY KEY,
    snapshot_date DATE NOT NULL,
    product_id UUID REFERENCES products(id),
    
    opening_stock INT,
    closing_stock INT,
    total_inbound INT DEFAULT 0,
    total_outbound INT DEFAULT 0,
    adjustments INT DEFAULT 0,
    average_cost DECIMAL(10,2),
    total_value DECIMAL(12,2),
    
    snapshot_created_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE (snapshot_date, product_id),
    INDEX idx_daily_snapshots_date (snapshot_date)
);
```

---

### **Implementation Patterns**

#### **Event-Driven Snapshots**
```go
// Order snapshot on state change
func (s *OrderService) HandlePaymentConfirmed(event PaymentConfirmedEvent) error {
    // Update order
    err := s.updateOrderStatus(event.OrderID, "confirmed")
    if err != nil {
        return err
    }
    
    // Create snapshot
    return s.snapshotService.CreateOrderSnapshot(event.OrderID, "confirmed", event)
}
```

#### **Chat Service Integration**
```go
// Chat conversation snapshot on significant events
func (s *ChatService) HandleOrderCompleted(sessionID string, orderID string) error {
    session := s.getSession(sessionID)
    
    // Create conversation snapshot
    snapshot := &ChatConversationSnapshot{
        SessionID:     sessionID,
        CustomerID:    session.CustomerID,
        SnapshotType:  "order_created",
        Platform:      session.Platform,
        BusinessOutcome: map[string]interface{}{
            "order_created": true,
            "order_id":      orderID,
            "order_value":   session.OrderValue,
            "conversion_rate": 1.0,
        },
        ConversationContext: s.buildConversationContext(session),
        KeyMessages:         s.extractKeyMessages(session),
    }
    
    return s.snapshotService.CreateChatSnapshot(snapshot)
}

func (s *ChatService) buildConversationContext(session *ChatSession) map[string]interface{} {
    return map[string]interface{}{
        "total_messages":        session.MessageCount,
        "conversation_duration": session.Duration,
        "intent_progression":    session.IntentHistory,
        "ai_confidence_scores":  session.ConfidenceScores,
        "human_handoff":         session.HumanHandoff,
    }
}

func (s *ChatService) shouldSnapshotSession(sessionID string) bool {
    session := s.getSession(sessionID)
    
    // Snapshot if: Order created, Complaint, Human handoff, Low AI confidence, Long duration
    return session.OrderCreated || session.HasComplaint || session.HumanHandoff ||
           session.AvgConfidence < 0.7 || session.Duration > 600 // 10 minutes
}
```

#### **Scheduled Snapshots**
```go
// Daily inventory snapshots
func (s *InventorySnapshotService) SetupDailySnapshots() {
    cron.AddFunc("59 23 * * *", func() {
        yesterday := time.Now().AddDate(0, 0, -1)
        err := s.GenerateDailySnapshots(yesterday)
        if err != nil {
            log.Error("Daily snapshot failed", "date", yesterday)
        }
    })
}
```

#### **Transaction-Based Snapshots**
```go
// Inventory transaction snapshot
func (s *InventoryService) DeductStock(productID string, quantity int, orderID string) error {
    // Create transaction (this IS the snapshot)
    transaction := &InventoryTransaction{
        ProductID:       productID,
        TransactionType: "deducted",
        Quantity:        -quantity,
        Reason:          "order_completed",
        ReferenceID:     orderID,
        ProductState:    s.getProductSnapshot(productID),
    }
    
    return s.recordTransaction(transaction)
}
```

---

## 📊 Usage Examples & Analytics

### **Order History Tracking**
```sql
-- ดู order timeline
SELECT 
    snapshot_type,
    created_at,
    snapshot_data->>'status' as status,
    snapshot_data->>'total_amount' as amount
FROM order_snapshots 
WHERE order_id = 'order_123'
ORDER BY created_at;
```

### **Chat Performance Analysis**
```sql
-- วิเคราะห์ประสิทธิภาพ Chat Service
SELECT 
    DATE(created_at) as date,
    COUNT(*) as total_conversations,
    COUNT(CASE WHEN business_outcome->>'order_created' = 'true' THEN 1 END) as converted_conversations,
    AVG((business_outcome->>'order_value')::DECIMAL) as avg_order_value,
    AVG(conversation_duration) as avg_duration,
    AVG(customer_satisfaction) as avg_satisfaction
FROM chat_conversation_snapshots 
WHERE snapshot_type = 'session_ended'
  AND created_at >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

### **AI Training Data Extraction**
```sql
-- ดึงข้อมูลสำหรับ train AI model
SELECT 
    key_messages,
    intent_progression,
    ai_confidence_scores,
    business_outcome
FROM chat_conversation_snapshots 
WHERE snapshot_type IN ('order_created', 'escalated')
  AND ai_confidence_scores IS NOT NULL
  AND customer_satisfaction >= 4;
```

### **Monthly Inventory Report**
```sql
-- รายงานสต็อกสิ้นเดือน
SELECT 
    p.name,
    mis.closing_stock,
    mis.total_value,
    mis.cost_of_goods_sold
FROM monthly_inventory_snapshots mis
JOIN products p ON mis.product_id = p.id
WHERE mis.snapshot_month = '2025-07-01'
ORDER BY mis.total_value DESC;
```

---

## ⚡ Performance & Best Practices

### **Performance Optimization**

| Area | Strategy | Implementation |
|------|----------|----------------|
| **Storage** | JSONB compression, Archive old data | Cold storage for 2+ year data |
| **Indexing** | Index frequently queried fields only | customer_id, session_id, timestamp |
| **Queries** | Denormalize, Materialized views | Cache analytics in Redis |
| **Partitioning** | Partition by month for large datasets | Chat snapshots by month |

### **Cleanup Strategy**
```sql
-- Archive old snapshots
INSERT INTO order_snapshots_archive 
SELECT * FROM order_snapshots 
WHERE created_at < NOW() - INTERVAL '2 years';

DELETE FROM order_snapshots 
WHERE created_at < NOW() - INTERVAL '2 years';

-- Cleanup chat snapshots (keep significant ones longer)
DELETE FROM chat_conversation_snapshots 
WHERE created_at < NOW() - INTERVAL '1 year'
  AND snapshot_type = 'session_ended'
  AND business_outcome->>'order_created' = 'false';
```

### **Benefits Summary**

| Snapshot Type | Business Value | Technical Value |
|---------------|----------------|-----------------|
| **Order** | Audit compliance, dispute resolution | Rollback capability, debugging |
| **Inventory Transactions** | Cost calculation, reconciliation | Data integrity, audit trail |
| **Chat Conversations** | Customer service quality, AI training | Conversation analytics, dispute resolution |
| **Daily Inventory** | Trend analysis, planning | Historical reporting |
| **Financial** | Accounting compliance, tax preparation | Audit trail, reconciliation |

### **Best Practices**

#### **Do's:**
✅ Snapshot critical business state changes  
✅ Include sufficient context in snapshot data  
✅ Use consistent snapshot data structure  
✅ Set up automated cleanup for old snapshots  
✅ Monitor snapshot generation failures  
✅ **Snapshot significant chat conversations** (orders, escalations, complaints)  
✅ **Extract key messages for business insights**  

#### **Don'ts:**
❌ Snapshot temporary or volatile data  
❌ Include sensitive data without encryption  
❌ Create snapshots for every minor change  
❌ Forget to handle snapshot generation failures  
❌ Skip validation of snapshot data integrity  
❌ **Snapshot every individual chat message**  
❌ **Store complete conversation logs in snapshots**  

#### **Chat-Specific Guidelines:**
- **Snapshot conversations that lead to business outcomes** (orders, complaints, escalations)
- **Store conversation context and key messages**, not complete transcripts
- **Use snapshots for AI training data** and customer service quality analysis
- **Include customer satisfaction scores** when available
- **Track conversion metrics** from chat to order

---

> 📸 **Complete snapshot strategy for business compliance, audit trail, and AI-powered customer service analytics!**