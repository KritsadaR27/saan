# Shipping Service Integration Guide

## Overview
The Shipping Service provides comprehensive delivery management, vehicle fleet operations, and third-party provider integrations for the Saan System. It supports both self-delivery operations and integration with external delivery providers.

## Architecture Integration

### Service Dependencies
```yaml
Upstream Dependencies:
  - Order Service: Delivery creation triggers
  - Customer Service: Address validation
  - Finance Service: COD and fee processing
  - Product Service: Package dimensions/weight

Downstream Dependencies:
  - Notification Service: Delivery status updates
  - Analytics Service: Performance metrics
  - External APIs: Third-party delivery providers
```

### Event-Driven Integration

#### Published Events
```typescript
// Delivery lifecycle events
interface DeliveryCreatedEvent {
  event_type: "delivery.created";
  delivery_id: string;
  order_id: string;
  customer_id: string;
  delivery_method: string;
  estimated_delivery_time: string;
  timestamp: string;
}

interface DeliveryStatusUpdatedEvent {
  event_type: "delivery.status_updated";
  delivery_id: string;
  old_status: string;
  new_status: string;
  location?: string;
  timestamp: string;
}

interface DeliveryCompletedEvent {
  event_type: "delivery.completed";
  delivery_id: string;
  order_id: string;
  actual_delivery_time: string;
  delivery_fee: number;
  cod_collected?: number;
  customer_rating?: number;
  timestamp: string;
}

// Route optimization events
interface RouteOptimizedEvent {
  event_type: "route.optimized";
  route_id: string;
  vehicle_id: string;
  delivery_count: number;
  total_distance: number;
  estimated_duration: number;
  optimization_savings: {
    distance_saved_km: number;
    time_saved_minutes: number;
    fuel_cost_saved: number;
  };
  timestamp: string;
}

// Vehicle tracking events
interface VehicleLocationUpdatedEvent {
  event_type: "vehicle.location_updated";
  vehicle_id: string;
  driver_id: string;
  location: {
    latitude: number;
    longitude: number;
    accuracy: number;
  };
  speed?: number;
  bearing?: number;
  timestamp: string;
}
```

#### Consumed Events
```typescript
// From Order Service
interface OrderConfirmedEvent {
  event_type: "order.confirmed";
  order_id: string;
  customer_id: string;
  customer_address_id: string;
  delivery_requirements: {
    method_preference?: string;
    delivery_instructions?: string;
    time_preference?: string;
  };
  package_info: {
    weight: number;
    dimensions: object;
    value: number;
    fragile: boolean;
  };
  timestamp: string;
}

interface OrderCancelledEvent {
  event_type: "order.cancelled";
  order_id: string;
  cancellation_reason: string;
  timestamp: string;
}

// From Customer Service
interface CustomerAddressUpdatedEvent {
  event_type: "customer.address_updated";
  customer_id: string;
  address_id: string;
  new_address: object;
  timestamp: string;
}

// From Payment Service
interface PaymentCompletedEvent {
  event_type: "payment.completed";
  order_id: string;
  payment_method: string;
  amount: number;
  timestamp: string;
}
```

## Database Integration

