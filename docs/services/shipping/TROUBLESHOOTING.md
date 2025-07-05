# Shipping Service Troubleshooting Guide

## Common Issues & Solutions

### 1. Delivery Creation Issues

#### Problem: Delivery creation fails with "Invalid delivery method"
**Symptoms:**
- API returns 400 Bad Request
- Error message: "Invalid delivery method"
- Delivery method is one of the supported types

**Diagnosis:**
```bash
# Check if delivery method is properly configured
curl -X GET "http://localhost:8086/api/v1/providers?active=true" \
  -H "Content-Type: application/json"

# Verify delivery method enum values
grep -r "DeliveryMethod" services/shipping/internal/domain/entity/
```

**Root Causes:**
1. Delivery method not in enum constants
2. Provider not configured for the method
3. Case sensitivity mismatch

**Solutions:**
```go
// 1. Verify enum constants in delivery.go
const (
    DeliveryMethodSelfDelivery   DeliveryMethod = "self_delivery"
    DeliveryMethodGrab           DeliveryMethod = "grab"
    DeliveryMethodLineMan        DeliveryMethod = "lineman"
    // ... other methods
)

// 2. Check provider configuration
func (d *DeliveryUsecase) validateDeliveryMethod(method DeliveryMethod) error {
    if method == DeliveryMethodSelfDelivery {
        return nil // Always available
    }
    
    provider, err := d.providerRepo.GetByCode(string(method))
    if err != nil {
        return fmt.Errorf("provider not found for method: %s", method)
    }
    
    if !provider.IsActive {
        return fmt.Errorf("provider inactive for method: %s", method)
    }
    
    return nil
}
```

#### Problem: "Address not in coverage area" error
**Symptoms:**
- Delivery creation fails for valid addresses
- Coverage check returns false for known service areas

**Diagnosis:**
```bash
# Test coverage check directly
curl -X POST "http://localhost:8086/api/v1/coverage/check" \
  -H "Content-Type: application/json" \
  -d '{
    "address": {
      "lat": 13.7563,
      "lng": 100.5018
    },
    "delivery_method": "grab"
  }'

# Check coverage area configuration
psql -d saan_shipping -c "SELECT provider_name, coverage_areas FROM delivery_providers WHERE is_active = true;"
```

**Root Causes:**
1. Coverage area data not properly loaded
2. Coordinate precision issues
3. Provider coverage areas not configured

**Solutions:**
```sql
-- Update coverage areas for Bangkok
UPDATE delivery_providers 
SET coverage_areas = '{
  "zones": [
    {
      "name": "Central Bangkok",
      "polygon": [
        [13.7000, 100.4500],
        [13.8000, 100.4500],
        [13.8000, 100.6000],
        [13.7000, 100.6000],
        [13.7000, 100.4500]
      ]
    }
  ]
}'
WHERE provider_code = 'GRAB';
```

### 2. Vehicle Management Issues

#### Problem: Vehicle location updates not working
**Symptoms:**
- Vehicle location shows as null or outdated
- Real-time tracking not functioning

**Diagnosis:**
```bash
# Check vehicle location update endpoint
curl -X PATCH "http://localhost:8086/api/v1/vehicles/{vehicle_id}/location" \
  -H "Content-Type: application/json" \
  -d '{
    "latitude": 13.7563,
    "longitude": 100.5018,
    "timestamp": "2024-01-01T14:30:00Z"
  }'

# Verify location data format
psql -d saan_shipping -c "SELECT id, license_plate, current_location FROM delivery_vehicles WHERE current_location IS NOT NULL;"
```

**Root Causes:**
1. Invalid JSON format for location data
2. Timestamp parsing issues
3. Database constraint violations

