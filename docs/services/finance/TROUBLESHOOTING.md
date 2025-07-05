# Finance Service Troubleshooting Guide

## 🚨 **Common Issues & Solutions**

### **Cash Flow Creation Failures**

#### **Issue**: "Location not found"
```
Error: 404 Not Found
Message: "Location 'location_123' not found"
```

**Root Causes:**
- Customer Service is down
- Invalid location ID
- Location has been deactivated
- Database connectivity issues

**Solutions:**
```bash
# 1. Check Customer Service health
curl http://customer:8110/health

# 2. Verify location exists
curl http://customer:8110/api/v1/locations/location_123

# 3. Check Customer Service logs
docker logs customer

# 4. Clear location cache if stale
redis-cli DEL "finance:location:location_123"

# 5. Check network connectivity
curl -v http://customer:8110/health
```

#### **Issue**: "Cash flow already exists for this date"
```
Error: 409 Conflict
Message: "Cash flow already exists for location 'location_123' on date '2024-01-15'"
```

**Root Causes:**
- Duplicate submission
- Previous failed deletion
- Race condition during creation

**Solutions:**
```bash
# 1. Check existing cash flow
curl http://finance:8088/api/v1/finance/cash-flows?location_id=location_123&date=2024-01-15

# 2. If status is 'draft', you can update it
curl -X PUT http://finance:8088/api/v1/finance/cash-flows/{existing_id}

# 3. If invalid, delete and recreate (admin only)
curl -X DELETE http://finance:8088/api/v1/finance/cash-flows/{existing_id}

# 4. Clear related cache
redis-cli DEL "finance:daily_sales:location_123:2024-01-15"
```

#### **Issue**: "Revenue calculation failed"
```
Error: 503 Service Unavailable
Message: "Unable to calculate revenue - Loyverse integration unavailable"
```

**Root Causes:**
- Loyverse Integration service is down
- Network connectivity to Loyverse API
- Loyverse API rate limits

**Solutions:**
```bash
# 1. Check Loyverse Integration service
curl http://loyverse:8091/health

# 2. Check Loyverse API connectivity
curl http://loyverse:8091/api/v1/status

# 3. Enable manual revenue entry mode
redis-cli SET "finance:fallback:manual_revenue" "enabled"

# 4. Create cash flow with manual revenue entry
curl -X POST http://finance:8088/api/v1/finance/cash-flows \
  -d '{"manual_revenue": true, "loyverse_receipts_total": 15000.00}'

# 5. Check Loyverse Integration logs
docker logs loyverse
```

---

## 💳 **Expense Management Issues**

#### **Issue**: "Receipt upload failed"
```
Error: 500 Internal Server Error
Message: "Failed to upload receipt image"
```

**Root Causes:**
- S3/file storage is down
- Image file too large
- Invalid image format
- Network connectivity issues

**Solutions:**
```bash
# 1. Check file storage connectivity
aws s3 ls s3://saan-finance-receipts/

# 2. Verify file size and format
# Max size: 10MB, formats: JPG, PNG, PDF
file receipt.jpg
ls -lh receipt.jpg

# 3. Test S3 upload directly
aws s3 cp test.jpg s3://saan-finance-receipts/test/

# 4. Record expense without receipt first
curl -X POST http://finance:8088/api/v1/finance/expenses \
  -d '{"amount": 1500, "description": "Fuel - receipt to follow"}'

# 5. Upload receipt separately later
curl -X PUT http://finance:8088/api/v1/finance/expenses/{id}/receipt \
  -F "receipt=@receipt.jpg"
```

#### **Issue**: "Expense approval timeout"
```
Problem: Expenses pending approval for > 24 hours
```

**Root Causes:**
- Approver notifications not sent
- Approver unavailable
- System notification issues

**Solutions:**
```bash
# 1. Check pending expenses
curl http://finance:8088/api/v1/finance/expenses?status=pending

# 2. Resend approval notifications
curl -X POST http://finance:8088/api/v1/finance/expenses/{id}/resend-notification

# 3. Escalate to higher authority (if configured)
curl -X POST http://finance:8088/api/v1/finance/expenses/{id}/escalate

# 4. Check notification service logs
docker logs notification

# 5. Manual approval (emergency)
curl -X POST http://finance:8088/api/v1/finance/expenses/{id}/approve \
  -d '{"emergency_approval": true, "reason": "System notification failure"}'
```