### Primary Database Schema
```sql
-- Core delivery tables
CREATE TABLE delivery_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id),
    customer_id UUID NOT NULL,
    customer_address_id UUID NOT NULL,
    delivery_method VARCHAR(50) NOT NULL,
    provider_id UUID REFERENCES delivery_providers(id),
    vehicle_id UUID REFERENCES delivery_vehicles(id),
    route_id UUID REFERENCES delivery_routes(id),
    tracking_number VARCHAR(100) UNIQUE,
    provider_order_id VARCHAR(100),
    scheduled_pickup_time TIMESTAMP,
    planned_delivery_date DATE NOT NULL,
    estimated_delivery_time TIMESTAMP,
    actual_pickup_time TIMESTAMP,
    actual_delivery_time TIMESTAMP,
    delivery_fee DECIMAL(10,2) NOT NULL DEFAULT 0,
    cod_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    notes TEXT,
    delivery_instructions TEXT,
    requires_manual_coordination BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Vehicle fleet management
CREATE TABLE delivery_vehicles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    license_plate VARCHAR(20) NOT NULL UNIQUE,
    vehicle_type VARCHAR(50) NOT NULL,
    brand VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    year INTEGER NOT NULL,
    max_weight DECIMAL(8,2) NOT NULL,
    max_volume DECIMAL(8,2) NOT NULL,
    fuel_type VARCHAR(50) NOT NULL,
    driver_id UUID,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    current_location JSONB,
    last_maintenance DATE,
    next_maintenance DATE,
    notes TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Route optimization
CREATE TABLE delivery_routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    route_name VARCHAR(200) NOT NULL,
    route_date DATE NOT NULL,
    assigned_vehicle_id UUID REFERENCES delivery_vehicles(id),
    assigned_driver_id UUID,
    planned_start_time TIMESTAMP,
    planned_end_time TIMESTAMP,
    total_planned_distance DECIMAL(8,2) DEFAULT 0,
    total_planned_orders INTEGER DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'planned',
    actual_start_time TIMESTAMP,
    actual_end_time TIMESTAMP,
    actual_distance DECIMAL(8,2) DEFAULT 0,
    actual_orders_delivered INTEGER DEFAULT 0,
    route_optimization_data JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Provider integrations
CREATE TABLE delivery_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_code VARCHAR(50) NOT NULL UNIQUE,
    provider_name VARCHAR(200) NOT NULL,
    provider_type VARCHAR(50) NOT NULL,
    api_base_url VARCHAR(500),
    api_version VARCHAR(20),
    has_api BOOLEAN NOT NULL DEFAULT false,
    auth_method VARCHAR(100),
    coverage_areas JSONB,
    supported_package_types JSONB,
    max_weight_kg DECIMAL(8,2),
    max_dimensions JSONB,
    base_rate DECIMAL(10,2) NOT NULL DEFAULT 0,
    per_km_rate DECIMAL(10,2) NOT NULL DEFAULT 0,
    weight_surcharge_rate DECIMAL(10,2) NOT NULL DEFAULT 0,
    same_day_surcharge DECIMAL(10,2) NOT NULL DEFAULT 0,
    cod_surcharge_rate DECIMAL(10,4) NOT NULL DEFAULT 0,
    standard_delivery_hours INTEGER DEFAULT 24,
    express_delivery_hours INTEGER DEFAULT 4,
    same_day_available BOOLEAN NOT NULL DEFAULT false,
    cod_available BOOLEAN NOT NULL DEFAULT false,
    tracking_available BOOLEAN NOT NULL DEFAULT false,
    insurance_available BOOLEAN NOT NULL DEFAULT false,
    daily_cutoff_time TIME,
    weekend_service BOOLEAN NOT NULL DEFAULT false,
    holiday_service BOOLEAN NOT NULL DEFAULT false,
    business_hours JSONB,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### Cross-Service Data Synchronization
```sql
-- Read-only views for other services
CREATE VIEW delivery_summary FOR analytics AS
SELECT 
    d.id,
    d.order_id,
    d.delivery_method,
    d.status,
    d.delivery_fee,
    d.cod_amount,
    d.planned_delivery_date,
    d.actual_delivery_time,
    p.provider_name,
    v.vehicle_type,
    EXTRACT(epoch FROM (d.actual_delivery_time - d.created_at))/3600 as delivery_duration_hours
FROM delivery_orders d
LEFT JOIN delivery_providers p ON d.provider_id = p.id
LEFT JOIN delivery_vehicles v ON d.vehicle_id = v.id
WHERE d.is_active = true;

-- Materialized view for performance metrics
CREATE MATERIALIZED VIEW delivery_performance_metrics AS
SELECT 
    DATE(planned_delivery_date) as delivery_date,
    delivery_method,
    COUNT(*) as total_deliveries,
    COUNT(CASE WHEN status = 'delivered' THEN 1 END) as successful_deliveries,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_deliveries,
    AVG(delivery_fee) as avg_delivery_fee,
    AVG(EXTRACT(epoch FROM (actual_delivery_time - created_at))/3600) as avg_delivery_duration_hours
FROM delivery_orders
WHERE is_active = true
GROUP BY DATE(planned_delivery_date), delivery_method;
```

## External API Integrations

### Grab Delivery API
```typescript
interface GrabDeliveryIntegration {
  // Order creation
  createDeliveryOrder(request: {
    origin: GrabLocation;
    destination: GrabLocation;
    package_detail: GrabPackageDetail;
    service_type: 'INSTANT' | 'SAME_DAY' | 'NEXT_DAY';
    cod_amount?: number;
  }): Promise<GrabDeliveryResponse>;

  // Order tracking
  getDeliveryStatus(delivery_id: string): Promise<GrabTrackingResponse>;

  // Webhook handlers
  handleStatusUpdate(webhook: GrabWebhook): void;
  handleDeliveryCompleted(webhook: GrabWebhook): void;
}

