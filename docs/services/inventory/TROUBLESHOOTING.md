# Inventory Service Troubleshooting Guide

## Common Issues and Solutions

### 1. Stock Level Discrepancies

#### Symptoms
- Stock levels in system don't match physical count
- Negative stock levels appearing
- Stock movements not reflecting in totals

#### Diagnosis
```bash
# Check recent stock movements for a product
curl -H "Authorization: Bearer $TOKEN" \
  "http://inventory-service:8083/api/inventory/movements?product_id=PRODUCT_ID&limit=50"

# Verify current stock calculation
curl -H "Authorization: Bearer $TOKEN" \
  "http://inventory-service:8083/api/inventory/products/PRODUCT_ID/stock"

# Check for concurrent movement records
SELECT product_id, store_id, COUNT(*), 
       SUM(quantity) as total_movement,
       MAX(quantity_after) as final_quantity
FROM stock_movements 
WHERE product_id = 'PRODUCT_ID' 
  AND store_id = 'STORE_ID'
  AND created_at > NOW() - INTERVAL '1 day'
GROUP BY product_id, store_id;
```

#### Solutions
1. **Recalculate Stock Levels**
```sql
-- Recalculate stock from movements
UPDATE stock_levels sl
SET quantity_on_hand = (
    SELECT COALESCE(SUM(sm.quantity), 0)
    FROM stock_movements sm
    WHERE sm.product_id = sl.product_id 
      AND sm.store_id = sl.store_id
),
last_updated = CURRENT_TIMESTAMP
WHERE sl.product_id = 'PRODUCT_ID';
```

2. **Fix Negative Stock**
```sql
-- Identify negative stock
SELECT * FROM stock_levels 
WHERE quantity_on_hand < 0;

-- Create adjustment movement
INSERT INTO stock_movements (
    product_id, store_id, movement_type, 
    quantity, quantity_before, quantity_after,
    reference, notes
) VALUES (
    'PRODUCT_ID', 'STORE_ID', 'ADJUSTMENT',
    ABS(negative_amount), negative_amount, 0,
    'SYSTEM_CORRECTION', 'Negative stock correction'
);
```

### 2. Performance Issues

#### Symptoms
- Slow API responses
- Database timeouts
- High memory usage

#### Diagnosis
```bash
# Check API response times
curl -w "@curl-format.txt" -H "Authorization: Bearer $TOKEN" \
  "http://inventory-service:8083/api/inventory/products"

# Monitor database connections
SELECT pid, usename, application_name, state, 
       query_start, state_change
FROM pg_stat_activity 
WHERE datname = 'inventory_db';

# Check slow queries
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements 
WHERE query LIKE '%stock_levels%' 
  OR query LIKE '%stock_movements%'
ORDER BY total_time DESC;
```

#### Solutions
1. **Database Optimization**
```sql
-- Add missing indexes
CREATE INDEX CONCURRENTLY idx_movements_product_store_date 
ON stock_movements (product_id, store_id, created_at);

CREATE INDEX CONCURRENTLY idx_stock_levels_low_stock 
ON stock_levels (is_low_stock) WHERE is_low_stock = true;

-- Update table statistics
ANALYZE stock_levels;
ANALYZE stock_movements;
```

2. **Cache Configuration**
```go
// Increase cache timeout for stock levels
const StockLevelCacheTimeout = 10 * time.Minute

// Implement batch cache loading
func (s *InventoryService) preloadStockLevels(storeID string) error {
    stocks, err := s.repo.GetAllStockByStore(storeID)
    if err != nil {
        return err
    }
    
    for _, stock := range stocks {
        key := fmt.Sprintf("stock:%s:%s", stock.ProductID, stock.StoreID)
        s.cache.Set(key, stock, StockLevelCacheTimeout)
    }
    
    return nil
}
```

### 3. Integration Failures

#### Symptoms
- Stock updates not syncing to Loyverse
- Order service reporting stock unavailable when stock exists
- Finance service receiving incorrect inventory values

