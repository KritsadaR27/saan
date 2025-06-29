# Order Service Events Documentation

## Overview
The Order Service publishes events to the message bus to notify other services about order state changes and important business events. This enables real-time updates, audit trails, and integration with external systems.

**Event System:** Apache Kafka (Development: Mock Publisher)  
**Message Format:** JSON  
**Delivery:** At-least-once with outbox pattern  

## Event Categories

### Order Lifecycle Events
Events related to order creation, updates, and state transitions.

### Inventory Events
Events related to stock management and inventory updates.

### Customer Events
Events related to customer interactions and notifications.

### System Events
Events related to service health and operational monitoring.

## Event List

### 1. OrderCreated

Published when a new order is successfully created.

**Topic:** `order.created`

**Payload:**
```json
{
  "event_id": "uuid",
  "event_type": "order.created",
  "timestamp": "2025-06-29T10:00:00Z",
  "version": "1.0",
  "source": "order-service",
  "data": {
    "order_id": "uuid",
    "order_number": "ORD-2025-001234",
    "customer_id": "uuid",
    "total_amount": 200.00,
    "currency": "THB",
    "status": "pending",
    "items": [
      {
        "product_id": "uuid",
        "quantity": 2,
        "unit_price": 100.00,
        "total_price": 200.00
      }
    ],
    "shipping_address": {
      "street": "123 Main St",
      "city": "Bangkok",
      "postal_code": "10100",
      "country": "Thailand"
    },
    "created_by": "uuid",
    "source_channel": "web" // web, mobile, chat, admin
  },
  "metadata": {
    "correlation_id": "uuid",
    "causation_id": "uuid",
    "user_id": "uuid"
  }
}
```

**Subscribers:**
- Inventory Service (stock reservation)
- Finance Service (revenue tracking)
- Notification Service (customer notifications)
- Reporting Service (analytics)

### 2. OrderUpdated

Published when order details are modified.

**Topic:** `order.updated`

**Payload:**
```json
{
  "event_id": "uuid",
  "event_type": "order.updated",
  "timestamp": "2025-06-29T10:05:00Z",
  "version": "1.0",
  "source": "order-service",
  "data": {
    "order_id": "uuid",
    "order_number": "ORD-2025-001234",
    "customer_id": "uuid",
    "changes": {
      "items": {
        "added": [...],
        "removed": [...],
        "modified": [...]
      },
      "shipping_address": {...},
      "total_amount": {
        "old": 200.00,
        "new": 250.00
      }
    },
    "updated_by": "uuid",
    "reason": "Customer requested item addition"
  },
  "metadata": {
    "correlation_id": "uuid",
    "causation_id": "uuid",
    "user_id": "uuid"
  }
}
```

**Subscribers:**
- Inventory Service (stock adjustments)
- Finance Service (revenue updates)
- Notification Service (change notifications)

### 3. OrderStatusChanged

Published when order status changes.

**Topic:** `order.status_changed`

**Payload:**
```json
{
  "event_id": "uuid",
  "event_type": "order.status_changed",
  "timestamp": "2025-06-29T10:10:00Z",
  "version": "1.0",
  "source": "order-service",
  "data": {
    "order_id": "uuid",
    "order_number": "ORD-2025-001234",
    "customer_id": "uuid",
    "status_change": {
      "from": "pending",
      "to": "confirmed"
    },
    "reason": "Payment confirmed",
    "changed_by": "uuid",
    "automatic": false
  },
  "metadata": {
    "correlation_id": "uuid",
    "causation_id": "uuid",
    "user_id": "uuid"
  }
}
```

**Subscribers:**
- Shipping Service (shipping preparation)
- Notification Service (status notifications)
- Finance Service (payment processing)
- Reporting Service (status analytics)

### 4. OrderConfirmed

Published when order is confirmed and ready for fulfillment.

**Topic:** `order.confirmed`

**Payload:**
```json
{
  "event_id": "uuid",
  "event_type": "order.confirmed",
  "timestamp": "2025-06-29T10:15:00Z",
  "version": "1.0",
  "source": "order-service",
  "data": {
    "order_id": "uuid",
    "order_number": "ORD-2025-001234",
    "customer_id": "uuid",
    "total_amount": 200.00,
    "confirmed_by": "uuid",
    "payment_confirmed": true,
    "inventory_reserved": true,
    "estimated_fulfillment_date": "2025-06-30T00:00:00Z"
  },
  "metadata": {
    "correlation_id": "uuid",
    "causation_id": "uuid",
    "user_id": "uuid"
  }
}
```