// Configuration
const grabConfig = {
  api_base_url: "https://partner-api.grab.com/grabexpress/v1",
  client_id: process.env.GRAB_CLIENT_ID,
  client_secret: process.env.GRAB_CLIENT_SECRET,
  webhook_secret: process.env.GRAB_WEBHOOK_SECRET,
  service_types: ['INSTANT', 'SAME_DAY'],
  max_weight_kg: 20,
  max_cod_amount: 10000
};
```

### LINE MAN API
```typescript
interface LineManDeliveryIntegration {
  // Service availability
  checkServiceArea(location: LineManLocation): Promise<LineManCoverageResponse>;
  
  // Price calculation
  calculatePrice(request: {
    pickup_location: LineManLocation;
    delivery_location: LineManLocation;
    package_size: 'S' | 'M' | 'L' | 'XL';
    is_express?: boolean;
  }): Promise<LineManPriceResponse>;

  // Order management
  createOrder(request: LineManOrderRequest): Promise<LineManOrderResponse>;
  cancelOrder(order_id: string, reason: string): Promise<void>;
  
  // Real-time tracking
  getOrderStatus(order_id: string): Promise<LineManStatusResponse>;
  subscribeToUpdates(order_id: string, callback: Function): void;
}
```

### Lalamove API
```typescript
interface LalamoveIntegration {
  // Quotation
  getQuotation(request: {
    service_type: 'MOTORCYCLE' | 'CAR' | 'VAN';
    stops: LalamoveStop[];
    is_routed?: boolean;
    requirements?: LalamoveRequirements;
  }): Promise<LalamoveQuotationResponse>;

  // Order placement
  placeOrder(quotation_id: string, sender: LalamoveSender): Promise<LalamoveOrderResponse>;
  
  // Order management
  cancelOrder(order_id: string): Promise<void>;
  getOrderDetails(order_id: string): Promise<LalamoveOrderDetails>;
  
  // Driver tracking
  trackDriver(order_id: string): Promise<LalamoveDriverLocation>;
}
```

## Cache Integration

### Redis Caching Strategy
```typescript
interface ShippingCacheKeys {
  // Provider rate calculations (TTL: 1 hour)
  provider_rates: `shipping:rates:${provider_id}:${location_hash}`;
  
  // Coverage area checks (TTL: 24 hours)
  coverage_check: `shipping:coverage:${method}:${location_hash}`;
  
  // Route optimizations (TTL: 1 hour)
  route_optimization: `shipping:route:${route_id}:optimization`;
  
  // Vehicle locations (TTL: 5 minutes)
  vehicle_location: `shipping:vehicle:${vehicle_id}:location`;
  
  // Delivery tracking (TTL: 10 minutes)
  delivery_tracking: `shipping:delivery:${delivery_id}:tracking`;
}

// Cache implementation
class ShippingCache {
  async cacheProviderRate(
    provider_id: string, 
    location_hash: string, 
    rate_data: ProviderRateData
  ): Promise<void> {
    const key = `shipping:rates:${provider_id}:${location_hash}`;
    await redis.setex(key, 3600, JSON.stringify(rate_data));
  }

  async getCachedProviderRate(
    provider_id: string, 
    location_hash: string
  ): Promise<ProviderRateData | null> {
    const key = `shipping:rates:${provider_id}:${location_hash}`;
    const cached = await redis.get(key);
    return cached ? JSON.parse(cached) : null;
  }

  async invalidateDeliveryCache(delivery_id: string): Promise<void> {
    const pattern = `shipping:delivery:${delivery_id}:*`;
    const keys = await redis.keys(pattern);
    if (keys.length > 0) {
      await redis.del(...keys);
    }
  }
}
```

## Message Queue Integration

### Job Queue Processing
```typescript
// Delivery creation workflow
interface DeliveryCreationJob {
  job_type: 'create_delivery';
  data: {
    order_id: string;
    delivery_request: CreateDeliveryRequest;
  };
}

// Route optimization jobs
interface RouteOptimizationJob {
  job_type: 'optimize_route';
  data: {
    route_id: string;
    optimization_params: RouteOptimizationParams;
  };
}

// Provider sync jobs
interface ProviderSyncJob {
  job_type: 'sync_provider_status';
  data: {
    provider_id: string;
    delivery_ids: string[];
  };
}

// Job processors
class ShippingJobProcessor {
  async processDeliveryCreation(job: DeliveryCreationJob): Promise<void> {
    const { order_id, delivery_request } = job.data;
    
    // 1. Validate order exists and is confirmed
    // 2. Check delivery coverage
    // 3. Calculate delivery rates
    // 4. Create delivery order
    // 5. Integrate with provider API
    // 6. Publish delivery created event
  }