**Solutions:**
```go
// Proper location update implementation
type LocationUpdate struct {
    Latitude  float64   `json:"latitude" validate:"required,min=-90,max=90"`
    Longitude float64   `json:"longitude" validate:"required,min=-180,max=180"`
    Timestamp time.Time `json:"timestamp" validate:"required"`
    Accuracy  *float64  `json:"accuracy,omitempty"`
    Speed     *float64  `json:"speed,omitempty"`
    Bearing   *float64  `json:"bearing,omitempty"`
}

func (v *VehicleHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
    var req LocationUpdate
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeBadRequestError(w, r, "Invalid location data")
        return
    }
    
    if err := validate.Struct(&req); err != nil {
        writeValidationError(w, r, err)
        return
    }
    
    locationJSON := map[string]interface{}{
        "lat": req.Latitude,
        "lng": req.Longitude,
        "timestamp": req.Timestamp,
        "accuracy": req.Accuracy,
        "speed": req.Speed,
        "bearing": req.Bearing,
    }
    
    err := v.vehicleUsecase.UpdateLocation(r.Context(), vehicleID, locationJSON)
    if err != nil {
        writeInternalServerError(w, r, err)
        return
    }
    
    writeJSONResponse(w, r, http.StatusOK, map[string]string{
        "message": "Location updated successfully"
    })
}
```

### 3. Route Optimization Issues

#### Problem: Route optimization fails or produces poor results
**Symptoms:**
- Optimization takes too long to complete
- Routes have obvious inefficiencies
- System times out during optimization

**Diagnosis:**
```bash
# Check route optimization logs
docker logs shipping-service | grep "route.optimization"

# Test optimization with small dataset
curl -X POST "http://localhost:8086/api/v1/routes/{route_id}/optimize" \
  -H "Content-Type: application/json" \
  -d '{
    "algorithm": "genetic_algorithm",
    "constraints": {
      "max_duration_hours": 8,
      "vehicle_capacity": 100.0,
      "time_windows": true
    }
  }'
```

**Root Causes:**
1. Too many delivery points for algorithm
2. Insufficient vehicle capacity constraints
3. Conflicting time window requirements
4. Memory or CPU limitations

**Solutions:**
```go
// Implement route optimization with proper constraints
func (r *RoutingUsecase) OptimizeRoute(ctx context.Context, routeID uuid.UUID, params OptimizationParams) error {
    route, err := r.routeRepo.GetByID(ctx, routeID)
    if err != nil {
        return err
    }
    
    deliveries, err := r.deliveryRepo.GetByRouteID(ctx, routeID)
    if err != nil {
        return err
    }
    
    // Limit optimization to reasonable number of stops
    if len(deliveries) > 25 {
        return errors.New("too many deliveries for optimization, maximum 25 allowed")
    }
    
    // Apply vehicle constraints
    vehicle, err := r.vehicleRepo.GetByID(ctx, *route.AssignedVehicleID)
    if err != nil {
        return err
    }
    
    totalWeight := calculateTotalWeight(deliveries)
    if totalWeight > vehicle.MaxWeight {
        return errors.New("total weight exceeds vehicle capacity")
    }
    
    // Run optimization with timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()
    
    optimizedSequence, err := r.optimizer.OptimizeSequence(ctx, deliveries, vehicle, params)
    if err != nil {
        return err
    }
    
    // Update route with optimized sequence
    return r.routeRepo.UpdateOptimization(ctx, routeID, optimizedSequence)
}
```

### 4. Provider Integration Issues

#### Problem: Third-party API calls failing
**Symptoms:**
- Delivery creation works for self-delivery but fails for external providers
- API timeouts or connection errors
- Authentication failures

**Diagnosis:**
```bash
# Test provider API connectivity
curl -X GET "https://partner-api.grab.com/grabexpress/v1/health" \
  -H "Authorization: Bearer $GRAB_ACCESS_TOKEN"

# Check API credentials configuration
echo $GRAB_CLIENT_ID
echo $GRAB_CLIENT_SECRET

# Verify webhook endpoints
curl -X POST "http://localhost:8086/webhooks/grab/status-update" \
  -H "Content-Type: application/json" \
  -H "X-Grab-Signature: test_signature" \
  -d '{"event": "test"}'
```

