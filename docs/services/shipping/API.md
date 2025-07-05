# Shipping Service API Documentation

## Overview
The Shipping Service manages delivery operations, route optimization, vehicle fleet management, and third-party delivery provider integrations for the Saan System.

## Base URL
```
http://localhost:8086/api/v1
```

## Authentication
All API endpoints require proper authentication headers (implementation pending).

## Core Entities

### Delivery Order
Represents a delivery order with tracking and status management.

```json
{
  "id": "uuid",
  "order_id": "uuid",
  "customer_id": "uuid", 
  "customer_address_id": "uuid",
  "delivery_method": "self_delivery|grab|lineman|lalamove|inter_express|nim_express|rot_rao",
  "provider_id": "uuid",
  "vehicle_id": "uuid",
  "route_id": "uuid",
  "tracking_number": "string",
  "provider_order_id": "string",
  "scheduled_pickup_time": "2024-01-01T10:00:00Z",
  "planned_delivery_date": "2024-01-01T15:00:00Z",
  "estimated_delivery_time": "2024-01-01T15:30:00Z",
  "actual_pickup_time": "2024-01-01T10:15:00Z",
  "actual_delivery_time": "2024-01-01T15:45:00Z",
  "delivery_fee": 50.00,
  "cod_amount": 500.00,
  "status": "pending|planned|dispatched|in_transit|delivered|failed|cancelled",
  "notes": "Handle with care",
  "delivery_instructions": "Leave at front door",
  "requires_manual_coordination": false,
  "is_active": true,
  "created_at": "2024-01-01T09:00:00Z",
  "updated_at": "2024-01-01T15:45:00Z"
}
```

### Delivery Vehicle
Represents vehicles used for self-delivery operations.