#### **Issue**: "Expense category validation failed"
```
Error: 422 Unprocessable Entity
Message: "Invalid expense category 'unknown'"
```

**Root Causes:**
- Invalid category name
- Category has been removed
- Typo in category name

**Solutions:**
```bash
# 1. Get valid categories
curl http://finance:8088/api/v1/finance/expenses/categories

# 2. Check if category exists
curl http://finance:8088/api/v1/finance/admin/categories | grep "unknown"

# 3. Use correct category name
# Valid categories: supplier, operational, fixed, personal

# 4. Create new category if needed (admin only)
curl -X POST http://finance:8088/api/v1/finance/admin/categories \
  -d '{"name": "new_category", "description": "New category"}'
```

---

## 💰 **Profit First Allocation Issues**

#### **Issue**: "Allocation percentages don't sum to 100%"
```
Error: 422 Unprocessable Entity
Message: "Allocation percentages must sum to 100%, current sum: 95%"
```

**Root Causes:**
- Mathematical error in percentage calculation
- Missing percentage allocation
- Rounding errors

**Solutions:**
```bash
# 1. Check current configuration
curl http://finance:8088/api/v1/finance/profit-first/config/location_123

# 2. Verify percentages sum to 100
# Example: Profit 5% + Owner Pay 10% + Tax 15% + Operating 70% = 100%

# 3. Update configuration with correct percentages
curl -X PUT http://finance:8088/api/v1/finance/profit-first/config/{id} \
  -d '{
    "profit_percentage": 5.00,
    "owner_pay_percentage": 10.00,
    "tax_percentage": 15.00,
    "operating_percentage": 70.00
  }'
```

#### **Issue**: "Negative profit allocation"
```
Error: 422 Unprocessable Entity
Message: "Cannot allocate profit - negative profit amount: -500.00"
```

**Root Causes:**
- Expenses exceed revenue
- Calculation error
- Cash flow data incomplete

**Solutions:**
```bash
# 1. Check cash flow details
curl http://finance:8088/api/v1/finance/cash-flows/{cash_flow_id}

# 2. Verify revenue and expense totals
# Total Revenue: cash_flow.total_sales
# Total Expenses: cash_flow.supplier_transfers + operational_expenses + fixed_expenses + personal_expenses

# 3. Review and correct expense entries
curl http://finance:8088/api/v1/finance/expenses?cash_flow_id={id}

# 4. Skip allocation for loss days (valid business scenario)
curl -X POST http://finance:8088/api/v1/finance/profit-first/allocations/{id}/skip \
  -d '{"reason": "Loss day - expenses exceeded revenue"}'
```

#### **Issue**: "Allocation execution failed"
```
Error: 500 Internal Server Error
Message: "Failed to execute profit allocation"
```

**Root Causes:**
- Bank transfer API failure
- Network connectivity issues
- Insufficient bank account balance

**Solutions:**
```bash
# 1. Check allocation status
curl http://finance:8088/api/v1/finance/profit-first/allocations/{id}

# 2. Retry allocation manually
curl -X POST http://finance:8088/api/v1/finance/profit-first/allocations/{id}/retry

# 3. Switch to manual transfer mode
curl -X POST http://finance:8088/api/v1/finance/profit-first/allocations/{id}/manual-mode

# 4. Check bank API connectivity
curl http://bank-api:8095/health

# 5. Create planned transfers instead
curl -X POST http://finance:8088/api/v1/finance/transfers/plan \
  -d '{"from_allocation_id": "{allocation_id}"}'
```

---

## 📊 **Reporting & Dashboard Issues**

#### **Issue**: "Daily report generation timeout"
```
Error: 504 Gateway Timeout
Message: "Report generation took longer than 30 seconds"
```

**Root Causes:**
- Large dataset processing
- Database performance issues
- Complex calculations