**Subscribers:**
- Shipping Service (create shipment)
- Inventory Service (commit stock reservation)
- Finance Service (revenue recognition)
- Notification Service (confirmation notification)

### 5. OrderCancelled

Published when order is cancelled.

**Topic:** `order.cancelled`

**Payload:**
```json
{
  "event_id": "uuid",
  "event_type": "order.cancelled",
  "timestamp": "2025-06-29T10:20:00Z",
  "version": "1.0",
  "source": "order-service",
  "data": {
    "order_id": "uuid",
    "order_number": "ORD-2025-001234",
    "customer_id": "uuid",
    "cancellation_reason": "Customer request",
    "cancelled_by": "uuid",
    "automatic": false,
    "refund_required": true,
    "items_to_restock": [
      {
        "product_id": "uuid",
        "quantity": 2
      }
    ]
  },
  "metadata": {
    "correlation_id": "uuid",
    "causation_id": "uuid",
    "user_id": "uuid"
  }
}
```

**Subscribers:**
- Inventory Service (release stock reservation)
- Finance Service (process refund)
- Notification Service (cancellation notification)
- Shipping Service (cancel shipment if exists)

### 6. OrderStockOverrideApplied

Published when stock override is applied to an order.

**Topic:** `order.stock_override_applied`

**Payload:**
```json
{
  "event_id": "uuid",
  "event_type": "order.stock_override_applied",
  "timestamp": "2025-06-29T10:25:00Z",
  "version": "1.0",
  "source": "order-service",
  "data": {
    "order_id": "uuid",
    "order_number": "ORD-2025-001234",
    "customer_id": "uuid",
    "override_details": {
      "reason": "VIP customer special request",
      "expected_restock_date": "2025-07-01T00:00:00Z",
      "approved_by": "uuid"
    },
    "items_overridden": [
      {
        "product_id": "uuid",
        "requested_quantity": 5,
        "available_quantity": 2,
        "override_quantity": 3
      }
    ]
  },
  "metadata": {
    "correlation_id": "uuid",
    "causation_id": "uuid",
    "user_id": "uuid"
  }
}
```

**Subscribers:**
- Inventory Service (negative stock tracking)
- Finance Service (special order tracking)
- Notification Service (override notifications)
- Reporting Service (override analytics)

### 7. ChatOrderCreated

Published when order is created through chat interface.

**Topic:** `order.chat_created`

**Payload:**
```json
{
  "event_id": "uuid",
  "event_type": "order.chat_created",
  "timestamp": "2025-06-29T10:30:00Z",
  "version": "1.0",
  "source": "order-service",
  "data": {
    "order_id": "uuid",
    "order_number": "ORD-2025-001234",
    "customer_id": "uuid",
    "chat_id": "uuid",
    "template_used": "quick_order",
    "ai_confidence": 0.95,
    "customer_preferences": {
      "delivery_speed": "standard",
      "communication_method": "sms"
    },
    "requires_confirmation": true
  },
  "metadata": {
    "correlation_id": "uuid",
    "causation_id": "uuid",
    "user_id": "uuid"
  }
}
```

**Subscribers:**
- Chatbot Service (conversation continuation)
- Notification Service (confirmation requests)
- Customer Service (manual review if needed)

### 8. OrderLinkedToChat

Published when existing order is linked to chat conversation.

**Topic:** `order.chat_linked`

**Payload:**
```json
{
  "event_id": "uuid",
  "event_type": "order.chat_linked",
  "timestamp": "2025-06-29T10:35:00Z",
  "version": "1.0",
  "source": "order-service",
  "data": {
    "order_id": "uuid",
    "order_number": "ORD-2025-001234",
    "chat_id": "uuid",
    "link_type": "support", // support, inquiry, complaint
    "linked_by": "uuid",
    "context": "Customer inquiry about delivery status"
  },
  "metadata": {
    "correlation_id": "uuid",
    "causation_id": "uuid",
    "user_id": "uuid"
  }
}
```

**Subscribers:**
- Chatbot Service (context enrichment)
- Customer Service (support ticket creation)

### 9. OrderAuditEvent

Published for all significant order operations for audit trail.

**Topic:** `order.audit`

