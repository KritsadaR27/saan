# Redis Error Handling Improvements

## Overview

This document outlines the improvements made to Redis error handling in the Loyverse integration service. The enhancements provide robust error handling, connection monitoring, and graceful degradation capabilities.

## Key Improvements

### 1. Enhanced Redis Client (`internal/redis/client.go`)

**Features:**
- **Configurable Connection Parameters**: Timeouts, retry policies, and connection pooling
- **Automatic Retry Logic**: Exponential backoff with configurable limits
- **Connection Health Monitoring**: Real-time connection status tracking
- **Safe Operations**: Wrapper methods with enhanced error handling
- **Statistics Logging**: Connection pool and performance metrics

**Configuration Options:**
```go
type Config struct {
    Addr            string        // Redis server address
    MaxRetries      int           // Maximum retry attempts (default: 3)
    MinRetryBackoff time.Duration // Minimum retry delay (default: 100ms)
    MaxRetryBackoff time.Duration // Maximum retry delay (default: 3s)
    DialTimeout     time.Duration // Connection timeout (default: 5s)
    ReadTimeout     time.Duration // Read timeout (default: 3s)
    WriteTimeout    time.Duration // Write timeout (default: 3s)
    PoolSize        int           // Connection pool size (default: 10)
    PoolTimeout     time.Duration // Pool timeout (default: 4s)
    IdleTimeout     time.Duration // Idle connection timeout (default: 5m)
}
```

**Safe Operations:**
- `SafeSet()` - Set with retry logic
- `SafeGet()` - Get with retry logic  
- `SafeDel()` - Delete with retry logic
- `SafeExists()` - Exists check with retry logic

### 2. Health Monitoring (`internal/redis/monitor.go`)

**Features:**
- **Continuous Health Monitoring**: Periodic connection health checks
- **Health History Tracking**: Maintains history of recent health checks
- **Alert Thresholds**: Configurable failure thresholds before alerting
- **Health Change Callbacks**: Custom callbacks for health status changes
- **Statistics Collection**: Success rates and failure patterns

**Monitoring Capabilities:**
- Real-time health status
- Consecutive failure counting
- Health check history (last 10 checks)
- Configurable check intervals
- Alert callbacks for health changes

### 3. Enhanced Repository (`internal/repository/cache.go`)

**Features:**
- **Graceful Degradation**: Continues operation when Redis is unavailable
- **Enhanced Error Context**: Detailed error messages with operation context
- **Health Status Integration**: Uses health monitor for operation decisions
- **Comprehensive Statistics**: Health and performance metrics endpoint

**Error Handling Strategy:**
- **Non-Critical Operations**: Log errors but continue (e.g., caching)
- **Critical Operations**: Return errors for essential functionality
- **Health-Aware Operations**: Skip operations when Redis is unhealthy

### 4. Improved Sync Services

**Enhanced Product Sync (`internal/sync/product_sync.go`):**
- Health check before sync operations
- Enhanced error logging with context
- Graceful degradation when Redis is unavailable
- Improved cache operation error handling

**Enhanced Payment Type Sync (`internal/sync/payment_type_sync.go`):**
- Consistent error handling patterns
- Enhanced logging for debugging
- Safe Redis operations with retry logic

### 5. Enhanced Health Endpoints

**Primary Health Endpoint (`/health`):**
```json
{
  "status": "healthy|degraded",
  "service": "loyverse-integration",
  "timestamp": "2024-01-15T10:30:00Z",
  "redis": {
    "healthy": true,
    "stats": {
      "healthy": true,
      "consecutive_fails": 0,
      "last_health_check": "2024-01-15T10:29:45Z",
      "health_history": [true, true, true, true, true],
      "redis_pool_stats": {
        "hits": 150,
        "misses": 5,
        "timeouts": 0,
        "total_conns": 5,
        "idle_conns": 3,
        "stale_conns": 0
      }
    }
  }
}
```

**Redis-Specific Health Endpoint (`/health/redis`):**
Provides detailed Redis health and performance statistics.

## Error Handling Strategies

### 1. Retry Logic
- **Exponential Backoff**: Progressively longer delays between retries
- **Maximum Attempts**: Configurable retry limits (default: 3 attempts)
- **Context Cancellation**: Respects context timeouts and cancellations
- **Selective Retries**: Only retries transient errors, not permanent failures

