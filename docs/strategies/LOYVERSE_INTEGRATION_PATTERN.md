# 🏪 Loyverse Multi-Store Integration Pattern

## 🎯 Overview

Integration pattern for Loyverse POS system with multi-store management, delivery context tracking, and automated store selection for SAAN payment service.

## 🏗️ Architecture Pattern

### **Store Selection Strategy**
```
Payment Request → Store Selection Engine → Receipt Creation
     ↓                       ↓                    ↓
Context Analysis    →  Selection Logic  →  Loyverse API
     ↓                       ↓                    ↓
Driver/Route Info   →  Store Assignment →  Receipt Note
```

### **Decision Matrix**
| Payment Channel | Payment Method | Store Selection Logic | Note Context |
|----------------|----------------|----------------------|---------------|
| POS | Cash/Transfer | Main Store | POS Transaction |
| Delivery | COD Cash | Driver's Store | Driver + Route Info |
| SAAN App | Prepaid | Main Store | App Order |
| SAAN Chat | COD | Driver's Store | Chat + Driver Info |

## 🔄 Integration Flow

### **1. Automatic Store Selection**
```go
func SelectStore(paymentContext PaymentContext) (*Store, error) {
    // Priority 1: Manual override
    if paymentContext.ForceStoreID != nil {
        return GetStoreByID(*paymentContext.ForceStoreID)
    }
    
    // Priority 2: Driver assignment (COD)
    if paymentContext.DeliveryDriverPhone != nil {
        return GetStoreByDriverPhone(*paymentContext.DeliveryDriverPhone)
    }
    
    // Priority 3: Payment channel logic
    switch paymentContext.Channel {
    case "delivery":
        return GetFirstDeliveryStore()
    case "pos":
        return GetMainStore()
    default:
        return GetDefaultStore()
    }
}
```

### **2. Receipt Note Generation**
```go
func GenerateReceiptNote(context PaymentDeliveryContext) string {
    var parts []string
    
    // Driver information
    if context.DriverName != nil {
        parts = append(parts, fmt.Sprintf("🚚 ส่งโดย: %s", *context.DriverName))
    }
    
    // Contact information
    if context.DriverPhone != nil {
        parts = append(parts, fmt.Sprintf("📞 %s", *context.DriverPhone))
    }
    
    // Route information
    if context.Route != nil {
        parts = append(parts, fmt.Sprintf("📍 เส้น: %s", *context.Route))
    }
    
    // Delivery app
    if context.DeliveryApp != nil {
        parts = append(parts, fmt.Sprintf("📱 App: %s", *context.DeliveryApp))
    }
    
    return strings.Join(parts, " | ")
}
```

## 📊 Store Configuration

### **Store Types & Capabilities**
```sql
-- Main Store (Physical location)
INSERT INTO loyverse_stores VALUES (
    'store_main_001', 
    'ลุงรวยหน้าร้าน', 
    'main', 
    true,  -- accepts_cash
    true,  -- accepts_transfer  
    false, -- accepts_cod
    NULL,  -- no driver
    NULL   -- no route
);

-- Delivery Stores (Virtual stores for each route)
INSERT INTO loyverse_stores VALUES (
    'store_delivery_001',
    'ลุงรวยรถส่งของ - เส้น A',
    'delivery',
    true,   -- accepts_cash (COD)
    true,   -- accepts_transfer (COD)
    true,   -- accepts_cod
    '+66812345001', -- driver phone
    'route_a'       -- delivery route
);
```

### **Store Selection Rules**
```yaml
rules:
  pos_payment:
    payment_channel: "loyverse_pos"
    target_store: "main"
    note_template: "POS Transaction"
    
  online_prepaid:
    payment_channel: ["saan_app", "saan_chat"]
    payment_timing: "prepaid"
    target_store: "main"
    note_template: "📱 Online Order"
    
  cod_delivery:
    payment_method: ["cod_cash", "cod_transfer"]
    selection_logic: "driver_assignment"
    fallback_store: "delivery_default"
    note_template: "🚚 COD Delivery | Driver: {driver_name} | Route: {route}"
    
  grab_delivery:
    delivery_app: "grab"
    target_store: "main"
    note_template: "🛵 Grab Delivery | Order: {order_id}"
```

## 🔧 Implementation Patterns

### **Repository Pattern**
```go
type LoyverseStoreRepository interface {
    GetActiveStores(ctx context.Context) ([]*LoyverseStore, error)
    GetStoreByID(ctx context.Context, storeID string) (*LoyverseStore, error)
    GetStoreByDriverPhone(ctx context.Context, phone string) (*LoyverseStore, error)
    GetStoresByType(ctx context.Context, storeType StoreType) ([]*LoyverseStore, error)
    GetDefaultStore(ctx context.Context) (*LoyverseStore, error)
}

type PaymentDeliveryContextRepository interface {
    Create(ctx context.Context, context *PaymentDeliveryContext) error
    GetByPaymentID(ctx context.Context, paymentID uuid.UUID) (*PaymentDeliveryContext, error)
    UpdateDeliveryInfo(ctx context.Context, paymentID uuid.UUID, updates map[string]interface{}) error
}
```

### **Use Case Pattern**
```go
type StoreSelectionUseCase struct {
    storeRepo         LoyverseStoreRepository
    deliveryContextRepo PaymentDeliveryContextRepository
    shippingClient    ShippingServiceClient
}

func (uc *StoreSelectionUseCase) SelectStoreForPayment(
    ctx context.Context, 
    request StoreSelectionRequest,
) (*StoreSelectionResult, error) {
    // 1. Validate input
    // 2. Apply selection logic
    // 3. Create delivery context if needed
    // 4. Generate receipt note
    // 5. Return result with audit trail
}
```