**Payload:**
```json
{
  "event_id": "uuid",
  "event_type": "order.audit",
  "timestamp": "2025-06-29T10:40:00Z",
  "version": "1.0",
  "source": "order-service",
  "data": {
    "order_id": "uuid",
    "order_number": "ORD-2025-001234",
    "action": "status_update", // create, update, delete, status_update, etc.
    "actor": {
      "user_id": "uuid",
      "user_type": "admin", // admin, system, customer, ai
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0..."
    },
    "changes": {
      "field": "status",
      "old_value": "pending",
      "new_value": "confirmed"
    },
    "reason": "Payment confirmed by payment gateway"
  },
  "metadata": {
    "correlation_id": "uuid",
    "causation_id": "uuid",
    "user_id": "uuid"
  }
}
```

**Subscribers:**
- Audit Service (compliance tracking)
- Reporting Service (audit reports)
- Security Service (fraud detection)

### 10. OrderStatisticsUpdated

Published when order statistics need to be recalculated.

**Topic:** `order.statistics_updated`

**Payload:**
```json
{
  "event_id": "uuid",
  "event_type": "order.statistics_updated",
  "timestamp": "2025-06-29T10:45:00Z",
  "version": "1.0",
  "source": "order-service",
  "data": {
    "trigger_event": "order.confirmed",
    "affected_metrics": [
      "daily_revenue",
      "customer_lifetime_value",
      "product_sales_count"
    ],
    "date_range": {
      "start": "2025-06-29T00:00:00Z",
      "end": "2025-06-29T23:59:59Z"
    },
    "customer_id": "uuid",
    "product_ids": ["uuid1", "uuid2"]
  },
  "metadata": {
    "correlation_id": "uuid",
    "causation_id": "uuid"
  }
}
```

**Subscribers:**
- Reporting Service (analytics update)
- Dashboard Service (real-time metrics)

## Event Schema Versioning

Events use semantic versioning (major.minor) in the `version` field:
- **Major version change**: Breaking changes to event structure
- **Minor version change**: Backward-compatible additions

Current versions:
- All events: `1.0` (initial release)

## Event Metadata

All events include standard metadata:
- `correlation_id`: Links related events across services
- `causation_id`: Points to the triggering event
- `user_id`: User who initiated the action (if applicable)

## Outbox Pattern Implementation

The Order Service uses the Outbox Pattern for reliable event publishing:

1. **Transaction Atomicity**: Events are stored in `order_events_outbox` table within the same database transaction as the business operation
2. **Outbox Worker**: Background worker polls the outbox table and publishes events to Kafka
3. **At-Least-Once Delivery**: Events may be delivered multiple times; consumers should be idempotent
4. **Retry Logic**: Failed events are retried with exponential backoff

### Outbox Configuration
```json
{
  "polling_interval": "5s",
  "batch_size": 10,
  "max_retries": 3,
  "retry_backoff": "exponential"
}
```

## Subscriber Guide

### Event Consumer Implementation

#### 1. Message Format Validation
```go
type OrderEvent struct {
    EventID   string      `json:"event_id"`
    EventType string      `json:"event_type"`
    Timestamp time.Time   `json:"timestamp"`
    Version   string      `json:"version"`
    Source    string      `json:"source"`
    Data      interface{} `json:"data"`
    Metadata  Metadata    `json:"metadata"`
}
```

#### 2. Consumer Example (Go)
```go
func (c *OrderEventConsumer) HandleOrderCreated(ctx context.Context, event OrderEvent) error {
    // Idempotency check
    if c.hasProcessed(event.EventID) {
        return nil // Already processed
    }
    
    // Process the event
    err := c.processOrderCreated(event.Data)
    if err != nil {
        return fmt.Errorf("failed to process order created: %w", err)
    }
    
    // Mark as processed
    return c.markProcessed(event.EventID)
}
```

#### 3. Error Handling
- **Transient Errors**: Retry with exponential backoff
- **Permanent Errors**: Log and send to dead letter queue
- **Schema Errors**: Log detailed error for investigation

#### 4. Consumer Configuration
```yaml
consumer:
  group_id: "inventory-service"
  topics: 
    - "order.created"
    - "order.status_changed"
    - "order.cancelled"
  auto_offset_reset: "earliest"
  enable_auto_commit: false
```

### Testing Events

#### 1. Development Environment
```bash
# Start Kafka (using docker-compose)
docker-compose up -d kafka

# Monitor events
docker exec -it kafka kafka-console-consumer.sh \
  --topic order.created \
  --from-beginning \
  --bootstrap-server localhost:9092
```