```json
{
  "id": "uuid",
  "license_plate": "ABC-123",
  "vehicle_type": "motorcycle|car|truck|van",
  "brand": "Honda",
  "model": "Wave 110i",
  "year": 2023,
  "max_weight": 100.0,
  "max_volume": 0.5,
  "fuel_type": "gasoline",
  "driver_id": "uuid",
  "status": "active|inactive|maintenance|on_route",
  "current_location": "{\"lat\": 13.7563, \"lng\": 100.5018}",
  "last_maintenance": "2024-01-01T00:00:00Z",
  "next_maintenance": "2024-04-01T00:00:00Z",
  "notes": "Regular maintenance required",
  "is_active": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Delivery Provider
Configuration for third-party delivery services.

```json
{
  "id": "uuid",
  "provider_code": "GRAB",
  "provider_name": "Grab Delivery",
  "provider_type": "api|manual|hybrid",
  "api_base_url": "https://api.grab.com/v1",
  "api_version": "v1",
  "has_api": true,
  "auth_method": "bearer_token",
  "coverage_areas": {},
  "supported_package_types": {},
  "max_weight_kg": "50.00",
  "max_dimensions": {},
  "base_rate": "25.00",
  "per_km_rate": "5.00", 
  "weight_surcharge_rate": "2.00",
  "same_day_surcharge": "20.00",
  "cod_surcharge_rate": "0.02",
  "standard_delivery_hours": 24,
  "express_delivery_hours": 4,
  "same_day_available": true,
  "cod_available": true,
  "tracking_available": true,
  "insurance_available": false,
  "daily_cutoff_time": "16:00:00",
  "weekend_service": true,
  "holiday_service": false,
  "business_hours": {},
  "is_active": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Delivery Route
Optimized routes for vehicle delivery planning.

```json
{
  "id": "uuid",
  "route_name": "Route A - Central Bangkok",
  "route_date": "2024-01-01",
  "assigned_vehicle_id": "uuid",
  "assigned_driver_id": "uuid",
  "planned_start_time": "2024-01-01T08:00:00Z",
  "planned_end_time": "2024-01-01T17:00:00Z",
  "total_planned_distance_km": 85.5,
  "total_planned_orders": 15,
  "status": "planned|in_progress|completed|cancelled",
  "actual_start_time": "2024-01-01T08:15:00Z",
  "actual_end_time": "2024-01-01T17:30:00Z",
  "actual_distance_km": 92.3,
  "actual_orders_delivered": 14,
  "route_optimization_data": {},
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T17:30:00Z"
}
```

## API Endpoints

### Delivery Management

#### Create Delivery
```http
POST /deliveries
Content-Type: application/json

{
  "order_id": "uuid",
  "customer_id": "uuid",
  "customer_address_id": "uuid", 
  "delivery_method": "grab",
  "delivery_fee": 50.00,
  "cod_amount": 500.00,
  "delivery_instructions": "Leave at front door",
  "scheduled_pickup_time": "2024-01-01T10:00:00Z",
  "planned_delivery_date": "2024-01-01T15:00:00Z"
}
```

**Response (201):**
```json
{
  "status": "success",
  "data": {
    "delivery": { /* Delivery Object */ },
    "tracking_number": "TRK123456789",
    "estimated_delivery_time": "2024-01-01T15:30:00Z"
  }
}
```

#### Get Delivery
```http
GET /deliveries/{id}
```

**Response (200):**
```json
{
  "status": "success", 
  "data": {
    "delivery": { /* Delivery Object */ }
  }
}
```

#### Update Delivery Status
```http
PATCH /deliveries/{id}/status
Content-Type: application/json

{
  "status": "dispatched",
  "notes": "Driver en route to pickup location"
}
```

**Response (200):**
```json
{
  "status": "success",
  "data": {
    "delivery": { /* Updated Delivery Object */ }
  }
}
```

#### Track Delivery
```http
GET /deliveries/{id}/tracking
```

**Response (200):**
```json
{
  "status": "success",
  "data": {
    "tracking_info": {
      "current_status": "in_transit",
      "last_update": "2024-01-01T14:30:00Z",
      "location": "Near Central Plaza",
      "estimated_arrival": "2024-01-01T15:30:00Z",
      "tracking_history": [
        {
          "status": "dispatched",
          "timestamp": "2024-01-01T10:15:00Z",
          "location": "Distribution Center"
        }
      ]
    }
  }
}
```

#### Search Deliveries
```http
GET /deliveries/search?status=in_transit&date_from=2024-01-01&date_to=2024-01-31&vehicle_id=uuid
```

**Response (200):**
```json
{
  "status": "success",
  "data": {
    "deliveries": [ /* Array of Delivery Objects */ ],
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total": 150,
      "total_pages": 8
    }
  }
}
```

### Vehicle Management

#### Get Vehicles
```http
GET /vehicles?status=active&type=motorcycle
```

**Response (200):**
```json
{
  "status": "success",
  "data": {
    "vehicles": [ /* Array of Vehicle Objects */ ]
  }
}
```

#### Create Vehicle
```http
POST /vehicles
Content-Type: application/json

{
  "license_plate": "ABC-123",
  "vehicle_type": "motorcycle",
  "brand": "Honda",
  "model": "Wave 110i",
  "year": 2023,
  "max_weight": 100.0,
  "max_volume": 0.5,
  "fuel_type": "gasoline"
}
```

**Response (201):**
```json
{
  "status": "success",
  "data": {
    "vehicle": { /* Vehicle Object */ }
  }
}
```

#### Update Vehicle Location
```http
PATCH /vehicles/{id}/location
Content-Type: application/json

{
  "latitude": 13.7563,
  "longitude": 100.5018,
  "timestamp": "2024-01-01T14:30:00Z"
}
```

### Route Management

#### Create Route
```http
POST /routes
Content-Type: application/json

{
  "route_name": "Route A - Central Bangkok",
  "route_date": "2024-01-01",
  "assigned_vehicle_id": "uuid",
  "assigned_driver_id": "uuid", 
  "delivery_orders": ["uuid1", "uuid2", "uuid3"]
}
```

**Response (201):**
```json
{
  "status": "success",
  "data": {
    "route": { /* Route Object */ },
    "optimization_summary": {
      "total_distance": 85.5,
      "estimated_duration_hours": 8.5,
      "fuel_cost_estimate": 280.50
    }
  }
}
```

#### Optimize Route
```http
POST /routes/{id}/optimize
Content-Type: application/json

{
  "algorithm": "genetic_algorithm",
  "constraints": {
    "max_duration_hours": 10,
    "vehicle_capacity": 100.0,
    "time_windows": true
  }
}
```

### Provider Management

#### Get Providers
```http
GET /providers?active=true&has_api=true
```

**Response (200):**
```json
{
  "status": "success",
  "data": {
    "providers": [ /* Array of Provider Objects */ ]
  }
}
```

#### Calculate Delivery Rate
```http
POST /providers/{id}/calculate-rate
Content-Type: application/json

{
  "pickup_address": {
    "lat": 13.7563,
    "lng": 100.5018
  },
  "delivery_address": {
    "lat": 13.7278,
    "lng": 100.5250
  },
  "package_weight": 2.5,
  "package_dimensions": {
    "length": 30,
    "width": 20, 
    "height": 15
  },
  "delivery_type": "standard",
  "cod_amount": 500.00
}
```

**Response (200):**
```json
{
  "status": "success",
  "data": {
    "rate_calculation": {
      "base_rate": 25.00,
      "distance_rate": 45.00,
      "weight_surcharge": 5.00,
      "cod_surcharge": 10.00,
      "total_rate": 85.00,
      "estimated_delivery_time": "2024-01-01T15:30:00Z"
    }
  }
}
```

### Coverage Area Management

#### Check Coverage
```http
POST /coverage/check
Content-Type: application/json

{
  "address": {
    "lat": 13.7563,
    "lng": 100.5018
  },
  "delivery_method": "grab"
}
```

**Response (200):**
```json
{
  "status": "success",
  "data": {
    "covered": true,
    "provider": "Grab Delivery",
    "estimated_delivery_time": "2024-01-01T15:30:00Z",
    "service_level": "standard"
  }
}
```

### Health Checks

#### Health Check
```http
GET /health
```

**Response (200):**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "version": "1.0.0"
}
```

#### Readiness Check
```http
GET /ready
```

**Response (200):**
```json
{
  "status": "ready",
  "dependencies": {
    "database": "healthy",
    "redis": "healthy",
    "external_apis": "healthy"
  }
}
```

## Status Codes

- **200 OK**: Successful operation
- **201 Created**: Resource created successfully
- **400 Bad Request**: Invalid request data
- **401 Unauthorized**: Authentication required
- **403 Forbidden**: Insufficient permissions
- **404 Not Found**: Resource not found
- **409 Conflict**: Resource conflict (e.g., duplicate tracking number)
- **422 Unprocessable Entity**: Validation errors
- **500 Internal Server Error**: Server error

## Error Response Format
```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid delivery method",
    "details": {
      "field": "delivery_method",
      "rejected_value": "invalid_method"
    }
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Delivery Methods

