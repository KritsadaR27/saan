# Order Service Troubleshooting Guide

## 🚨 **Common Issues & Solutions**

### **Order Creation Failures**

#### **Issue**: "Product validation failed"
```
Error: 422 Unprocessable Entity
Message: "Product 'product_123' not found or unavailable"
```

**Root Causes:**
- Product Service is down
- Product has been deleted from catalog
- Product is marked as unavailable
- Network connectivity issues

**Solutions:**
```bash
# 1. Check Product Service health
curl http://product:8083/health

# 2. Verify product exists
curl http://product:8083/api/v1/products/product_123

# 3. Check Product Service logs
docker logs product

# 4. Clear product cache if stale
redis-cli DEL "product:product_123"
redis-cli DEL "order:pricing:product_123:*"

# 5. Check network connectivity
curl -v http://product:8083/api/v1/health
```

#### **Issue**: "Customer validation failed"
```
Error: 404 Not Found  
Message: "Customer 'customer_456' not found"
```

**Root Causes:**
- Customer Service is down
- Customer ID is invalid
- Customer has been deleted
- Database connection issues

**Solutions:**
```bash
# 1. Verify Customer Service
curl http://customer:8110/health

# 2. Check customer exists
curl http://customer:8110/api/v1/customers/customer_456

# 3. Check database connection
docker exec postgres pg_isready -U saan

# 4. Clear customer cache
redis-cli DEL "order:customer:customer_456"

# 5. Check Customer Service logs
docker logs customer
```

#### **Issue**: "Inventory check timeout"
```
Error: 503 Service Unavailable
Message: "Unable to verify product availability"
```

**Root Causes:**
- Inventory Service is slow/down
- High inventory service load
- Database deadlocks

**Solutions:**
```bash
# 1. Check Inventory Service status
curl http://inventory:8082/health

# 2. Check service load
docker stats inventory

# 3. Clear inventory cache
redis-cli DEL "inventory:*"

# 4. Check for database locks
docker exec postgres psql -U saan -d saan_db -c "
  SELECT pid, query, state 
  FROM pg_stat_activity 
  WHERE state = 'active';"

# 5. Restart if needed
docker restart inventory
```

---

## 💳 **Payment Issues**

#### **Issue**: "Payment processing failed"
```
Error: 500 Internal Server Error
Message: "Payment gateway error"
```

**Root Causes:**
- Payment Service is down
- Payment gateway connectivity
- Invalid payment credentials
- Insufficient funds

**Solutions:**
```bash
# 1. Check Payment Service
curl http://payment:8087/health

# 2. Verify payment configuration
echo $STRIPE_SECRET_KEY | wc -c  # Should be > 50 chars

# 3. Test payment gateway directly
curl -X POST https://api.stripe.com/v1/payment_methods \
  -u ${STRIPE_SECRET_KEY}: \
  -d type=card

# 4. Check payment logs
docker logs payment | grep ERROR

# 5. Enable COD fallback
redis-cli SET "payment:fallback:cod" "enabled"
```

#### **Issue**: "Payment confirmed but order not updated"
```
Order Status: pending
Payment Status: completed
```

**Root Causes:**
- Event delivery failure
- Kafka consumer lag
- Race condition

**Solutions:**
```bash
# 1. Check Kafka topic lag
docker exec kafka kafka-consumer-groups.sh \
  --bootstrap-server kafka:9092 \
  --describe --group order-service

# 2. Manual order status update
curl -X PATCH http://order:8081/api/v1/orders/order_123/status \
  -H "Content-Type: application/json" \
  -d '{"status": "confirmed"}'

# 3. Republish payment event
curl -X POST http://payment:8087/api/v1/payments/pay_456/republish

# 4. Check event processing logs
docker logs order | grep "payment.confirmed"
```

---

## 🚚 **Delivery Issues**

#### **Issue**: "Delivery cost calculation failed"
```
Error: 503 Service Unavailable
Message: "Unable to calculate delivery fee"
```

**Root Causes:**
- Shipping Service is down
- External shipping API limits
- Invalid address format

**Solutions:**
```bash
# 1. Use fallback delivery fee
redis-cli SET "shipping:fallback:fee" "35.00"

# 2. Check Shipping Service
curl http://shipping:8086/health

# 3. Validate address format
curl -X POST http://shipping:8086/api/v1/address/validate \
  -d '{"address": "123 Main St", "district": "Downtown"}'

# 4. Check external API limits
docker logs shipping | grep "rate limit"

# 5. Switch to pickup mode temporarily
# Update frontend to promote pickup orders
```