**Solutions:**
```bash
# 1. Check database performance
docker exec postgres psql -U saan -d saan_db -c "
  SELECT query, mean_time, calls 
  FROM pg_stat_statements 
  WHERE query LIKE '%daily_cash_flows%'
  ORDER BY mean_time DESC LIMIT 5;"

# 2. Check if report is cached
redis-cli GET "finance:daily_report:location_123:2024-01-15"

# 3. Generate report in background
curl -X POST http://finance:8088/api/v1/finance/reports/daily/generate-async \
  -d '{"location_id": "location_123", "date": "2024-01-15"}'

# 4. Clear problematic cache
redis-cli DEL "finance:daily_report:location_123:*"

# 5. Simplify report (exclude heavy calculations)
curl "http://finance:8088/api/v1/finance/reports/daily?location_id=location_123&date=2024-01-15&simple=true"
```

#### **Issue**: "Dashboard showing stale data"
```
Problem: Dashboard shows yesterday's data for current day
```

**Root Causes:**
- Cache not invalidated
- Real-time data sync issues
- Clock synchronization problems

**Solutions:**
```bash
# 1. Clear dashboard cache
redis-cli DEL "finance:dashboard_data:*"

# 2. Force cache refresh
curl -X POST http://finance:8088/api/v1/finance/dashboard/refresh

# 3. Check system time
date
docker exec finance date

# 4. Verify cash flow data is current
curl "http://finance:8088/api/v1/finance/cash-flows?date=$(date +%Y-%m-%d)"

# 5. Check if scheduled jobs are running
curl http://finance:8088/api/v1/finance/admin/scheduler/status
```

#### **Issue**: "Excel export fails"
```
Error: 500 Internal Server Error
Message: "Failed to generate Excel file"
```

**Root Causes:**
- Excel generation library issues
- Temporary disk space full
- Large dataset size

**Solutions:**
```bash
# 1. Check disk space
df -h

# 2. Check temp directory
ls -la /tmp/

# 3. Generate smaller report first
curl "http://finance:8088/api/v1/finance/reports/export/excel?type=daily&date=2024-01-15&simple=true"

# 4. Try PDF export instead
curl "http://finance:8088/api/v1/finance/reports/export/pdf?type=daily&date=2024-01-15"

# 5. Clean up old export files
find /tmp -name "*.xlsx" -mtime +1 -delete
```

---

## 🗄️ **Database Issues**

#### **Issue**: "Database connection timeout"
```
Error: 500 Internal Server Error
Message: "dial tcp postgres:5432: connect: connection refused"
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

# 3. Check connection pool status
curl http://finance:8088/metrics | grep db_connections

# 4. Clear connection pool
docker restart finance

# 5. Check PostgreSQL logs
docker logs postgres | tail -50
```

#### **Issue**: "Foreign key constraint violation"
```
Error: 500 Internal Server Error
Message: "insert or update on table violates foreign key constraint"
```

**Root Causes:**
- Referenced record doesn't exist
- Database state inconsistency
- Race condition

**Solutions:**
```sql
-- 1. Check for orphaned records
SELECT * FROM expense_transactions 
WHERE cash_flow_id NOT IN (SELECT id FROM daily_cash_flows);

-- 2. Check specific constraint
SELECT 
    conname AS constraint_name,
    conrelid::regclass AS table_name,
    confrelid::regclass AS referenced_table
FROM pg_constraint 
WHERE contype = 'f' AND conname LIKE '%expense%';

-- 3. Fix orphaned data
DELETE FROM expense_transactions 
WHERE cash_flow_id = 'invalid_cash_flow_id';

-- 4. Verify data integrity
SELECT cf.id, cf.business_date, COUNT(et.id) as expense_count
FROM daily_cash_flows cf
LEFT JOIN expense_transactions et ON cf.id = et.cash_flow_id
GROUP BY cf.id, cf.business_date
ORDER BY cf.business_date DESC;
```

#### **Issue**: "Cash flow calculation mismatch"
```
Problem: Total expenses don't match sum of individual expenses
Cash Flow Total: 5000.00
Sum of Expenses: 4500.00
```

**Root Causes:**
- Manual total override
- Calculation bug
- Data corruption