**Root Causes:**
1. Invalid API credentials
2. Network connectivity issues
3. API endpoint changes
4. Rate limiting
5. Webhook signature validation failures

**Solutions:**
```go
// Implement robust API client with retry logic
type GrabAPIClient struct {
    client     *http.Client
    baseURL    string
    clientID   string
    secret     string
    circuitBreaker *CircuitBreaker
}

func (g *GrabAPIClient) CreateDelivery(ctx context.Context, req GrabDeliveryRequest) (*GrabDeliveryResponse, error) {
    // Circuit breaker protection
    if g.circuitBreaker.IsOpen() {
        return nil, errors.New("grab API circuit breaker is open")
    }
    
    // Get access token with retry
    token, err := g.getAccessTokenWithRetry(ctx, 3)
    if err != nil {
        g.circuitBreaker.RecordFailure()
        return nil, fmt.Errorf("failed to get access token: %w", err)
    }
    
    // Create HTTP request
    body, _ := json.Marshal(req)
    httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/deliveries", bytes.NewBuffer(body))
    if err != nil {
        return nil, err
    }
    
    httpReq.Header.Set("Authorization", "Bearer "+token)
    httpReq.Header.Set("Content-Type", "application/json")
    
    // Execute with timeout
    resp, err := g.client.Do(httpReq)
    if err != nil {
        g.circuitBreaker.RecordFailure()
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        g.circuitBreaker.RecordFailure()
        return nil, fmt.Errorf("grab API error: %d", resp.StatusCode)
    }
    
    var result GrabDeliveryResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    g.circuitBreaker.RecordSuccess()
    return &result, nil
}

// Webhook signature validation
func (g *GrabAPIClient) ValidateWebhookSignature(payload []byte, signature string) bool {
    mac := hmac.New(sha256.New, []byte(g.secret))
    mac.Write(payload)
    expectedSignature := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
```

### 5. Database Performance Issues

#### Problem: Slow delivery queries
**Symptoms:**
- API response times > 5 seconds
- Database CPU usage high
- Query timeouts

**Diagnosis:**
```sql
-- Check slow queries
SELECT query, mean_time, calls, total_time 
FROM pg_stat_statements 
WHERE query LIKE '%delivery_orders%' 
ORDER BY mean_time DESC 
LIMIT 10;

-- Analyze table statistics
ANALYZE delivery_orders;
EXPLAIN ANALYZE SELECT * FROM delivery_orders WHERE status = 'in_transit' AND planned_delivery_date = CURRENT_DATE;

-- Check index usage
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch 
FROM pg_stat_user_indexes 
WHERE tablename = 'delivery_orders';
```

**Root Causes:**
1. Missing database indexes
2. Inefficient query patterns
3. Large result sets without pagination
4. Outdated table statistics

**Solutions:**
```sql
-- Add performance indexes
CREATE INDEX CONCURRENTLY idx_delivery_orders_status_date 
ON delivery_orders (status, planned_delivery_date) 
WHERE is_active = true;

CREATE INDEX CONCURRENTLY idx_delivery_orders_tracking 
ON delivery_orders (tracking_number) 
WHERE tracking_number IS NOT NULL;

CREATE INDEX CONCURRENTLY idx_delivery_orders_customer_date 
ON delivery_orders (customer_id, created_at DESC) 
WHERE is_active = true;

CREATE INDEX CONCURRENTLY idx_delivery_orders_vehicle_route 
ON delivery_orders (vehicle_id, route_id) 
WHERE vehicle_id IS NOT NULL;

-- Optimize common queries
-- Instead of: SELECT * FROM delivery_orders WHERE status = 'in_transit'
-- Use: 
SELECT id, order_id, tracking_number, status, estimated_delivery_time 
FROM delivery_orders 
WHERE status = 'in_transit' 
  AND is_active = true 
  AND planned_delivery_date >= CURRENT_DATE - INTERVAL '7 days'
ORDER BY created_at DESC 
LIMIT 50;
```

