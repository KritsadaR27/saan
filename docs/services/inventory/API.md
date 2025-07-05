# Inventory Service API Documentation

## Overview

The Inventory Service manages product stock levels, inventory movements, and stock-related operations across multiple stores in the SAAN system.

## Base URL
```
http://localhost:8083
```

## Authentication
All API endpoints require authentication via JWT token in the Authorization header:
```
Authorization: Bearer <jwt_token>
```

## Endpoints

### Product Management

#### Get Product Stock Levels
```http
GET /api/inventory/products/{product_id}/stock
```

**Response:**
```json
{
  "product_id": "uuid",
  "product_name": "Product Name",
  "sku": "SKU123",
  "stock_levels": [
    {
      "store_id": "uuid",
      "store_name": "Store Name",
      "quantity_on_hand": 100.0,
      "reorder_level": 20.0,
      "max_stock": 500.0,
      "is_low_stock": false,
      "last_updated": "2024-01-15T10:30:00Z"
    }
  ]
}
```

#### Get All Products Inventory
```http
GET /api/inventory/products
```

**Query Parameters:**
- `store_id` (optional): Filter by store
- `low_stock` (optional): Filter low stock items only
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20)

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Product Name",
      "sku": "SKU123",
      "barcode": "1234567890",
      "category_id": "uuid",
      "category_name": "Category Name",
      "supplier_id": "uuid",
      "supplier_name": "Supplier Name",
      "cost_price": 50.00,
      "sell_price": 75.00,
      "unit": "pcs",
      "description": "Product description",
      "is_active": true,
      "stock_levels": [
        {
          "store_id": "uuid",
          "store_name": "Store Name",
          "quantity_on_hand": 100.0,
          "reorder_level": 20.0,
          "max_stock": 500.0,
          "is_low_stock": false,
          "last_updated": "2024-01-15T10:30:00Z"
        }
      ],
      "last_updated": "2024-01-15T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

### Stock Movement

#### Record Stock Movement
```http
POST /api/inventory/movements
```

**Request Body:**
```json
{
  "product_id": "uuid",
  "store_id": "uuid",
  "movement_type": "SALE|PURCHASE|ADJUSTMENT|TRANSFER",
  "quantity": 10.0,
  "reference": "ORDER123",
  "notes": "Movement description"
}
```

**Response:**
```json
{
  "id": "uuid",
  "product_id": "uuid",
  "store_id": "uuid",
  "movement_type": "SALE",
  "quantity": 10.0,
  "quantity_before": 100.0,
  "quantity_after": 90.0,
  "reference": "ORDER123",
  "notes": "Movement description",
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### Get Stock Movement History
```http
GET /api/inventory/movements
```

**Query Parameters:**
- `product_id` (optional): Filter by product
- `store_id` (optional): Filter by store
- `movement_type` (optional): Filter by movement type
- `from_date` (optional): Start date (YYYY-MM-DD)
- `to_date` (optional): End date (YYYY-MM-DD)
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20)

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "product_id": "uuid",
      "store_id": "uuid",
      "movement_type": "SALE",
      "quantity": 10.0,
      "quantity_before": 100.0,
      "quantity_after": 90.0,
      "reference": "ORDER123",
      "notes": "Movement description",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 500,
    "total_pages": 25
  }
}
```

### Stock Management

#### Update Stock Level
```http
PUT /api/inventory/stock
```

**Request Body:**
```json
{
  "product_id": "uuid",
  "store_id": "uuid",
  "quantity_on_hand": 150.0,
  "reorder_level": 25.0,
  "max_stock": 600.0
}
```

**Response:**
```json
{
  "product_id": "uuid",
  "store_id": "uuid",
  "store_name": "Store Name",
  "quantity_on_hand": 150.0,
  "reorder_level": 25.0,
  "max_stock": 600.0,
  "is_low_stock": false,
  "last_updated": "2024-01-15T10:30:00Z"
}
```

#### Stock Adjustment
```http
POST /api/inventory/adjustments
```

**Request Body:**
```json
{
  "product_id": "uuid",
  "store_id": "uuid",
  "adjustment_quantity": 5.0,
  "reason": "DAMAGED|EXPIRED|COUNT_CORRECTION|OTHER",
  "notes": "Adjustment reason"
}
```

**Response:**
```json
{
  "id": "uuid",
  "product_id": "uuid",
  "store_id": "uuid",
  "quantity_before": 100.0,
  "quantity_after": 105.0,
  "adjustment_quantity": 5.0,
  "reason": "COUNT_CORRECTION",
  "notes": "Adjustment reason",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Low Stock Alerts

#### Get Low Stock Items
```http
GET /api/inventory/low-stock
```

**Query Parameters:**
- `store_id` (optional): Filter by store

**Response:**
```json
{
  "data": [
    {
      "product_id": "uuid",
      "product_name": "Product Name",
      "sku": "SKU123",
      "store_id": "uuid",
      "store_name": "Store Name",
      "quantity_on_hand": 15.0,
      "reorder_level": 20.0,
      "suggested_order_quantity": 50.0,
      "last_updated": "2024-01-15T10:30:00Z"
    }
  ]
}
```

### Reports

#### Inventory Summary Report
```http
GET /api/inventory/reports/summary
```

**Query Parameters:**
- `store_id` (optional): Filter by store
- `category_id` (optional): Filter by category

**Response:**
```json
{
  "total_products": 250,
  "total_value": 125000.00,
  "low_stock_items": 15,
  "out_of_stock_items": 3,
  "stores": [
    {
      "store_id": "uuid",
      "store_name": "Store Name",
      "total_products": 100,
      "total_value": 50000.00,
      "low_stock_items": 5,
      "out_of_stock_items": 1
    }
  ],
  "generated_at": "2024-01-15T10:30:00Z"
}
```

#### Stock Movement Report
```http
GET /api/inventory/reports/movements
```

**Query Parameters:**
- `store_id` (optional): Filter by store
- `from_date` (required): Start date (YYYY-MM-DD)
- `to_date` (required): End date (YYYY-MM-DD)
- `movement_type` (optional): Filter by movement type

**Response:**
```json
{
  "period": {
    "from_date": "2024-01-01",
    "to_date": "2024-01-15"
  },
  "summary": {
    "total_movements": 1250,
    "total_sales": 500,
    "total_purchases": 300,
    "total_adjustments": 50,
    "total_transfers": 400
  },
  "movements_by_type": [
    {
      "movement_type": "SALE",
      "count": 500,
      "total_quantity": 2500.0
    }
  ],
  "generated_at": "2024-01-15T10:30:00Z"
}
```

## Error Responses

### Error Format
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request data",
    "details": [
      {
        "field": "product_id",
        "message": "Product ID is required"
      }
    ]
  }
}
```

### Error Codes
- `VALIDATION_ERROR` - Invalid request data
- `PRODUCT_NOT_FOUND` - Product not found
- `STORE_NOT_FOUND` - Store not found
- `INSUFFICIENT_STOCK` - Not enough stock for operation
- `DUPLICATE_ENTRY` - Duplicate entry
- `INTERNAL_ERROR` - Internal server error

## Status Codes
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `404` - Not Found
- `409` - Conflict
- `500` - Internal Server Error
