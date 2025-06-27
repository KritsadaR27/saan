export interface ApiResponse<T = any> {
  data: T
  message?: string
  success: boolean
}

// Chat types
export interface Message {
  id: string
  chat_id: string
  content: string
  sender_type: 'customer' | 'ai' | 'agent'
  sender_id: string
  metadata?: Record<string, any>
  created_at: string
}

export interface Chat {
  id: string
  customer_id: string
  customer_name: string
  status: 'active' | 'closed' | 'waiting'
  last_message?: string
  created_at: string
  updated_at: string
}

// Order types
export interface OrderItem {
  id: string
  order_id: string
  product_id: string
  name: string
  quantity: number
  unit_price: number
  total: number
}

export interface Order {
  id: string
  customer_id: string
  customer_name: string
  status: 'pending' | 'confirmed' | 'processing' | 'shipped' | 'delivered' | 'cancelled'
  items: OrderItem[]
  total_amount: number
  currency: string
  notes: string
  chat_id?: string
  created_at: string
  updated_at: string
}

// Product types
export interface Product {
  id: string
  sku: string
  name: string
  description: string
  category: string
  price: number
  currency: string
  stock: number
  min_stock: number
  max_stock: number
  status: 'active' | 'inactive' | 'out_of_stock' | 'discontinued'
  created_at: string
  updated_at: string
}

// Delivery types
export interface Delivery {
  id: string
  order_id: string
  tracking_number: string
  status: 'pending' | 'picked_up' | 'in_transit' | 'delivered' | 'failed'
  pickup_address: string
  delivery_address: string
  estimated_delivery: string
  actual_delivery?: string
  created_at: string
  updated_at: string
}

// User types
export interface User {
  id: string
  name: string
  email: string
  phone?: string
  role: 'customer' | 'admin' | 'agent'
  created_at: string
}