### **Factory Pattern for Receipt Creation**
```go
type LoyverseReceiptFactory struct {
    client          *LoyverseClient
    storeSelector   *StoreSelectionUseCase
    noteGenerator   *ReceiptNoteGenerator
}

func (f *LoyverseReceiptFactory) CreateReceipt(
    ctx context.Context,
    payment *PaymentTransaction,
    order *Order,
    customer *Customer,
) (*LoyverseReceipt, error) {
    // 1. Select appropriate store
    storeResult, err := f.storeSelector.SelectStoreForPayment(ctx, StoreSelectionRequest{
        PaymentID:      payment.ID,
        PaymentMethod:  payment.PaymentMethod,
        PaymentChannel: payment.PaymentChannel,
    })
    
    // 2. Generate receipt note
    note := f.noteGenerator.Generate(storeResult.DeliveryContext, storeResult.SelectedStore)
    
    // 3. Create Loyverse receipt
    return f.client.CreateReceipt(ctx, LoyverseReceiptRequest{
        StoreID:     storeResult.SelectedStore.StoreID,
        CustomerID:  customer.LoyverseID,
        LineItems:   f.buildLineItems(order.Items),
        TotalAmount: payment.Amount,
        Note:        note,
        PaymentType: payment.GetLoyversePaymentType(),
    })
}
```

## 📈 Analytics & Reporting

### **Store Performance Metrics**
```sql
-- Revenue per store
SELECT 
    ls.store_name,
    ls.store_type,
    DATE(pt.created_at) as date,
    COUNT(*) as transaction_count,
    SUM(pt.amount) as total_revenue,
    AVG(pt.amount) as avg_transaction
FROM payment_transactions pt
JOIN loyverse_stores ls ON pt.assigned_store_id = ls.store_id
WHERE pt.status = 'completed'
  AND pt.created_at >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY ls.store_id, ls.store_name, ls.store_type, DATE(pt.created_at)
ORDER BY date DESC, total_revenue DESC;
```

### **Driver Performance Tracking**
```sql
-- Driver commission calculation
SELECT 
    pdc.delivery_driver_name,
    pdc.delivery_driver_phone,
    pdc.delivery_route,
    COUNT(*) as deliveries_count,
    SUM(pt.amount) as total_sales,
    SUM(pt.amount) * 0.05 as commission_5_percent
FROM payment_delivery_context pdc
JOIN payment_transactions pt ON pdc.payment_transaction_id = pt.id
WHERE pt.status = 'completed'
  AND pt.created_at >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY pdc.delivery_driver_name, pdc.delivery_driver_phone, pdc.delivery_route
ORDER BY total_sales DESC;
```

### **Store Comparison Dashboard**
```sql
-- Store efficiency comparison
SELECT 
    ls.store_name,
    COUNT(pt.id) as total_transactions,
    SUM(CASE WHEN pt.status = 'completed' THEN 1 ELSE 0 END) as successful_transactions,
    ROUND(
        SUM(CASE WHEN pt.status = 'completed' THEN 1 ELSE 0 END) * 100.0 / COUNT(pt.id), 
        2
    ) as success_rate,
    SUM(CASE WHEN pt.status = 'completed' THEN pt.amount ELSE 0 END) as total_revenue
FROM loyverse_stores ls
LEFT JOIN payment_transactions pt ON ls.store_id = pt.assigned_store_id
WHERE ls.is_active = true
  AND (pt.created_at IS NULL OR pt.created_at >= CURRENT_DATE - INTERVAL '30 days')
GROUP BY ls.store_id, ls.store_name
ORDER BY success_rate DESC, total_revenue DESC;
```

## 🔍 Monitoring & Alerting

### **Key Metrics to Monitor**
- **Store Selection Accuracy**: % of automatic selections vs manual overrides
- **Receipt Creation Success Rate**: % of successful Loyverse API calls
- **Driver Assignment Accuracy**: % of correct driver-to-store assignments
- **Payment-to-Receipt Lag**: Time between payment and receipt creation

### **Alert Conditions**
```yaml
alerts:
  loyverse_api_failure:
    condition: "receipt_creation_failure_rate > 5%"
    severity: "high"
    notification: "slack + email"
    
  store_selection_failure:
    condition: "auto_selection_failure_rate > 10%"
    severity: "medium"
    notification: "slack"
    
  driver_assignment_mismatch:
    condition: "driver_store_mismatch_rate > 15%"
    severity: "low"
    notification: "daily_report"
    
  receipt_creation_lag:
    condition: "avg_receipt_creation_time > 30_seconds"
    severity: "medium"
    notification: "slack"
```

## 🚀 Best Practices

### **Configuration Management**
- Store configurations should be manageable via admin API
- Support for A/B testing different store selection strategies
- Environment-specific store mappings (dev/staging/prod)

### **Error Handling**
- Graceful fallback to default store when selection fails
- Retry mechanism for Loyverse API failures
- Comprehensive error logging with context

### **Security Considerations**
- Encrypt sensitive driver information
- Rate limiting for Loyverse API calls
- Audit trail for all store selection decisions

### **Scalability Patterns**
- Cache store configurations in Redis
- Async receipt creation for better performance
- Database partitioning for large transaction volumes

---

> 🏪 **Robust Loyverse Multi-Store Integration Pattern for scalable payment processing with clear financial separation!**