### 2. Circuit Breaker Pattern
- **Health Monitoring**: Continuous connection health assessment
- **Automatic Recovery**: Detects when Redis becomes available again
- **Graceful Degradation**: Continues operation with reduced functionality

### 3. Comprehensive Logging
- **Error Context**: Detailed error messages with operation context
- **Performance Metrics**: Regular logging of connection statistics
- **Health Status**: Periodic health status reports
- **Alert Conditions**: Specific alerts for consecutive failures

## Usage Examples

### Basic Usage
```go
// Initialize enhanced Redis client
config := redis.DefaultConfig()
config.Addr = "localhost:6379"
client := redis.NewClient(config)

// Safe operations with automatic retry
err := client.SafeSet(ctx, "key", "value", time.Hour)
if err != nil {
    log.Printf("Set operation failed: %v", err)
}

value, err := client.SafeGet(ctx, "key")
if err != nil {
    log.Printf("Get operation failed: %v", err)
}
```

### Health Monitoring
```go
// Create health monitor
monitor := redis.NewHealthMonitor(client, 30*time.Second)

// Set up health change callback
monitor.SetHealthChangeCallback(func(healthy bool) {
    if healthy {
        log.Println("Redis connection restored")
    } else {
        log.Println("Redis connection lost")
    }
})

// Start monitoring
go monitor.Start(context.Background())
```

### Repository Usage
```go
// Initialize repository with enhanced error handling
config := redis.DefaultConfig()
repo := repository.NewRedisRepository(config)

// Operations automatically handle Redis health
err := repo.Set(ctx, "product:123", productData, time.Hour)
// Will log error but continue if Redis is unavailable

// Check health status
if repo.IsHealthy() {
    log.Println("Redis is healthy")
}

// Get health statistics
stats := repo.GetHealthStats()
log.Printf("Health stats: %+v", stats)
```

## Configuration Recommendations

### Development Environment
```go
config := redis.Config{
    Addr:            "localhost:6379",
    MaxRetries:      2,
    MinRetryBackoff: 50 * time.Millisecond,
    MaxRetryBackoff: 1 * time.Second,
    DialTimeout:     3 * time.Second,
    ReadTimeout:     2 * time.Second,
    WriteTimeout:    2 * time.Second,
    PoolSize:        5,
}
```

### Production Environment
```go
config := redis.Config{
    Addr:            "redis-cluster:6379",
    MaxRetries:      5,
    MinRetryBackoff: 100 * time.Millisecond,
    MaxRetryBackoff: 5 * time.Second,
    DialTimeout:     10 * time.Second,
    ReadTimeout:     5 * time.Second,
    WriteTimeout:    5 * time.Second,
    PoolSize:        20,
    PoolTimeout:     6 * time.Second,
}
```

## Monitoring and Alerting

### Key Metrics to Monitor
1. **Connection Health**: Overall Redis connectivity status
2. **Success Rate**: Percentage of successful operations
3. **Consecutive Failures**: Number of consecutive failed health checks
4. **Pool Statistics**: Connection pool utilization and performance
5. **Operation Latency**: Time taken for Redis operations

### Alert Conditions
1. **Redis Unavailable**: 3+ consecutive health check failures
2. **High Error Rate**: >10% operation failure rate
3. **Pool Exhaustion**: All connections in use for extended periods
4. **High Latency**: Operations taking >1 second consistently

## Benefits

1. **Improved Reliability**: Service continues operating even when Redis is unavailable
2. **Better Observability**: Comprehensive health and performance monitoring
3. **Faster Recovery**: Automatic detection and recovery from Redis issues
4. **Enhanced Debugging**: Detailed error logging with context
5. **Production Ready**: Robust error handling suitable for production environments

## Migration Notes

The enhanced Redis client is backward compatible with existing code. Key changes:
- Replace direct Redis operations with `Safe*` methods for enhanced error handling
- Use the new health monitoring for operational awareness
- Update health check endpoints to include Redis status
- Consider graceful degradation strategies in your application logic

## Testing

To test the enhanced error handling:

1. **Connection Failure**: Stop Redis server and observe graceful degradation
2. **Network Issues**: Introduce network latency to test retry logic
3. **Health Recovery**: Restart Redis and verify automatic recovery
4. **Load Testing**: Test connection pool behavior under high load

The system now provides robust Redis error handling that ensures the Loyverse integration service remains operational even during Redis connectivity issues.
