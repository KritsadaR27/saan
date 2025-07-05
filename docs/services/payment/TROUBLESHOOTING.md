# Payment Service Troubleshooting Guide

## Common Issues and Solutions

### 1. Payment Gateway Connection Issues

#### Symptoms
- Payment creation fails with gateway errors
- Webhooks not being received
- Payment status not updating

#### Diagnosis
```bash
# Check gateway connectivity
curl -f https://api.omise.co/charges
curl -f https://api.c2p.com/status
curl -f https://api.truemoney.com/health

# Verify webhook endpoints
curl -X POST http://payment-service:8084/api/webhooks/payment \
  -H "Content-Type: application/json" \
  -d '{"test": "webhook"}'

# Check gateway credentials
SELECT provider_name, is_enabled, configuration 
FROM payment_providers 
WHERE is_enabled = true;
```

#### Solutions
1. **Gateway Credential Issues**
```bash
# Verify credentials in environment
echo $OMISE_SECRET_KEY | grep "skey_"
echo $OMISE_PUBLIC_KEY | grep "pkey_"

# Test credentials with gateway
curl -u $OMISE_SECRET_KEY: https://api.omise.co/account
```

2. **Network Connectivity**
```bash
# Check DNS resolution
nslookup api.omise.co
nslookup api.c2p.com

# Test SSL connectivity
openssl s_client -connect api.omise.co:443 -servername api.omise.co

# Verify firewall rules
iptables -L | grep -E "(omise|c2p|truemoney)"
```

3. **Webhook Configuration**
```sql
-- Update webhook URLs
UPDATE payment_providers 
SET configuration = jsonb_set(
    configuration, 
    '{webhook_url}', 
    '"https://your-domain.com/api/webhooks/payment"'
) 
WHERE provider_name = 'omise';
```

### 2. Payment Status Synchronization Issues

#### Symptoms
- Payments stuck in "processing" status
- Duplicate payment records
- Status updates not reflecting in order service

#### Diagnosis
```bash
# Check stuck payments
curl -H "Authorization: Bearer $TOKEN" \
  "http://payment-service:8084/api/payments?status=processing&from_date=2024-01-01"

# Verify webhook delivery logs
SELECT event_type, status, attempts, last_attempt, error_message
FROM webhook_delivery_log 
WHERE created_at > NOW() - INTERVAL '1 hour'
ORDER BY created_at DESC;

# Check status history
SELECT p.id, p.status, h.status, h.created_at
FROM payments p
JOIN payment_status_history h ON p.id = h.payment_id
WHERE p.created_at > NOW() - INTERVAL '1 day'
ORDER BY h.created_at DESC;
```

#### Solutions
1. **Manual Status Update**
```sql
-- Update stuck payment status
UPDATE payments 
SET status = 'completed',
    paid_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = 'PAYMENT_ID' 
  AND external_transaction_id = 'EXTERNAL_TXN_ID';

-- Record status change
INSERT INTO payment_status_history (payment_id, status, reason)
VALUES ('PAYMENT_ID', 'completed', 'Manual correction');
```

2. **Resync Payment Status**
```bash
# Force status check with gateway
curl -X POST -H "Authorization: Bearer $TOKEN" \
  "http://payment-service:8084/api/payments/PAYMENT_ID/sync-status"

# Bulk status sync for stuck payments
curl -X POST -H "Authorization: Bearer $TOKEN" \
  "http://payment-service:8084/api/payments/bulk-sync" \
  -d '{"status": "processing", "hours_ago": 24}'
```

3. **Fix Order Service Integration**
```go
// Retry order status update
func (s *PaymentService) retryOrderStatusUpdate(paymentID uuid.UUID) error {
    payment, err := s.repo.GetByID(paymentID)
    if err != nil {
        return err
    }
    
    return s.orderService.UpdatePaymentStatus(payment.OrderID, payment.Status)
}
```

### 3. Payment Processing Failures

#### Symptoms
- High payment failure rates
- Card declined errors
- Gateway timeout errors

#### Diagnosis
```bash
# Check failure rates by provider
SELECT 
    payment_provider,
    COUNT(*) as total_payments,
    SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_payments,
    ROUND(
        SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 
        2
    ) as failure_rate
FROM payments 
WHERE created_at > NOW() - INTERVAL '24 hours'
GROUP BY payment_provider;

# Analyze failure reasons
SELECT 
    payment_details->>'failure_reason' as failure_reason,
    COUNT(*) as count
FROM payments 
WHERE status = 'failed'
  AND created_at > NOW() - INTERVAL '24 hours'
GROUP BY payment_details->>'failure_reason'
ORDER BY count DESC;
```