### 6. Cache Issues

#### Problem: Inconsistent cached data
**Symptoms:**
- Old delivery status shown in API responses
- Rate calculations don't update after provider changes
- Cache hit rate very low

**Diagnosis:**
```bash
# Check Redis connectivity
redis-cli -h localhost -p 6379 ping

# Inspect cache keys
redis-cli -h localhost -p 6379 --scan --pattern "shipping:*"

# Check cache hit rates
redis-cli -h localhost -p 6379 info stats | grep cache

# Verify cache TTL
redis-cli -h localhost -p 6379 ttl "shipping:delivery:123:tracking"
```

**Root Causes:**
1. Cache invalidation not working properly
2. Inconsistent cache key generation
3. Race conditions in cache updates
4. Memory pressure causing evictions

**Solutions:**
```go
// Implement proper cache invalidation
type ShippingCacheManager struct {
    redis   redis.Client
    keyGen  *CacheKeyGenerator
}

func (c *ShippingCacheManager) InvalidateDeliveryCache(deliveryID uuid.UUID) error {
    patterns := []string{
        fmt.Sprintf("shipping:delivery:%s:*", deliveryID),
        fmt.Sprintf("shipping:tracking:%s:*", deliveryID),
    }
    
    for _, pattern := range patterns {
        keys, err := c.redis.Keys(context.Background(), pattern).Result()
        if err != nil {
            return err
        }
        
        if len(keys) > 0 {
            if err := c.redis.Del(context.Background(), keys...).Err(); err != nil {
                return err
            }
        }
    }
    
    return nil
}

// Cache with proper TTL and invalidation
func (c *ShippingCacheManager) CacheDeliveryTracking(deliveryID uuid.UUID, tracking TrackingData) error {
    key := c.keyGen.DeliveryTracking(deliveryID)
    data, err := json.Marshal(tracking)
    if err != nil {
        return err
    }
    
    // Cache for 10 minutes with automatic expiration
    return c.redis.SetEX(context.Background(), key, data, 10*time.Minute).Err()
}

// Implement cache warming for frequently accessed data
func (c *ShippingCacheManager) WarmActiveDeliveryCache() error {
    activeDeliveries, err := c.deliveryRepo.GetActive(context.Background())
    if err != nil {
        return err
    }
    
    for _, delivery := range activeDeliveries {
        tracking, err := c.generateTrackingData(delivery)
        if err != nil {
            continue // Skip errors, don't fail entire warming
        }
        
        c.CacheDeliveryTracking(delivery.ID, tracking)
    }
    
    return nil
}
```

### 7. Monitoring & Alerting

#### Problem: Missing visibility into delivery performance
**Symptoms:**
- No alerts when deliveries fail
- Performance degradation not detected
- Customer complaints about delivery issues

**Solutions:**
```go
// Implement comprehensive metrics
type ShippingMetrics struct {
    deliverySuccessRate    *prometheus.GaugeVec
    deliveryDuration      *prometheus.HistogramVec
    providerResponseTime  *prometheus.HistogramVec
    activeDeliveries      *prometheus.GaugeVec
    failedDeliveries      *prometheus.CounterVec
}

func NewShippingMetrics() *ShippingMetrics {
    return &ShippingMetrics{
        deliverySuccessRate: prometheus.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "shipping_delivery_success_rate",
                Help: "Success rate of deliveries by method",
            },
            []string{"method", "provider"},
        ),
        deliveryDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "shipping_delivery_duration_hours",
                Help: "Duration of deliveries in hours",
                Buckets: []float64{1, 4, 8, 24, 48, 72},
            },
            []string{"method", "provider"},
        ),
        // ... other metrics
    }
}

// Set up health checks
func (s *ShippingService) HealthCheck() map[string]string {
    health := make(map[string]string)
    
    // Database connectivity
    if err := s.db.Ping(); err != nil {
        health["database"] = "unhealthy: " + err.Error()
    } else {
        health["database"] = "healthy"
    }
    
    // Redis connectivity
    if err := s.redis.Ping().Err(); err != nil {
        health["redis"] = "unhealthy: " + err.Error()
    } else {
        health["redis"] = "healthy"
    }
    
    // Provider API health
    for _, provider := range s.providers {
        if err := provider.HealthCheck(); err != nil {
            health[provider.Name()] = "unhealthy: " + err.Error()
        } else {
            health[provider.Name()] = "healthy"
        }
    }
    
    return health
}
```

