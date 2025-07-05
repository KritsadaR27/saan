# Payment Service API Documentation

## Overview

The Payment Service handles all payment processing, transaction management, and integration with various payment gateways including Omise, C2P, TrueMoney, and traditional methods.

## Base URL
```
http://localhost:8084
```

## Authentication
All API endpoints require authentication via JWT token in the Authorization header:
```
Authorization: Bearer <jwt_token>
```

## Endpoints

### Payment Creation

#### Create Payment
```http
POST /api/payments
```

**Request Body:**
```json
{
  "order_id": "uuid",
  "customer_id": "uuid",
  "payment_method": "cash|bank_transfer|credit_card|qr_code|omise|c2p|true_money|cod",
  "payment_provider": "omise|c2p|truemoney|manual",
  "amount": 1500.00,
  "currency": "THB",
  "return_url": "https://app.saan.com/payment/return",
  "webhook_url": "https://api.saan.com/webhooks/payment"
}
```

**Response:**
```json
{
  "id": "uuid",
  "order_id": "uuid",
  "customer_id": "uuid",
  "payment_method": "omise",
  "payment_provider": "omise",
  "amount": 1500.00,
  "currency": "THB",
  "status": "pending",
  "payment_url": "https://pay.omise.co/...",
  "external_payment_id": "chrg_12345",
  "expires_at": "2024-01-15T11:30:00Z",
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### Get Payment Details
```http
GET /api/payments/{payment_id}
```

**Response:**
```json
{
  "id": "uuid",
  "order_id": "uuid",
  "customer_id": "uuid",
  "payment_method": "omise",
  "payment_provider": "omise",
  "amount": 1500.00,
  "currency": "THB",
  "status": "completed",
  "payment_gateway_fee": 45.00,
  "net_amount": 1455.00,
  "external_transaction_id": "txn_12345",
  "external_payment_id": "chrg_12345",
  "payment_details": {
    "card_brand": "visa",
    "card_last_four": "4242",
    "authorization_code": "123456"
  },
  "paid_at": "2024-01-15T10:35:00Z",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:35:00Z"
}
```

### Payment Processing

#### Process Payment
```http
POST /api/payments/{payment_id}/process
```

**Request Body:**
```json
{
  "payment_token": "token_from_gateway",
  "payment_details": {
    "card_token": "card_token",
    "source_type": "card"
  }
}
```

**Response:**
```json
{
  "id": "uuid",
  "status": "processing",
  "external_payment_id": "chrg_12345",
  "redirect_url": "https://pay.omise.co/redirects/...",
  "processing_fee": 45.00,
  "estimated_completion": "2024-01-15T10:35:00Z"
}
```

#### Confirm Payment
```http
POST /api/payments/{payment_id}/confirm
```

**Request Body:**
```json
{
  "external_transaction_id": "txn_12345",
  "confirmation_code": "CONF123",
  "receipt_url": "https://receipt.provider.com/123"
}
```

**Response:**
```json
{
  "id": "uuid",
  "status": "completed",
  "confirmed_at": "2024-01-15T10:35:00Z",
  "receipt_url": "https://receipt.provider.com/123",
  "net_amount": 1455.00
}
```

### Payment Management

#### Cancel Payment
```http
POST /api/payments/{payment_id}/cancel
```

**Request Body:**
```json
{
  "reason": "customer_request|timeout|fraud_detected|other",
  "notes": "Customer requested cancellation"
}
```

**Response:**
```json
{
  "id": "uuid",
  "status": "cancelled",
  "cancellation_reason": "customer_request",
  "cancellation_notes": "Customer requested cancellation",
  "cancelled_at": "2024-01-15T10:40:00Z"
}
```

#### Refund Payment
```http
POST /api/payments/{payment_id}/refund
```

**Request Body:**
```json
{
  "amount": 1500.00,
  "reason": "return|defective|duplicate|other",
  "notes": "Product return - defective item"
}
```

**Response:**
```json
{
  "refund_id": "uuid",
  "payment_id": "uuid",
  "amount": 1500.00,
  "refund_fee": 15.00,
  "net_refund": 1485.00,
  "status": "processing",
  "reason": "return",
  "notes": "Product return - defective item",
  "external_refund_id": "rfnd_12345",
  "estimated_completion": "2024-01-18T10:30:00Z",
  "created_at": "2024-01-15T11:00:00Z"
}
```

### Payment Queries

#### List Payments
```http
GET /api/payments
```

**Query Parameters:**
- `order_id` (optional): Filter by order
- `customer_id` (optional): Filter by customer
- `status` (optional): Filter by status
- `payment_method` (optional): Filter by method
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
      "order_id": "uuid",
      "customer_id": "uuid",
      "payment_method": "omise",
      "amount": 1500.00,
      "status": "completed",
      "paid_at": "2024-01-15T10:35:00Z",
      "created_at": "2024-01-15T10:30:00Z"
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

#### Get Payment Status
```http
GET /api/payments/{payment_id}/status
```

**Response:**
```json
{
  "payment_id": "uuid",
  "status": "completed",
  "last_updated": "2024-01-15T10:35:00Z",
  "status_history": [
    {
      "status": "pending",
      "timestamp": "2024-01-15T10:30:00Z"
    },
    {
      "status": "processing",
      "timestamp": "2024-01-15T10:32:00Z"
    },
    {
      "status": "completed",
      "timestamp": "2024-01-15T10:35:00Z"
    }
  ]
}
```

### Payment Methods

#### Get Available Payment Methods
```http
GET /api/payment-methods
```

**Query Parameters:**
- `amount` (optional): Filter by supported amount
- `currency` (optional): Filter by currency

**Response:**
```json
{
  "data": [
    {
      "method": "omise",
      "provider": "omise",
      "name": "Credit/Debit Card",
      "description": "Pay with Visa, Mastercard, or other cards",
      "supported_currencies": ["THB"],
      "min_amount": 20.00,
      "max_amount": 200000.00,
      "processing_fee_percentage": 3.0,
      "processing_fee_fixed": 0.0,
      "is_enabled": true,
      "icon_url": "https://static.saan.com/icons/omise.png"
    },
    {
      "method": "true_money",
      "provider": "truemoney",
      "name": "TrueMoney Wallet",
      "description": "Pay with TrueMoney e-wallet",
      "supported_currencies": ["THB"],
      "min_amount": 1.00,
      "max_amount": 50000.00,
      "processing_fee_percentage": 2.5,
      "processing_fee_fixed": 0.0,
      "is_enabled": true,
      "icon_url": "https://static.saan.com/icons/truemoney.png"
    }
  ]
}
```

### Webhooks

#### Process Payment Webhook
```http
POST /api/webhooks/payment
```

**Request Body (from payment provider):**
```json
{
  "event_type": "payment.completed",
  "payment_id": "chrg_12345",
  "status": "successful",
  "amount": 1500.00,
  "fee": 45.00,
  "net_amount": 1455.00,
  "transaction_id": "txn_12345",
  "timestamp": "2024-01-15T10:35:00Z",
  "signature": "webhook_signature"
}
```

**Response:**
```json
{
  "received": true,
  "payment_updated": true,
  "internal_payment_id": "uuid"
}
```

### Reports

#### Payment Summary Report
```http
GET /api/payments/reports/summary
```

**Query Parameters:**
- `from_date` (required): Start date (YYYY-MM-DD)
- `to_date` (required): End date (YYYY-MM-DD)
- `payment_method` (optional): Filter by method
- `status` (optional): Filter by status

**Response:**
```json
{
  "period": {
    "from_date": "2024-01-01",
    "to_date": "2024-01-15"
  },
  "summary": {
    "total_payments": 1250,
    "total_amount": 187500.00,
    "total_fees": 5625.00,
    "net_amount": 181875.00,
    "successful_payments": 1200,
    "failed_payments": 30,
    "cancelled_payments": 20,
    "refunded_payments": 15,
    "refund_amount": 2250.00
  },
  "by_method": [
    {
      "payment_method": "omise",
      "count": 800,
      "amount": 120000.00,
      "fees": 3600.00,
      "success_rate": 0.96
    },
    {
      "payment_method": "true_money",
      "count": 300,
      "amount": 45000.00,
      "fees": 1125.00,
      "success_rate": 0.98
    }
  ],
  "generated_at": "2024-01-15T10:30:00Z"
}
```

#### Transaction Report
```http
GET /api/payments/reports/transactions
```

**Query Parameters:**
- `from_date` (required): Start date (YYYY-MM-DD)
- `to_date` (required): End date (YYYY-MM-DD)
- `format` (optional): csv|json (default: json)

**Response:**
```json
{
  "data": [
    {
      "payment_id": "uuid",
      "order_id": "uuid",
      "payment_method": "omise",
      "amount": 1500.00,
      "fee": 45.00,
      "net_amount": 1455.00,
      "status": "completed",
      "external_transaction_id": "txn_12345",
      "paid_at": "2024-01-15T10:35:00Z"
    }
  ],
  "total_records": 1250,
  "generated_at": "2024-01-15T10:30:00Z"
}
```

## Error Responses

### Error Format
```json
{
  "error": {
    "code": "PAYMENT_FAILED",
    "message": "Payment processing failed",
    "details": {
      "provider_error": "insufficient_funds",
      "provider_message": "The card has insufficient funds"
    }
  }
}
```

### Error Codes
- `VALIDATION_ERROR` - Invalid request data
- `PAYMENT_NOT_FOUND` - Payment not found
- `PAYMENT_ALREADY_PROCESSED` - Payment already completed
- `INSUFFICIENT_FUNDS` - Insufficient funds in account
- `CARD_DECLINED` - Card was declined
- `PAYMENT_EXPIRED` - Payment session expired
- `GATEWAY_ERROR` - Payment gateway error
- `REFUND_FAILED` - Refund processing failed
- `INVALID_WEBHOOK` - Invalid webhook signature

## Status Codes
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `404` - Not Found
- `409` - Conflict
- `422` - Unprocessable Entity
- `500` - Internal Server Error
- `502` - Gateway Error
- `503` - Service Unavailable