#### **Issue**: "Delivery tracking not working"
```
Error: Order shows "delivering" but no tracking info
```

**Root Causes:**
- Driver app offline
- GPS tracking disabled
- Manual delivery process

**Solutions:**
```bash
# 1. Check driver status
curl http://shipping:8086/api/v1/drivers/driver_123/status

# 2. Manual tracking update
curl -X POST http://shipping:8086/api/v1/deliveries/del_456/location \
  -d '{"lat": 13.7563, "lng": 100.5018, "timestamp": "2024-01-15T11:00:00Z"}'

# 3. Contact driver directly
# Phone number in order delivery details

# 4. Enable SMS notifications
curl -X POST http://shipping:8086/api/v1/deliveries/del_456/notify \
  -d '{"type": "sms", "message": "Your order is on the way"}'
```

---

## 💾 **Database Issues**

#### **Issue**: "Database connection timeout"
```
Error: 500 Internal Server Error
Message: "dial tcp [::1]:5432: connect: connection refused"
```

**Root Causes:**
- PostgreSQL is down
- Connection pool exhausted
- Network connectivity issues

**Solutions:**
```bash
# 1. Check PostgreSQL status
docker exec postgres pg_isready -U saan

# 2. Restart PostgreSQL if needed
docker restart postgres

# 3. Check connection pool
curl http://order:8081/metrics | grep db_connections

# 4. Clear connection pool
docker restart order

# 5. Check PostgreSQL logs
docker logs postgres | tail -50
```

#### **Issue**: "Order data inconsistency"
```
Problem: Order total doesn't match item prices
Order Total: 535.00
Calculated Total: 425.00
```

**Root Causes:**
- Price changes during order creation
- VIP discount calculation error
- Tax calculation bug

**Solutions:**
```sql
-- 1. Recalculate order total
UPDATE orders 
SET total = (
  SELECT SUM(oi.quantity * oi.unit_price) + COALESCE(delivery_fee, 0)
  FROM order_items oi 
  WHERE oi.order_id = orders.id
)
WHERE id = 'order_123';

-- 2. Check price history
SELECT * FROM product_price_history 
WHERE product_id = 'product_456' 
AND created_at BETWEEN '2024-01-15 10:00:00' AND '2024-01-15 11:00:00';

-- 3. Audit order calculation
SELECT 
  o.id,
  o.total as order_total,
  SUM(oi.quantity * oi.unit_price) as items_total,
  o.delivery_fee,
  o.tax_amount,
  o.discount_amount
FROM orders o
JOIN order_items oi ON o.id = oi.order_id
WHERE o.id = 'order_123'
GROUP BY o.id;
```

---

## 🗄️ **Cache Issues**

#### **Issue**: "Stale product prices in orders"
```
Problem: Customer sees old price, but order created with new price
```

**Root Causes:**
- Cache not invalidated after price update
- Redis memory pressure
- Cache TTL too long

**Solutions:**
```bash
# 1. Clear all product pricing cache
redis-cli --scan --pattern "order:pricing:*" | xargs redis-cli DEL

# 2. Check Redis memory usage
redis-cli INFO memory

# 3. Increase cache refresh frequency
redis-cli CONFIG SET maxmemory-policy allkeys-lru

# 4. Force cache invalidation after price changes
curl -X POST http://product:8083/api/v1/products/product_456/invalidate-cache
```

#### **Issue**: "Redis connection failures"
```
Error: "dial tcp redis:6379: connect: connection refused"
```

**Root Causes:**
- Redis container is down
- Redis out of memory
- Network connectivity

**Solutions:**
```bash
# 1. Check Redis status
docker exec redis redis-cli ping

# 2. Restart Redis if needed
docker restart redis

# 3. Check Redis logs
docker logs redis

# 4. Clear Redis data if corrupted
docker exec redis redis-cli FLUSHDB

# 5. Monitor Redis metrics
docker exec redis redis-cli INFO stats
```

---

## 📊 **Performance Issues**

#### **Issue**: "Slow order creation (>5 seconds)"
```
Problem: Order creation taking too long
Response Time: 8.5 seconds
Expected: <2 seconds
```

**Root Causes:**
- External service latency
- Database query optimization needed
- Network bottlenecks

