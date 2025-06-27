import { BaseEntity } from './index'

// Delivery Types
export interface Delivery extends BaseEntity {
  order_id: string
  tracking_number: string
  status: DeliveryStatus
  pickup_address: Address
  delivery_address: Address
  estimated_pickup?: string
  actual_pickup?: string
  estimated_delivery?: string
  actual_delivery?: string
  delivery_fee: number
  notes?: string
  driver_id?: string
  tracking_events: TrackingEvent[]
}

export interface TrackingEvent extends BaseEntity {
  delivery_id: string
  status: DeliveryStatus
  description: string
  location?: string
  timestamp: string
  user_id?: string
}

export interface Address {
  street: string
  city: string
  state: string
  postal_code: string
  country: string
  phone?: string
  coordinates?: {
    latitude: number
    longitude: number
  }
}

export type DeliveryStatus = 
  | 'pending' 
  | 'confirmed' 
  | 'picked_up' 
  | 'in_transit' 
  | 'out_for_delivery' 
  | 'delivered' 
  | 'failed' 
  | 'returned'

// Delivery API Request/Response Types
export interface CreateDeliveryRequest {
  order_id: string
  pickup_address: Address
  delivery_address: Address
  estimated_pickup?: string
  estimated_delivery?: string
  delivery_fee: number
  notes?: string
}

export interface UpdateDeliveryStatusRequest {
  status: DeliveryStatus
  description: string
  location?: string
  notes?: string
}

export interface DeliveryListResponse {
  deliveries: Delivery[]
  total: number
}

export interface DeliveryStatsResponse {
  total_deliveries: number
  pending_deliveries: number
  in_transit_deliveries: number
  completed_deliveries: number
  failed_deliveries: number
  average_delivery_time: number
  delivery_performance: {
    on_time: number
    delayed: number
    failed: number
  }
}