#### Solutions
1. **Gateway-Specific Issues**
```bash
# Omise-specific troubleshooting
curl -u $OMISE_SECRET_KEY: \
  "https://api.omise.co/charges/CHARGE_ID"

# Check Omise account limits
curl -u $OMISE_SECRET_KEY: \
  "https://api.omise.co/account"
```

2. **Implement Retry Logic**
```go
func (s *PaymentService) processPaymentWithRetry(req PaymentRequest) error {
    maxRetries := 3
    backoff := time.Second
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        err := s.processPayment(req)
        if err == nil {
            return nil
        }
        
        // Don't retry certain errors
        if isNonRetryableError(err) {
            return err
        }
        
        if attempt < maxRetries-1 {
            time.Sleep(backoff)
            backoff *= 2 // Exponential backoff
        }
    }
    
    return errors.New("payment failed after retries")
}
```

3. **Payment Method Fallback**
```go
func (s *PaymentService) processWithFallback(req PaymentRequest) error {
    primaryProvider := s.getPrimaryProvider(req.PaymentMethod)
    err := primaryProvider.ProcessPayment(req)
    
    if err != nil && s.shouldFallback(err) {
        fallbackProvider := s.getFallbackProvider(req.PaymentMethod)
        if fallbackProvider != nil {
            return fallbackProvider.ProcessPayment(req)
        }
    }
    
    return err
}
```

### 4. Refund Processing Issues

#### Symptoms
- Refunds stuck in processing status
- Refund amount discrepancies
- Customer not receiving refunds

#### Diagnosis
```bash
# Check pending refunds
SELECT r.*, p.payment_method, p.payment_provider
FROM refunds r
JOIN payments p ON r.payment_id = p.id
WHERE r.status = 'processing'
  AND r.created_at < NOW() - INTERVAL '1 hour';

# Verify refund amounts
SELECT 
    payment_id,
    SUM(amount) as total_refunded,
    (SELECT amount FROM payments WHERE id = payment_id) as original_amount
FROM refunds 
WHERE status IN ('completed', 'processing')
GROUP BY payment_id
HAVING SUM(amount) > (SELECT amount FROM payments WHERE id = payment_id);
```

#### Solutions
1. **Manual Refund Processing**
```sql
-- Update refund status
UPDATE refunds 
SET status = 'completed',
    processed_at = CURRENT_TIMESTAMP,
    external_refund_id = 'EXTERNAL_REFUND_ID'
WHERE id = 'REFUND_ID';
```

2. **Recalculate Refund Amounts**
```sql
-- Check for over-refunds
WITH refund_totals AS (
    SELECT 
        payment_id,
        SUM(amount) as total_refunded
    FROM refunds 
    WHERE status = 'completed'
    GROUP BY payment_id
)
SELECT p.id, p.amount, rt.total_refunded,
       p.amount - rt.total_refunded as remaining_amount
FROM payments p
JOIN refund_totals rt ON p.id = rt.payment_id
WHERE rt.total_refunded > p.amount;
```

3. **Retry Failed Refunds**
```bash
# Retry refund processing
curl -X POST -H "Authorization: Bearer $TOKEN" \
  "http://payment-service:8084/api/refunds/REFUND_ID/retry"
```

### 5. Performance and Scalability Issues

#### Symptoms
- Slow payment processing
- Database timeouts
- High memory usage

#### Diagnosis
```bash
# Check API performance
curl -w "@curl-format.txt" -H "Authorization: Bearer $TOKEN" \
  "http://payment-service:8084/api/payments"

# Monitor database performance
SELECT 
    schemaname,
    tablename,
    attname,
    n_distinct,
    most_common_vals
FROM pg_stats 
WHERE schemaname = 'public' 
  AND tablename IN ('payments', 'refunds');

# Check connection pool
SELECT COUNT(*) as active_connections,
       state
FROM pg_stat_activity 
WHERE datname = 'payment_db'
GROUP BY state;
```

#### Solutions
1. **Database Optimization**
```sql
-- Add missing indexes
CREATE INDEX CONCURRENTLY idx_payments_customer_status 
ON payments (customer_id, status, created_at);

CREATE INDEX CONCURRENTLY idx_payments_external_id 
ON payments (external_payment_id) WHERE external_payment_id IS NOT NULL;

-- Partition large tables
CREATE TABLE payments_2024_01 PARTITION OF payments
FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- Update statistics
ANALYZE payments;
ANALYZE refunds;
```

