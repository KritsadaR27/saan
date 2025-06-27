import { BaseEntity } from './index'

// Order Types
export interface Order extends BaseEntity {
  customer_id: string
  customer_name: string
  status: OrderStatus
  items: OrderItem[]
  total_amount: number
  currency: string
  notes?: string
  chat_id?: string
  payment_status: PaymentStatus
  shipping_address?: Address
  billing_address?: Address
}

export interface OrderItem extends BaseEntity {
  order_id: string
  product_id: string
  name: string
  quantity: number
  unit_price: number
  total: number
  metadata?: Record<string, any>
}

export interface Address {
  street: string
  city: string
  state: string
  postal_code: string
  country: string
  phone?: string
}

export type OrderStatus = 
  | 'pending' 
  | 'confirmed' 
  | 'processing' 
  | 'shipped' 
  | 'delivered' 
  | 'cancelled' 
  | 'refunded'

export type PaymentStatus = 
  | 'pending' 
  | 'paid' 
  | 'failed' 
  | 'refunded' 
  | 'partial'

// Order API Request/Response Types
export interface CreateOrderRequest {
  customer_id: string
  customer_name: string
  items: Omit<OrderItem, 'id' | 'order_id' | 'total' | 'created_at' | 'updated_at'>[]
  notes?: string
  chat_id?: string
  shipping_address?: Address
  billing_address?: Address
}

export interface UpdateOrderStatusRequest {
  status: OrderStatus
  notes?: string
}

export interface AddOrderItemRequest {
  product_id: string
  name: string
  quantity: number
  unit_price: number
  metadata?: Record<string, any>
}

export interface OrderListResponse {
  orders: Order[]
  total: number
}

export interface OrderStatsResponse {
  total_orders: number
  total_revenue: number
  orders_by_status: Record<OrderStatus, number>
  revenue_by_period: {
    period: string
    revenue: number
  }[]
}