**Solutions:**
```sql
-- 1. Recalculate cash flow totals
UPDATE daily_cash_flows 
SET 
  supplier_transfers = (
    SELECT COALESCE(SUM(amount), 0) 
    FROM expense_transactions 
    WHERE cash_flow_id = daily_cash_flows.id 
    AND expense_category = 'supplier'
  ),
  operational_expenses = (
    SELECT COALESCE(SUM(amount), 0) 
    FROM expense_transactions 
    WHERE cash_flow_id = daily_cash_flows.id 
    AND expense_category = 'operational'
  ),
  fixed_expenses = (
    SELECT COALESCE(SUM(amount), 0) 
    FROM expense_transactions 
    WHERE cash_flow_id = daily_cash_flows.id 
    AND expense_category = 'fixed'
  ),
  personal_expenses = (
    SELECT COALESCE(SUM(amount), 0) 
    FROM expense_transactions 
    WHERE cash_flow_id = daily_cash_flows.id 
    AND expense_category = 'personal'
  )
WHERE id = 'cash_flow_123';

-- 2. Verify calculation
SELECT 
  cf.id,
  cf.supplier_transfers + cf.operational_expenses + cf.fixed_expenses + cf.personal_expenses AS calculated_total,
  (SELECT SUM(amount) FROM expense_transactions WHERE cash_flow_id = cf.id) AS actual_total
FROM daily_cash_flows cf
WHERE cf.id = 'cash_flow_123';

-- 3. Find discrepancies
SELECT 
  cf.id,
  cf.business_date,
  cf.location_name,
  (cf.supplier_transfers + cf.operational_expenses + cf.fixed_expenses + cf.personal_expenses) AS cf_total,
  COALESCE(SUM(et.amount), 0) AS expense_sum,
  ABS((cf.supplier_transfers + cf.operational_expenses + cf.fixed_expenses + cf.personal_expenses) - COALESCE(SUM(et.amount), 0)) AS difference
FROM daily_cash_flows cf
LEFT JOIN expense_transactions et ON cf.id = et.cash_flow_id
GROUP BY cf.id, cf.business_date, cf.location_name, cf.supplier_transfers, cf.operational_expenses, cf.fixed_expenses, cf.personal_expenses
HAVING ABS((cf.supplier_transfers + cf.operational_expenses + cf.fixed_expenses + cf.personal_expenses) - COALESCE(SUM(et.amount), 0)) > 0.01
ORDER BY difference DESC;
```

---

## 🗄️ **Cache Issues**

#### **Issue**: "Stale financial data in dashboard"
```
Problem: Dashboard shows old revenue figures
Cached: 15,000.00 (from 2 hours ago)
Actual: 18,500.00
```

**Root Causes:**
- Cache invalidation failed
- Redis memory pressure
- Cache TTL too long

**Solutions:**
```bash
# 1. Clear specific cache keys
redis-cli DEL "finance:daily_sales:location_123:$(date +%Y-%m-%d)"
redis-cli DEL "finance:dashboard_data:*"

# 2. Check Redis memory usage
redis-cli INFO memory

# 3. Force cache refresh
curl -X POST http://finance:8088/api/v1/finance/cache/refresh \
  -d '{"location_id": "location_123", "date": "2024-01-15"}'

# 4. Verify cache invalidation is working
redis-cli MONITOR | grep "finance:"

# 5. Reduce cache TTL temporarily
redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

#### **Issue**: "Redis connection failures"
```
Error: "dial tcp redis:6379: connect: connection refused"
```

**Root Causes:**
- Redis container is down
- Redis out of memory
- Network connectivity issues

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

# 5. Check Redis memory and config
docker exec redis redis-cli INFO memory
docker exec redis redis-cli CONFIG GET maxmemory
```

---

## 📊 **Performance Issues**

#### **Issue**: "Slow profit allocation calculation (>30 seconds)"
```
Problem: Daily profit allocation taking too long
Response Time: 35 seconds
Expected: <5 seconds
```

**Root Causes:**
- Complex calculations on large datasets
- Database query optimization needed
- Insufficient database indexes

