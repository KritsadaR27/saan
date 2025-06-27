import { BaseEntity } from './index'

// Product Types
export interface Product extends BaseEntity {
  sku: string
  name: string
  description?: string
  category: string
  price: number
  currency: string
  stock: number
  min_stock: number
  max_stock: number
  status: ProductStatus
  images?: string[]
  attributes?: ProductAttribute[]
  supplier_id?: string
}

export interface ProductAttribute {
  name: string
  value: string
  type: AttributeType
}

export interface StockMovement extends BaseEntity {
  product_id: string
  type: MovementType
  quantity: number
  reason: string
  reference?: string
  previous_stock: number
  new_stock: number
  user_id?: string
}

export type ProductStatus = 'active' | 'inactive' | 'out_of_stock' | 'discontinued'
export type AttributeType = 'text' | 'number' | 'boolean' | 'select'
export type MovementType = 'in' | 'out' | 'adjustment' | 'reserve' | 'release'

// Inventory API Request/Response Types
export interface CreateProductRequest {
  sku: string
  name: string
  description?: string
  category: string
  price: number
  currency?: string
  stock: number
  min_stock: number
  max_stock: number
  attributes?: ProductAttribute[]
  supplier_id?: string
}

export interface UpdateProductRequest {
  name?: string
  description?: string
  category?: string
  price?: number
  min_stock?: number
  max_stock?: number
  status?: ProductStatus
  attributes?: ProductAttribute[]
}

export interface StockOperationRequest {
  product_id: string
  quantity: number
  reason: string
  reference?: string
}

export interface ProductListResponse {
  products: Product[]
  total: number
}

export interface StockMovementListResponse {
  movements: StockMovement[]
  total: number
}

export interface InventoryStatsResponse {
  total_products: number
  low_stock_products: number
  out_of_stock_products: number
  total_value: number
  stock_by_category: {
    category: string
    count: number
    value: number
  }[]
}