#### 2. Event Simulation
```bash
# Trigger order creation via API
curl -X POST http://localhost:8081/api/v1/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{...}'

# Check outbox table
docker exec postgres psql -U saan -d saan_db \
  -c "SELECT * FROM order_events_outbox ORDER BY created_at DESC LIMIT 5;"
```

### Monitoring and Observability

#### 1. Event Metrics
- Events published per topic
- Publishing latency
- Consumer lag
- Failed event count

#### 2. Health Checks
- Outbox worker status
- Kafka connectivity
- Consumer group status

#### 3. Alerting
- High consumer lag (> 1000 messages)
- Publishing failures (> 5% error rate)
- Dead letter queue accumulation

## Message Bus Configuration

### Kafka Topics
| Topic | Partitions | Replication | Retention |
|-------|------------|-------------|-----------|
| order.created | 3 | 2 | 7 days |
| order.updated | 3 | 2 | 7 days |
| order.status_changed | 3 | 2 | 7 days |
| order.confirmed | 3 | 2 | 7 days |
| order.cancelled | 3 | 2 | 7 days |
| order.stock_override_applied | 1 | 2 | 30 days |
| order.chat_created | 2 | 2 | 7 days |
| order.chat_linked | 1 | 2 | 7 days |
| order.audit | 1 | 3 | 90 days |
| order.statistics_updated | 1 | 2 | 1 day |

### Environment Variables
```bash
# Kafka Configuration
KAFKA_BROKERS=kafka:9092
KAFKA_CONSUMER_GROUP=order-service-consumer
KAFKA_PRODUCER_ACKS=all
KAFKA_PRODUCER_RETRIES=3

# Event Publishing
EVENT_OUTBOX_ENABLED=true
EVENT_OUTBOX_POLLING_INTERVAL=5s
EVENT_OUTBOX_BATCH_SIZE=10
```

## Integration Examples

### 1. Inventory Service Integration
```go
// Subscribe to order events for inventory management
func (s *InventoryService) HandleOrderCreated(event OrderEvent) error {
    orderData := event.Data.(OrderCreatedData)
    
    // Reserve inventory for each item
    for _, item := range orderData.Items {
        err := s.reserveStock(item.ProductID, item.Quantity)
        if err != nil {
            // Publish inventory shortage event
            s.publishInventoryShortage(orderData.OrderID, item)
            return err
        }
    }
    
    return nil
}
```

### 2. Notification Service Integration
```go
// Subscribe to order events for customer notifications
func (s *NotificationService) HandleOrderStatusChanged(event OrderEvent) error {
    statusData := event.Data.(OrderStatusChangedData)
    
    // Send appropriate notification based on status
    switch statusData.StatusChange.To {
    case "confirmed":
        return s.sendOrderConfirmation(statusData.CustomerID, statusData.OrderID)
    case "shipped":
        return s.sendShippingNotification(statusData.CustomerID, statusData.OrderID)
    case "delivered":
        return s.sendDeliveryNotification(statusData.CustomerID, statusData.OrderID)
    }
    
    return nil
}
```

### 3. Finance Service Integration
```go
// Subscribe to order events for financial tracking
func (s *FinanceService) HandleOrderConfirmed(event OrderEvent) error {
    orderData := event.Data.(OrderConfirmedData)
    
    // Record revenue
    revenue := &Revenue{
        OrderID:    orderData.OrderID,
        CustomerID: orderData.CustomerID,
        Amount:     orderData.TotalAmount,
        Currency:   "THB",
        Date:       time.Now(),
    }
    
    return s.recordRevenue(revenue)
}
```

## Development Guidelines

### 1. Event Design Principles
- **Immutable**: Events represent facts that happened
- **Self-contained**: Include all necessary data
- **Backward compatible**: Don't break existing consumers
- **Meaningful**: Clear event names and data structure

### 2. Versioning Strategy
- Use semantic versioning for event schemas
- Support multiple versions during transition periods
- Deprecate old versions gradually

### 3. Testing Strategy
- Unit tests for event creation
- Integration tests for event publishing
- Contract tests with event consumers
- End-to-end tests for event flows

### 4. Performance Considerations
- Batch event publishing when possible
- Use async processing for non-critical events
- Monitor consumer lag and scale accordingly
- Implement circuit breakers for external dependencies

Following PROJECT_RULES.md:
- Use service names for internal communication
- Test with docker-compose
- Monitor with `docker-compose logs -f order`