  async processRouteOptimization(job: RouteOptimizationJob): Promise<void> {
    const { route_id, optimization_params } = job.data;
    
    // 1. Fetch route and delivery orders
    // 2. Run optimization algorithm
    // 3. Update route with optimized sequence
    // 4. Notify affected deliveries
    // 5. Publish route optimized event
  }

  async processProviderSync(job: ProviderSyncJob): Promise<void> {
    const { provider_id, delivery_ids } = job.data;
    
    // 1. Fetch delivery statuses from provider API
    // 2. Update local delivery records
    // 3. Publish status update events
    // 4. Handle delivery completion logic
  }
}
```

## Integration Patterns

### Service-to-Service Communication
```typescript
// Order Service integration
class OrderServiceIntegration {
  async notifyDeliveryCreated(delivery: DeliveryOrder): Promise<void> {
    await this.orderServiceClient.post('/internal/deliveries', {
      order_id: delivery.order_id,
      delivery_id: delivery.id,
      tracking_number: delivery.tracking_number,
      estimated_delivery_time: delivery.estimated_delivery_time
    });
  }

  async notifyDeliveryCompleted(delivery: DeliveryOrder): Promise<void> {
    await this.orderServiceClient.patch(`/internal/orders/${delivery.order_id}/delivery-status`, {
      status: 'delivered',
      actual_delivery_time: delivery.actual_delivery_time,
      delivery_fee: delivery.delivery_fee
    });
  }
}

// Customer Service integration
class CustomerServiceIntegration {
  async validateCustomerAddress(customer_id: string, address_id: string): Promise<CustomerAddress> {
    const response = await this.customerServiceClient.get(
      `/internal/customers/${customer_id}/addresses/${address_id}`
    );
    return response.data.address;
  }

  async getCustomerDeliveryPreferences(customer_id: string): Promise<DeliveryPreferences> {
    const response = await this.customerServiceClient.get(
      `/internal/customers/${customer_id}/delivery-preferences`
    );
    return response.data.preferences;
  }
}

// Finance Service integration
class FinanceServiceIntegration {
  async recordDeliveryRevenue(delivery: DeliveryOrder): Promise<void> {
    await this.financeServiceClient.post('/internal/cash-flow', {
      entity_type: 'delivery',
      entity_id: delivery.id,
      transaction_type: 'inflow',
      amount: delivery.delivery_fee,
      description: `Delivery fee for order ${delivery.order_id}`,
      category: 'delivery_revenue'
    });
  }

  async recordCODCollection(delivery: DeliveryOrder): Promise<void> {
    if (delivery.cod_amount > 0) {
      await this.financeServiceClient.post('/internal/cash-flow', {
        entity_type: 'delivery',
        entity_id: delivery.id,
        transaction_type: 'inflow',
        amount: delivery.cod_amount,
        description: `COD collection for order ${delivery.order_id}`,
        category: 'cod_collection'
      });
    }
  }
}
```

### Error Handling & Resilience
```typescript
// Circuit breaker for external APIs
class ProviderCircuitBreaker {
  private failures: Map<string, number> = new Map();
  private lastFailureTime: Map<string, number> = new Map();
  private readonly maxFailures = 5;
  private readonly timeoutMs = 30000;

  async callProvider<T>(provider_id: string, operation: () => Promise<T>): Promise<T> {
    if (this.isCircuitOpen(provider_id)) {
      throw new Error(`Circuit breaker open for provider ${provider_id}`);
    }

    try {
      const result = await operation();
      this.onSuccess(provider_id);
      return result;
    } catch (error) {
      this.onFailure(provider_id);
      throw error;
    }
  }

  private isCircuitOpen(provider_id: string): boolean {
    const failures = this.failures.get(provider_id) || 0;
    const lastFailure = this.lastFailureTime.get(provider_id) || 0;
    
    if (failures >= this.maxFailures) {
      return (Date.now() - lastFailure) < this.timeoutMs;
    }
    
    return false;
  }
}

// Retry mechanism for critical operations
class DeliveryRetryHandler {
  async withRetry<T>(
    operation: () => Promise<T>,
    maxRetries: number = 3,
    backoffMs: number = 1000
  ): Promise<T> {
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        return await operation();
      } catch (error) {
        if (attempt === maxRetries) {
          throw error;
        }
        
        await this.delay(backoffMs * Math.pow(2, attempt - 1));
      }
    }
    
    throw new Error('Max retries exceeded');
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}
```

## Monitoring & Observability

### Metrics Collection
```typescript
interface ShippingMetrics {
  // Delivery performance
  delivery_success_rate: number;
  average_delivery_time_hours: number;
  delivery_cost_per_order: number;
  
