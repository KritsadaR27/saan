// Common API Response Types
export interface ApiResponse<T = any> {
  data: T
  message?: string
  success: boolean
  timestamp: string
}

export interface PaginatedResponse<T> extends ApiResponse<T[]> {
  pagination: {
    page: number
    limit: number
    total: number
    totalPages: number
  }
}

export interface ErrorResponse {
  error: string
  message: string
  statusCode: number
  timestamp: string
}

// Common Entity Types
export interface BaseEntity {
  id: string
  created_at: string
  updated_at: string
}

// User Types
export interface User extends BaseEntity {
  name: string
  email: string
  phone?: string
  role: UserRole
  status: UserStatus
}

export type UserRole = 'customer' | 'admin' | 'agent' | 'manager'
export type UserStatus = 'active' | 'inactive' | 'suspended'

// Re-export specific types
export * from './chat'
export * from './order'
export * from './inventory'
export * from './delivery'
export * from './finance'
export * from './api'