**Solutions:**
```bash
# 1. Check database query performance
docker exec postgres psql -U saan -d saan_db -c "
  EXPLAIN ANALYZE 
  SELECT SUM(amount) FROM expense_transactions 
  WHERE cash_flow_id = 'cash_flow_123';"

# 2. Add missing indexes
docker exec postgres psql -U saan -d saan_db -c "
  CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_expenses_cash_flow_amount 
  ON expense_transactions(cash_flow_id, amount);"

# 3. Check for table locks
docker exec postgres psql -U saan -d saan_db -c "
  SELECT blocked_locks.pid AS blocked_pid,
         blocked_activity.usename AS blocked_user,
         blocking_locks.pid AS blocking_pid,
         blocking_activity.usename AS blocking_user,
         blocked_activity.query AS blocked_statement,
         blocking_activity.query AS current_statement_in_blocking_process
  FROM pg_catalog.pg_locks blocked_locks
  JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
  JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
  JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
  WHERE NOT blocked_locks.granted;"

# 4. Use simpler calculation method temporarily
curl -X POST http://finance:8088/api/v1/finance/profit-first/allocate \
  -d '{"business_date": "2024-01-15", "location_id": "location_123", "simple_mode": true}'
```

#### **Issue**: "High memory usage (>2GB RAM)"
```
Problem: Finance Service consuming excessive memory
Expected: <500MB
```

**Root Causes:**
- Memory leaks in calculations
- Large cache objects
- Goroutine leaks

**Solutions:**
```bash
# 1. Check memory usage
docker stats finance

# 2. Profile memory usage (if debug endpoints enabled)
curl http://finance:8088/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# 3. Check for goroutine leaks
curl http://finance:8088/debug/pprof/goroutine?debug=1

# 4. Clear large cache objects
redis-cli --scan --pattern "finance:*" | grep -E "(report|calculation)" | xargs redis-cli DEL

# 5. Restart service if needed
docker restart finance
```

---

## 🔧 **Configuration Issues**

#### **Issue**: "Service discovery failures"
```
Error: "no such host: loyverse"
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

# 2. Verify Loyverse service is running
docker ps | grep loyverse

# 3. Test DNS resolution
docker exec finance nslookup loyverse

# 4. Check service names in docker-compose.yml
grep "container_name" docker-compose.yml | grep loyverse

# 5. Test connectivity
docker exec finance curl http://loyverse:8091/health
```

#### **Issue**: "File storage upload failures"
```
Error: "S3 upload failed - access denied"
```

**Root Causes:**
- Invalid S3 credentials
- Bucket doesn't exist
- Permission issues

**Solutions:**
```bash
# 1. Check S3 credentials
echo $S3_ACCESS_KEY | wc -c
echo $S3_SECRET_KEY | wc -c

# 2. Test S3 connectivity
aws s3 ls s3://saan-finance-receipts/

# 3. Check bucket permissions
aws s3api get-bucket-policy --bucket saan-finance-receipts

# 4. Use local file storage temporarily
export FILE_STORAGE_TYPE=local
export LOCAL_STORAGE_PATH=/app/uploads

# 5. Create bucket if doesn't exist
aws s3 mb s3://saan-finance-receipts --region ap-southeast-1
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
docker restart postgres redis
# Wait 30 seconds
docker restart loyverse customer order payment
# Wait 30 seconds
docker restart finance

# 3. Verify health
curl http://finance:8088/health
```

### **Data Recovery**
```sql
-- Backup before recovery
pg_dump -U saan -h localhost -p 5432 saan_db > backup_$(date +%Y%m%d_%H%M%S).sql

-- Find incomplete cash flows
SELECT id, business_date, location_name, status, total_sales
FROM daily_cash_flows 
WHERE status = 'draft' 
AND business_date < CURRENT_DATE - INTERVAL '2 days';

-- Complete cash flows if all data is present
UPDATE daily_cash_flows 
SET status = 'completed'
WHERE id IN ('cash_flow_123', 'cash_flow_456')
AND total_sales > 0;

-- Find and fix orphaned expenses
SELECT et.id, et.description, et.amount, et.cash_flow_id
FROM expense_transactions et
LEFT JOIN daily_cash_flows cf ON et.cash_flow_id = cf.id
WHERE cf.id IS NULL;
```

