# Order Service API Documentation

## Overview
The Order Service manages all order-related operations in the Saan System. It provides comprehensive APIs for order lifecycle management, customer order tracking, and statistical reporting.

**Base URL:** `http://order:8081` (Internal) / `http://localhost:8081` (External)  
**API Version:** v1  
**Protocol:** HTTP/REST  

## Authentication

The API uses JWT-based authentication with role-based access control (RBAC).

### Roles
- `sales`: Can view and create orders
- `manager`: Can view, create, update orders and access statistics  
- `admin`: Full access to all operations including bulk updates and exports
- `ai_assistant`: Special role for chat-based order operations

### Headers
```
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

## API Endpoints

### Health Check

#### GET /health
Check service health status.

**Authentication:** None required

**Response:**
```json
{
  "service": "order-service",
  "status": "healthy"
}
```

### Order Management

#### POST /api/v1/orders
Create a new order.

**Authentication:** Required (sales, manager, admin)

**Request Body:**
```json
{
  "customer_id": "uuid",
  "items": [
    {
      "product_id": "uuid",
      "quantity": 2,
      "unit_price": 100.00
    }
  ],
  "shipping_address": {
    "street": "123 Main St",
    "city": "Bangkok",
    "postal_code": "10100",
    "country": "Thailand"
  },
  "notes": "Special delivery instructions"
}
```

**Response (201):**
```json
{
  "id": "uuid",
  "order_number": "ORD-2025-001234",
  "customer_id": "uuid",
  "status": "pending",
  "total_amount": 200.00,
  "created_at": "2025-06-29T10:00:00Z",
  "items": [...],
  "shipping_address": {...}
}
```

#### GET /api/v1/orders
List orders with pagination and filtering.

**Authentication:** Required (sales, manager, admin)

**Query Parameters:**
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 10, max: 100)
- `status` (string): Filter by order status
- `customer_id` (uuid): Filter by customer
- `start_date` (date): Filter orders from date (YYYY-MM-DD)
- `end_date` (date): Filter orders to date (YYYY-MM-DD)

**Response (200):**
```json
{
  "orders": [...],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 150,
    "total_pages": 15,
    "has_next": true,
    "has_prev": false
  }
}
```

#### GET /api/v1/orders/:id
Get order details by ID.

**Authentication:** Required (sales, manager, admin)

**Path Parameters:**
- `id` (uuid): Order ID

**Response (200):**
```json
{
  "id": "uuid",
  "order_number": "ORD-2025-001234",
  "customer_id": "uuid",
  "status": "confirmed",
  "total_amount": 200.00,
  "created_at": "2025-06-29T10:00:00Z",
  "updated_at": "2025-06-29T10:05:00Z",
  "items": [
    {
      "id": "uuid",
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
  "audit_trail": [
    {
      "action": "created",
      "timestamp": "2025-06-29T10:00:00Z",
      "user_id": "uuid"
    }
  ]
}
```

#### PUT /api/v1/orders/:id
Update order details.

**Authentication:** Required (manager, admin)

**Path Parameters:**
- `id` (uuid): Order ID

**Request Body:** Same as POST /orders

**Response (200):** Updated order object

#### DELETE /api/v1/orders/:id
Delete an order (soft delete).

**Authentication:** Required (admin)

**Path Parameters:**
- `id` (uuid): Order ID

**Response (204):** No content

#### PATCH /api/v1/orders/:id/status
Update order status.

**Authentication:** Required (manager, admin)

**Path Parameters:**
- `id` (uuid): Order ID

**Request Body:**
```json
{
  "status": "confirmed",
  "reason": "Customer confirmed payment"
}
```

**Response (200):** Updated order object

#### POST /api/v1/orders/:id/confirm-with-override
Confirm order with stock override (when inventory is insufficient).

**Authentication:** Required (orders:override_stock permission)

**Path Parameters:**
- `id` (uuid): Order ID

**Request Body:**
```json
{
  "override_reason": "Special customer request",
  "expected_restock_date": "2025-07-01T00:00:00Z"
}
```

**Response (200):** Updated order object

#### GET /api/v1/orders/status/:status
Get orders by status.

**Authentication:** Required (sales, manager, admin)

**Path Parameters:**
- `status` (string): Order status (pending, confirmed, shipped, delivered, cancelled)

**Query Parameters:** Same as GET /orders

**Response (200):** Orders list with pagination

#### GET /api/v1/customers/:customerId/orders
Get orders for a specific customer.

**Authentication:** Required (sales, manager, admin)

**Path Parameters:**
- `customerId` (uuid): Customer ID

**Query Parameters:** Same as GET /orders

**Response (200):** Orders list with pagination

### Chat-Based Order Operations

#### POST /api/v1/chat/orders
Create order from chat conversation.

**Authentication:** Required (ai_assistant, manager, admin)

**Request Body:**
```json
{
  "chat_id": "uuid",
  "customer_id": "uuid",
  "order_data": {
    "items": [...],
    "preferences": {...}
  },
  "template_type": "quick_order"
}
```

**Response (201):** Created order object

#### POST /api/v1/chat/orders/:id/confirm
Confirm chat-initiated order.

**Authentication:** Required (ai_assistant, manager, admin)

**Response (200):** Updated order object

#### POST /api/v1/chat/orders/:id/cancel
Cancel chat-initiated order.

**Authentication:** Required (ai_assistant, manager, admin)

**Response (200):** Updated order object

#### POST /api/v1/chat/orders/:id/summary
Generate order summary for chat.

**Authentication:** Required (ai_assistant, manager, admin)

**Response (200):**
```json
{
  "summary": "Order #ORD-2025-001234 for 2 items, total ฿200.00",
  "details": {...},
  "suggested_responses": [...]
}
```

### Admin Operations

#### POST /api/v1/admin/orders
Create order for specific customer (admin).

**Authentication:** Required (admin:create_order permission)

**Request Body:** Extended order creation with admin fields

#### POST /api/v1/admin/orders/:id/link-chat
Link order to chat conversation.

**Authentication:** Required (admin:link_chat permission)

**Request Body:**
```json
{
  "chat_id": "uuid",
  "link_type": "support"
}
```

#### POST /api/v1/admin/orders/bulk-status
Bulk update order statuses.

**Authentication:** Required (admin:bulk_update permission)

**Request Body:**
```json
{
  "order_ids": ["uuid1", "uuid2"],
  "status": "shipped",
  "reason": "Bulk shipment processing"
}
```

#### GET /api/v1/admin/orders/export
Export orders to CSV/Excel.

**Authentication:** Required (admin:export permission)

**Query Parameters:**
- `format` (string): csv or excel
- `start_date`, `end_date`: Date range
- `status`: Filter by status

**Response:** File download

### Statistics

#### GET /api/v1/stats/daily
Get daily order statistics.

**Authentication:** Required (manager, admin)

**Query Parameters:**
- `date` (date): Specific date (default: today)

**Response (200):**
```json
{
  "date": "2025-06-29",
  "total_orders": 45,
  "total_revenue": 12500.00,
  "average_order_value": 277.78,
  "orders_by_status": {
    "pending": 10,
    "confirmed": 20,
    "shipped": 15
  }
}
```

#### GET /api/v1/stats/monthly
Get monthly order statistics.

**Authentication:** Required (manager, admin)

**Query Parameters:**
- `year` (int): Year (default: current year)
- `month` (int): Month (default: current month)

**Response (200):** Monthly statistics object

#### GET /api/v1/stats/top-products
Get top-selling products.

**Authentication:** Required (manager, admin)

**Query Parameters:**
- `limit` (int): Number of products (default: 10)
- `period` (string): daily, weekly, monthly (default: monthly)

**Response (200):**
```json
{
  "products": [
    {
      "product_id": "uuid",
      "product_name": "Product A",
      "total_quantity": 150,
      "total_revenue": 15000.00,
      "order_count": 75
    }
  ]
}
```

#### GET /api/v1/stats/customer/:customer_id
Get customer order statistics.

**Authentication:** Required (manager, admin)

**Path Parameters:**
- `customer_id` (uuid): Customer ID

**Response (200):** Customer statistics object

#### GET /api/v1/stats/overview
Get overall order statistics.

**Authentication:** Required (manager, admin)

**Response (200):** Overall statistics dashboard data

## Error Codes

| Code | Description |
|------|-------------|
| 400  | Bad Request - Invalid input data |
| 401  | Unauthorized - Missing or invalid authentication |
| 403  | Forbidden - Insufficient permissions |
| 404  | Not Found - Resource not found |
| 409  | Conflict - Resource conflict (e.g., duplicate order) |
| 422  | Unprocessable Entity - Validation errors |
| 500  | Internal Server Error - Server error |

### Error Response Format
```json
{
  "error": "error_code",
  "message": "Human readable error message",
  "details": {
    "field": "Specific field error"
  },
  "timestamp": "2025-06-29T10:00:00Z"
}
```

## Integration Guide

### Getting Started

1. **Obtain API Token**
   ```bash
   # Get JWT token from User Service
   curl -X POST http://user:8088/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"user@example.com","password":"password"}'
   ```

2. **Create Your First Order**
   ```bash
   curl -X POST http://order:8081/api/v1/orders \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "customer_id": "customer-uuid",
       "items": [
         {
           "product_id": "product-uuid",
           "quantity": 1,
           "unit_price": 100.00
         }
       ]
     }'
   ```

3. **Monitor Order Status**
   ```bash
   curl -H "Authorization: Bearer $TOKEN" \
     http://order:8081/api/v1/orders/order-uuid
   ```

### Rate Limiting
- Default: 100 requests per hour per user
- Manager/Admin: 1000 requests per hour
- AI Assistant: 500 requests per hour

### Pagination
All list endpoints support pagination:
- Use `page` and `limit` parameters
- Maximum `limit` is 100
- Response includes pagination metadata

### Filtering and Sorting
- Date filters use ISO 8601 format (YYYY-MM-DD)
- Status filters are case-insensitive
- Results are sorted by `created_at` DESC by default

### Event Integration
The Order Service publishes events to the message bus for real-time updates. See [events.md](events.md) for details.

### Development Environment
Following PROJECT_RULES.md:
- Internal service communication: `http://order:8081`
- External access: `http://localhost:8081`
- Use docker-compose for local development
- Database: PostgreSQL at `postgres:5432`

### Support
For technical support and API questions:
- Documentation: `/docs`
- Health check: `/health`
- Service logs: `docker-compose logs -f order`