  // Provider performance
  provider_api_response_time_ms: Record<string, number>;
  provider_success_rate: Record<string, number>;
  provider_cost_comparison: Record<string, number>;
  
  // Fleet utilization
  vehicle_utilization_rate: number;
  average_deliveries_per_route: number;
  fuel_efficiency_km_per_liter: number;
  
  // Route optimization
  route_optimization_savings_percent: number;
  distance_optimization_km_saved: number;
  time_optimization_minutes_saved: number;
}

// Metrics publishing
class ShippingMetricsPublisher {
  async publishDeliveryMetrics(delivery: DeliveryOrder): Promise<void> {
    const metrics = {
      delivery_completed: 1,
      delivery_duration_hours: this.calculateDeliveryDuration(delivery),
      delivery_cost: delivery.delivery_fee,
      delivery_method: delivery.delivery_method,
      success: delivery.status === 'delivered' ? 1 : 0
    };
    
    await this.metricsClient.publish('shipping.delivery', metrics);
  }

  async publishVehicleMetrics(vehicle: DeliveryVehicle, route: DeliveryRoute): Promise<void> {
    const metrics = {
      vehicle_distance_km: route.actual_distance,
      vehicle_orders_delivered: route.actual_orders_delivered,
      vehicle_utilization_hours: this.calculateUtilizationHours(route),
      fuel_consumption_estimate: this.estimateFuelConsumption(vehicle, route)
    };
    
    await this.metricsClient.publish('shipping.vehicle', metrics);
  }
}
```

## Security Integration

### Authentication & Authorization
```typescript
// JWT token validation for internal service calls
class ShippingAuthMiddleware {
  async validateServiceToken(token: string): Promise<ServiceClaims> {
    const decoded = jwt.verify(token, process.env.SERVICE_JWT_SECRET);
    return decoded as ServiceClaims;
  }

  async validateUserPermissions(user_id: string, action: string, resource: string): Promise<boolean> {
    const response = await this.authServiceClient.post('/internal/check-permission', {
      user_id,
      action,
      resource
    });
    
    return response.data.allowed;
  }
}

// Data encryption for sensitive information
class ShippingDataEncryption {
  encryptTrackingData(data: TrackingData): string {
    return encrypt(JSON.stringify(data), process.env.TRACKING_ENCRYPTION_KEY);
  }

  decryptTrackingData(encrypted: string): TrackingData {
    const decrypted = decrypt(encrypted, process.env.TRACKING_ENCRYPTION_KEY);
    return JSON.parse(decrypted);
  }
}
```

## Configuration Management

### Environment Configuration
```bash
# Database
SHIPPING_DB_HOST=localhost
SHIPPING_DB_PORT=5432
SHIPPING_DB_NAME=saan_shipping
SHIPPING_DB_USER=shipping_user
SHIPPING_DB_PASSWORD=shipping_pass

# Redis Cache
SHIPPING_REDIS_HOST=localhost
SHIPPING_REDIS_PORT=6379
SHIPPING_REDIS_PASSWORD=redis_pass
SHIPPING_REDIS_DB=3

# External APIs
GRAB_CLIENT_ID=grab_client_id
GRAB_CLIENT_SECRET=grab_client_secret
GRAB_WEBHOOK_SECRET=grab_webhook_secret

LINEMAN_API_KEY=lineman_api_key
LINEMAN_MERCHANT_ID=lineman_merchant_id

LALAMOVE_API_KEY=lalamove_api_key
LALAMOVE_MERCHANT_SECRET=lalamove_secret

# Message Queue
RABBITMQ_URL=amqp://localhost:5672
SHIPPING_QUEUE_NAME=shipping_queue

# Service Discovery
ORDER_SERVICE_URL=http://localhost:8083
CUSTOMER_SERVICE_URL=http://localhost:8082
FINANCE_SERVICE_URL=http://localhost:8085
NOTIFICATION_SERVICE_URL=http://localhost:8087

# Security
SERVICE_JWT_SECRET=service_jwt_secret
TRACKING_ENCRYPTION_KEY=tracking_encryption_key
```

This integration guide provides a comprehensive overview of how the Shipping Service integrates with other services in the Saan System, external APIs, databases, caching, messaging, and monitoring systems.