#### Diagnosis
```bash
# Check integration endpoint status
curl -f http://loyverse-integration:8090/health
curl -f http://order-service:8082/health
curl -f http://finance-service:8085/health

# Verify webhook deliveries
curl -H "Authorization: Bearer $TOKEN" \
  "http://inventory-service:8083/api/inventory/webhooks/status"

# Check event publishing
SELECT event_type, COUNT(*), 
       MIN(created_at) as oldest,
       MAX(created_at) as newest
FROM event_log 
WHERE event_type LIKE 'inventory.%'
  AND created_at > NOW() - INTERVAL '1 hour'
GROUP BY event_type;
```

#### Solutions
1. **Retry Failed Webhooks**
```bash
# Manually retry webhook delivery
curl -X POST -H "Authorization: Bearer $TOKEN" \
  "http://inventory-service:8083/api/inventory/webhooks/retry" \
  -d '{"webhook_id": "WEBHOOK_ID"}'
```

2. **Resync Stock Data**
```bash
# Force full stock sync with Loyverse
curl -X POST -H "Authorization: Bearer $TOKEN" \
  "http://inventory-service:8083/api/inventory/sync/loyverse" \
  -d '{"store_id": "STORE_ID", "force": true}'
```

3. **Fix Event Publishing**
```go
// Republish failed events
func (s *InventoryService) republishFailedEvents() error {
    failedEvents, err := s.eventRepo.GetFailedEvents("inventory.%", time.Hour)
    if err != nil {
        return err
    }
    
    for _, event := range failedEvents {
        if err := s.eventPublisher.Publish(event); err != nil {
            log.Error("Failed to republish event", "event_id", event.ID, "error", err)
            continue
        }
        s.eventRepo.MarkAsProcessed(event.ID)
    }
    
    return nil
}
```

### 4. Data Consistency Issues

#### Symptoms
- Reserved stock not released after order expiration
- Stock movements recorded but totals not updated
- Duplicate movement records

#### Diagnosis
```sql
-- Check for expired reservations
SELECT * FROM stock_reservations 
WHERE expires_at < NOW() AND status = 'active';

-- Find duplicate movements
SELECT product_id, store_id, reference, COUNT(*)
FROM stock_movements 
WHERE reference IS NOT NULL
GROUP BY product_id, store_id, reference
HAVING COUNT(*) > 1;

-- Check orphaned stock levels
SELECT sl.* FROM stock_levels sl
LEFT JOIN products_inventory pi ON sl.product_id = pi.product_id
WHERE pi.product_id IS NULL;
```

#### Solutions
1. **Clean Up Expired Reservations**
```sql
-- Release expired reservations
UPDATE stock_reservations 
SET status = 'expired', 
    updated_at = CURRENT_TIMESTAMP
WHERE expires_at < NOW() 
  AND status = 'active';

-- Clean up old expired reservations
DELETE FROM stock_reservations 
WHERE status = 'expired' 
  AND updated_at < NOW() - INTERVAL '7 days';
```

2. **Fix Duplicate Movements**
```sql
-- Remove duplicate movements (keep the latest)
DELETE FROM stock_movements 
WHERE id NOT IN (
    SELECT MAX(id) 
    FROM stock_movements 
    GROUP BY product_id, store_id, reference
);
```

3. **Automated Cleanup Job**
```go
func (s *InventoryService) runCleanupJob() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            s.cleanupExpiredReservations()
            s.validateStockConsistency()
            s.republishFailedEvents()
        }
    }
}
```

### 5. Memory and Resource Issues

#### Symptoms
- High memory usage
- CPU spikes during stock calculations
- Database connection pool exhaustion

#### Diagnosis
```bash
# Check memory usage
docker stats inventory-service

# Monitor goroutines
curl http://inventory-service:8083/debug/pprof/goroutine?debug=1

# Database connection pool
SELECT COUNT(*) as active_connections,
       MAX(application_name) as app_name
FROM pg_stat_activity 
WHERE datname = 'inventory_db'
GROUP BY application_name;
```