## Emergency Procedures

### 1. Provider API Outage
```bash
# 1. Check provider status
curl -X GET "http://localhost:8086/api/v1/providers" | jq '.data[] | select(.is_active == true)'

# 2. Disable affected provider temporarily
curl -X PATCH "http://localhost:8086/api/v1/providers/{provider_id}" \
  -H "Content-Type: application/json" \
  -d '{"is_active": false}'

# 3. Redirect deliveries to backup providers
curl -X POST "http://localhost:8086/internal/failover" \
  -H "Content-Type: application/json" \
  -d '{
    "failed_provider": "grab",
    "backup_provider": "lineman"
  }'
```

### 2. Database Performance Crisis
```sql
-- Emergency performance fixes
-- 1. Kill long-running queries
SELECT pg_terminate_backend(pid) 
FROM pg_stat_activity 
WHERE datname = 'saan_shipping' 
  AND state = 'active' 
  AND query_start < NOW() - INTERVAL '5 minutes'
  AND query NOT LIKE '%pg_stat_activity%';

-- 2. Temporary index creation for immediate relief
CREATE INDEX CONCURRENTLY idx_emergency_delivery_status 
ON delivery_orders (status) 
WHERE is_active = true;

-- 3. Clear cache to force fresh data
TRUNCATE TABLE pg_stat_statements;
```

### 3. Service Recovery
```bash
# 1. Restart service with health checks
docker-compose restart shipping-service

# 2. Verify service health
curl -X GET "http://localhost:8086/health"

# 3. Check critical endpoints
curl -X GET "http://localhost:8086/ready"

# 4. Monitor error rates
docker logs shipping-service --tail=100 | grep ERROR

# 5. Warm up cache
curl -X POST "http://localhost:8086/internal/cache/warm"
```

## Debugging Tools

### 1. Log Analysis
```bash
# Filter delivery creation issues
docker logs shipping-service | grep "delivery.creation" | jq '.'

# Track specific delivery
docker logs shipping-service | grep "delivery_id:abc-123" | tail -20

# Monitor API performance
docker logs shipping-service | grep "http.request" | awk '{print $6}' | sort -n
```

### 2. Database Debugging
```sql
-- Track delivery status changes
SELECT d.id, d.status, d.updated_at, 
       LAG(d.status) OVER (PARTITION BY d.id ORDER BY d.updated_at) as previous_status
FROM delivery_orders d
WHERE d.id = 'delivery-uuid'
ORDER BY d.updated_at DESC;

-- Find stuck deliveries
SELECT id, order_id, status, created_at, updated_at,
       EXTRACT(epoch FROM (NOW() - updated_at))/3600 as hours_since_update
FROM delivery_orders
WHERE status IN ('dispatched', 'in_transit')
  AND updated_at < NOW() - INTERVAL '4 hours'
  AND is_active = true;
```

### 3. Performance Profiling
```go
// Enable pprof endpoints for debugging
import _ "net/http/pprof"

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // Rest of application
}
```

```bash
# CPU profiling
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profiling
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

This troubleshooting guide provides comprehensive solutions for common issues in the Shipping Service, along with emergency procedures and debugging tools for maintaining service reliability.