**Solutions:**
```bash
# 1. Check external service response times
curl -w "@curl-format.txt" http://product:8083/api/v1/products/123

# 2. Profile database queries
docker exec postgres psql -U saan -d saan_db -c "
  SELECT query, mean_time, calls 
  FROM pg_stat_statements 
  ORDER BY mean_time DESC LIMIT 10;"

# 3. Check network latency
ping product
ping customer
ping payment

# 4. Enable query optimization
docker exec postgres psql -U saan -d saan_db -c "
  SET log_min_duration_statement = 1000;"

# 5. Scale services if needed
docker-compose up --scale order=2
```

#### **Issue**: "High memory usage"
```
Problem: Order Service consuming >2GB RAM
Expected: <500MB
```

**Root Causes:**
- Memory leaks
- Large cache objects
- Goroutine leaks

**Solutions:**
```bash
# 1. Check memory usage
docker stats order

# 2. Profile memory usage
curl http://order:8081/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# 3. Check for goroutine leaks
curl http://order:8081/debug/pprof/goroutine?debug=1

# 4. Clear cache to free memory
redis-cli FLUSHDB

# 5. Restart service
docker restart order
```

---

## 🔧 **Configuration Issues**

#### **Issue**: "Service discovery failures"
```
Error: "no such host: product"
```

**Root Causes:**
- Docker network configuration
- Service name mismatch
- Container not in same network

**Solutions:**
```bash
# 1. Check Docker network
docker network ls
docker network inspect saan_saan-network

# 2. Verify service is running
docker ps | grep product

# 3. Test DNS resolution
docker exec order nslookup product

# 4. Check service names in docker-compose.yml
grep "container_name" docker-compose.yml

# 5. Restart network stack
docker-compose down && docker-compose up -d
```

---

## 📱 **API Gateway Issues**

#### **Issue**: "CORS errors from frontend"
```
Error: "Access to fetch blocked by CORS policy"
```

**Root Causes:**
- CORS headers not configured
- Wrong API gateway configuration
- Development vs production settings

**Solutions:**
```bash
# 1. Check CORS headers
curl -v -H "Origin: http://localhost:3008" \
  http://localhost:8081/api/v1/orders

# 2. Update CORS configuration
# In order service: Allow localhost:3008, localhost:3010

# 3. Check API gateway routing
curl http://localhost:8080/api/order/v1/health

# 4. Verify environment configuration
echo $CORS_ALLOWED_ORIGINS
```

---

## 🚑 **Emergency Procedures**

### **Complete Service Outage**
```bash
# 1. Check all dependencies
curl http://postgres:5432  # Should connection refuse (good)
curl http://redis:6379     # Should connection refuse (good)  
docker exec postgres pg_isready -U saan
docker exec redis redis-cli ping

# 2. Restart in dependency order
docker restart postgres redis kafka
# Wait 30 seconds
docker restart product customer inventory payment shipping
# Wait 30 seconds  
docker restart order

# 3. Verify health
curl http://order:8081/health
```

### **Data Recovery**
```sql
-- Backup before recovery
pg_dump -U saan -h localhost -p 5432 saan_db > backup_$(date +%Y%m%d_%H%M%S).sql

-- Find problematic orders
SELECT id, status, created_at, total 
FROM orders 
WHERE status = 'pending' 
AND created_at < NOW() - INTERVAL '1 hour';

-- Manual order completion (if payment confirmed)
UPDATE orders 
SET status = 'completed', completed_at = NOW()
WHERE id IN ('order_123', 'order_456');
```

---

## 📞 **Escalation Contacts**

### **Immediate Response Needed**
- **Payment Issues**: DevOps team + Finance team
- **Database Corruption**: Senior Backend Engineer
- **Complete Outage**: All hands + Management notification

### **Business Hours Response**
- **Performance Issues**: Backend team
- **Integration Issues**: Service owner + Integration team
- **Cache Issues**: DevOps team

---

## 📝 **Debugging Checklist**

### **Before Escalating**
- [ ] Check service health endpoints
- [ ] Verify service logs (last 100 lines)
- [ ] Test with curl commands  
- [ ] Check Redis/Database connectivity
- [ ] Verify configuration variables
- [ ] Clear relevant cache keys
- [ ] Check external service status

### **Information to Gather**
- [ ] Exact error message and stack trace
- [ ] Request/Response examples that fail
- [ ] Time when issue started
- [ ] Recent deployments or changes
- [ ] Customer impact assessment
- [ ] Steps already taken to resolve

---

> 🚨 **When in doubt, restart services in dependency order: Database → Cache → External Services → Order Service**