#### Solutions
1. **Memory Optimization**
```go
// Implement pagination for large datasets
func (s *InventoryService) GetAllProducts(limit, offset int) (*ProductList, error) {
    if limit > 1000 {
        limit = 1000 // Prevent large memory allocations
    }
    
    return s.repo.GetProducts(limit, offset)
}

// Use streaming for large exports
func (s *InventoryService) ExportMovements(w io.Writer, filter MovementFilter) error {
    rows, err := s.db.Query(buildMovementQuery(filter))
    if err != nil {
        return err
    }
    defer rows.Close()
    
    encoder := json.NewEncoder(w)
    for rows.Next() {
        var movement StockMovement
        if err := rows.Scan(&movement); err != nil {
            return err
        }
        if err := encoder.Encode(movement); err != nil {
            return err
        }
    }
    
    return nil
}
```

2. **Database Connection Management**
```go
// Configure connection pool
func NewDatabase() *sql.DB {
    db, err := sql.Open("postgres", connString)
    if err != nil {
        panic(err)
    }
    
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    return db
}
```

## Monitoring and Alerts

### Key Metrics to Monitor

1. **Stock Level Metrics**
```prometheus
# Low stock items count
inventory_low_stock_items{store_id="store1"} 5

# Total inventory value
inventory_total_value{store_id="store1"} 125000.50

# Stock movement frequency
inventory_movements_total{type="sale",store_id="store1"} 150
```

2. **Performance Metrics**
```prometheus
# API response time
inventory_api_duration_seconds{endpoint="/products",method="GET"} 0.250

# Database query time
inventory_db_query_duration_seconds{query="get_stock_levels"} 0.050

# Cache hit rate
inventory_cache_hit_rate{cache="stock_levels"} 0.85
```

3. **Error Metrics**
```prometheus
# Integration failures
inventory_integration_errors_total{service="loyverse"} 3

# Stock reservation failures
inventory_reservation_failures_total{reason="insufficient_stock"} 12

# Database errors
inventory_db_errors_total{operation="insert"} 1
```

### Alerting Rules

```yaml
groups:
  - name: inventory_alerts
    rules:
      - alert: HighLowStockItems
        expr: inventory_low_stock_items > 20
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High number of low stock items"
          
      - alert: InventoryAPIErrors
        expr: rate(inventory_api_errors_total[5m]) > 0.1
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High error rate in inventory API"
          
      - alert: StockMovementFailures
        expr: inventory_stock_movement_failures_total > 5
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Stock movement recording failures"
```

## Debug Endpoints

### Development Debug Endpoints
```bash
# Check cache status
GET /debug/cache/status

# View recent errors
GET /debug/errors/recent

# Force cache clear
POST /debug/cache/clear

# Manual stock recalculation
POST /debug/stock/recalculate
{
  "product_id": "uuid",
  "store_id": "uuid"
}

# View current reservations
GET /debug/reservations

# Health check with detailed status
GET /debug/health/detailed
```

## Recovery Procedures

### Emergency Stock Reset
```sql
-- Backup current state
CREATE TABLE stock_levels_backup AS SELECT * FROM stock_levels;
CREATE TABLE stock_movements_backup AS SELECT * FROM stock_movements;

-- Reset stock levels to zero
UPDATE stock_levels SET quantity_on_hand = 0;

-- Recalculate from movements
UPDATE stock_levels sl
SET quantity_on_hand = COALESCE((
    SELECT SUM(sm.quantity)
    FROM stock_movements sm
    WHERE sm.product_id = sl.product_id 
      AND sm.store_id = sl.store_id
), 0);
```

### Service Recovery
```bash
# Restart service with clean state
docker restart inventory-service

# Clear Redis cache
redis-cli FLUSHDB

# Verify service health
curl http://inventory-service:8083/health

# Resync critical stock data
curl -X POST http://inventory-service:8083/api/inventory/sync/full
```