### **Financial Data Validation**
```sql
-- Daily validation queries
-- 1. Check revenue consistency
SELECT 
  cf.business_date,
  cf.location_name,
  cf.total_sales,
  cf.loyverse_receipts_total + cf.delivery_orders_total + cf.other_income AS calculated_revenue,
  ABS(cf.total_sales - (cf.loyverse_receipts_total + cf.delivery_orders_total + cf.other_income)) AS variance
FROM daily_cash_flows cf
WHERE cf.business_date >= CURRENT_DATE - INTERVAL '7 days'
AND ABS(cf.total_sales - (cf.loyverse_receipts_total + cf.delivery_orders_total + cf.other_income)) > 0.01;

-- 2. Check expense allocation
SELECT 
  cf.business_date,
  cf.location_name,
  (cf.supplier_transfers + cf.operational_expenses + cf.fixed_expenses + cf.personal_expenses) AS cf_expenses,
  COALESCE(SUM(et.amount), 0) AS actual_expenses,
  ABS((cf.supplier_transfers + cf.operational_expenses + cf.fixed_expenses + cf.personal_expenses) - COALESCE(SUM(et.amount), 0)) AS variance
FROM daily_cash_flows cf
LEFT JOIN expense_transactions et ON cf.id = et.cash_flow_id
WHERE cf.business_date >= CURRENT_DATE - INTERVAL '7 days'
GROUP BY cf.id, cf.business_date, cf.location_name, cf.supplier_transfers, cf.operational_expenses, cf.fixed_expenses, cf.personal_expenses
HAVING ABS((cf.supplier_transfers + cf.operational_expenses + cf.fixed_expenses + cf.personal_expenses) - COALESCE(SUM(et.amount), 0)) > 0.01;

-- 3. Check profit allocation consistency
SELECT 
  pa.business_date,
  pa.location_id,
  pa.allocatable_revenue,
  pa.profit_amount + pa.owner_pay_amount + pa.tax_amount + pa.operating_amount AS total_allocated,
  ABS(pa.allocatable_revenue - (pa.profit_amount + pa.owner_pay_amount + pa.tax_amount + pa.operating_amount)) AS variance
FROM profit_allocations pa
WHERE pa.business_date >= CURRENT_DATE - INTERVAL '7 days'
AND ABS(pa.allocatable_revenue - (pa.profit_amount + pa.owner_pay_amount + pa.tax_amount + pa.operating_amount)) > 0.01;
```

---

## 📞 **Escalation Contacts**

### **Immediate Response Needed**
- **Financial Data Loss**: Senior Backend Engineer + Finance Manager
- **Payment Integration Issues**: DevOps team + Finance team + Payment provider
- **Complete Service Outage**: All hands + Management notification

### **Business Hours Response**
- **Expense Approval Issues**: Finance Manager + Branch Managers
- **Profit Allocation Failures**: Finance Manager + DevOps team
- **Reporting Issues**: Backend team + Finance team
- **Cache/Performance Issues**: DevOps team

---

## 📝 **Debugging Checklist**

### **Before Escalating**
- [ ] Check service health endpoints
- [ ] Verify service logs (last 100 lines)
- [ ] Test with curl commands
- [ ] Check Redis/Database connectivity
- [ ] Verify configuration variables
- [ ] Clear relevant cache keys
- [ ] Check external service status (Loyverse, S3)

### **Information to Gather**
- [ ] Exact error message and stack trace
- [ ] Steps to reproduce the issue
- [ ] Time when issue started
- [ ] Recent deployments or changes
- [ ] Business impact assessment (revenue loss, blocked operations)
- [ ] Current cash flow status for affected locations

---

## 📊 **Health Check Commands**

### **Quick Service Verification**
```bash
# Service health
curl http://finance:8088/health

# Database connectivity
curl http://finance:8088/api/v1/finance/admin/health/database

# Cache connectivity
curl http://finance:8088/api/v1/finance/admin/health/cache

# External services
curl http://finance:8088/api/v1/finance/admin/health/dependencies

# Recent cash flows
curl "http://finance:8088/api/v1/finance/cash-flows?limit=5&sort=created_at_desc"
```

### **Performance Monitoring**
```bash
# Service metrics
curl http://finance:8088/metrics

# Database performance
docker exec postgres psql -U saan -d saan_db -c "
  SELECT schemaname, tablename, n_tup_ins, n_tup_upd, n_tup_del 
  FROM pg_stat_user_tables 
  WHERE tablename LIKE '%cash_flow%' OR tablename LIKE '%expense%';"

# Cache hit rates
redis-cli INFO stats | grep keyspace
```

---

> 🚨 **When in doubt, restart services in dependency order: Database → Cache → External Services → Finance Service. Always backup financial data before major recovery operations.**