### Self Delivery
- **Code**: `self_delivery`
- **Description**: In-house delivery using company vehicles
- **Features**: Real-time tracking, route optimization, driver management

### Third-Party Providers
- **Grab**: `grab` - API integration for food/package delivery
- **LINE MAN**: `lineman` - API integration for various delivery services
- **Lalamove**: `lalamove` - On-demand logistics platform
- **Inter Express**: `inter_express` - Manual coordination required
- **Nim Express**: `nim_express` - Manual coordination required  
- **Rot Rao**: `rot_rao` - Manual coordination required

## Business Rules

### Delivery Creation
- Order must exist and be confirmed before creating delivery
- Customer address must be validated and within coverage area
- Delivery method must be available for the target location
- COD amount cannot exceed configured limits

### Status Transitions
- **Pending** → **Planned**: Route assignment completed
- **Planned** → **Dispatched**: Driver begins pickup
- **Dispatched** → **In Transit**: Package picked up
- **In Transit** → **Delivered**: Successful delivery
- **Any Status** → **Failed**: Delivery attempt unsuccessful
- **Any Status** → **Cancelled**: Delivery cancelled

### Route Optimization
- Maximum 20 deliveries per route
- Total route duration cannot exceed 10 hours
- Vehicle capacity constraints must be respected
- Time window preferences honored when possible

### Provider Integration
- API providers: Real-time status updates
- Manual providers: Status updates via admin interface
- Automatic failover to backup providers when available