2. **Caching Implementation**
```go
// Cache payment provider configurations
func (s *PaymentService) getProviderCached(name string) (*Provider, error) {
    key := fmt.Sprintf("provider:%s", name)
    
    // Try cache first
    if cached, err := s.redis.Get(key); err == nil {
        var provider Provider
        if err := json.Unmarshal([]byte(cached), &provider); err == nil {
            return &provider, nil
        }
    }
    
    // Fetch from database
    provider, err := s.repo.GetProvider(name)
    if err != nil {
        return nil, err
    }
    
    // Cache for 1 hour
    data, _ := json.Marshal(provider)
    s.redis.Set(key, data, time.Hour)
    
    return provider, nil
}
```

3. **Connection Pool Tuning**
```go
func NewDatabase() *sql.DB {
    db, err := sql.Open("postgres", connString)
    if err != nil {
        panic(err)
    }
    
    db.SetMaxOpenConns(50)  // Increased from default
    db.SetMaxIdleConns(10)  // Keep connections alive
    db.SetConnMaxLifetime(1 * time.Hour)
    
    return db
}
```

## Monitoring and Alerts

### Key Metrics to Monitor

1. **Payment Metrics**
```prometheus
# Payment success rate
payment_success_rate{provider="omise"} 0.96

# Payment processing time
payment_processing_duration_seconds{provider="omise"} 2.5

# Payment volume
payment_total{provider="omise",method="credit_card"} 1250
```

2. **Gateway Metrics**
```prometheus
# Gateway response time
payment_gateway_response_time_seconds{provider="omise"} 1.2

# Gateway error rate
payment_gateway_errors_total{provider="omise",error="timeout"} 5

# Webhook delivery success
payment_webhook_delivery_success_rate{provider="omise"} 0.98
```

3. **Business Metrics**
```prometheus
# Daily payment volume
payment_daily_volume{currency="THB"} 187500.00

# Refund rate
payment_refund_rate 0.02

# Average transaction value
payment_average_value{currency="THB"} 1500.00
```

### Alerting Rules

```yaml
groups:
  - name: payment_alerts
    rules:
      - alert: HighPaymentFailureRate
        expr: payment_failure_rate > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High payment failure rate detected"
          
      - alert: PaymentGatewayDown
        expr: payment_gateway_up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Payment gateway is unreachable"
          
      - alert: StuckPayments
        expr: payment_stuck_processing_count > 10
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Multiple payments stuck in processing"
          
      - alert: RefundProcessingDelay
        expr: refund_processing_delay_minutes > 60
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Refunds taking too long to process"
```

## Debug Endpoints

### Development Debug Endpoints
```bash
# Check payment provider status
GET /debug/providers/status

# View webhook delivery logs
GET /debug/webhooks/logs

# Manual payment sync
POST /debug/payments/{id}/sync

# Test gateway connectivity
GET /debug/gateways/test

# View payment processing metrics
GET /debug/metrics/payments

# Cache status
GET /debug/cache/status
```

### Webhook Testing
```bash
# Test webhook endpoint
curl -X POST http://payment-service:8084/api/webhooks/test \
  -H "Content-Type: application/json" \
  -H "X-Signature: test_signature" \
  -d '{
    "event_type": "payment.completed",
    "payment_id": "test_payment",
    "status": "successful"
  }'
```

## Recovery Procedures

### Payment Data Recovery
```sql
-- Backup current state
CREATE TABLE payments_backup AS SELECT * FROM payments;
CREATE TABLE refunds_backup AS SELECT * FROM refunds;

-- Recover from specific point in time
SELECT * FROM payments_backup 
WHERE created_at < '2024-01-15 10:00:00'
  AND status = 'completed';
```

### Gateway Resync
```bash
# Full payment status sync
curl -X POST -H "Authorization: Bearer $TOKEN" \
  "http://payment-service:8084/api/admin/sync-all-payments" \
  -d '{"provider": "omise", "from_date": "2024-01-01"}'

# Verify sync results
curl -H "Authorization: Bearer $TOKEN" \
  "http://payment-service:8084/api/admin/sync-status"
```

### Service Recovery
```bash
# Restart service with clean state
docker restart payment-service

# Clear Redis cache
redis-cli FLUSHDB

# Verify service health
curl http://payment-service:8084/health

# Reload provider configurations
curl -X POST -H "Authorization: Bearer $TOKEN" \
  "http://payment-service:8084/api/admin/reload-providers"
```
