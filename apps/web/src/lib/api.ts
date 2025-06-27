import axios from 'axios'

const API_BASE_URL = {
  chat: process.env.NEXT_PUBLIC_CHAT_API_URL || 'http://localhost:8001',
  order: process.env.NEXT_PUBLIC_ORDER_API_URL || 'http://localhost:8002',
  inventory: process.env.NEXT_PUBLIC_INVENTORY_API_URL || 'http://localhost:8003',
  delivery: process.env.NEXT_PUBLIC_DELIVERY_API_URL || 'http://localhost:8004',
  finance: process.env.NEXT_PUBLIC_FINANCE_API_URL || 'http://localhost:8005',
}

// Create axios instances for each service
export const chatApi = axios.create({
  baseURL: `${API_BASE_URL.chat}/api/v1`,
  timeout: 10000,
})

export const orderApi = axios.create({
  baseURL: `${API_BASE_URL.order}/api/v1`,
  timeout: 10000,
})

export const inventoryApi = axios.create({
  baseURL: `${API_BASE_URL.inventory}/api/v1`,
  timeout: 10000,
})

export const deliveryApi = axios.create({
  baseURL: `${API_BASE_URL.delivery}/api/v1`,
  timeout: 10000,
})

export const financeApi = axios.create({
  baseURL: `${API_BASE_URL.finance}/api/v1`,
  timeout: 10000,
})

// Request interceptors for adding auth headers
const addAuthInterceptor = (api: typeof chatApi) => {
  api.interceptors.request.use((config) => {
    // Add auth token if available
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  })
}

// Add auth interceptors to all APIs
addAuthInterceptor(chatApi)
addAuthInterceptor(orderApi)
addAuthInterceptor(inventoryApi)
addAuthInterceptor(deliveryApi)
addAuthInterceptor(financeApi)

// Response interceptors for error handling
const addErrorInterceptor = (api: typeof chatApi) => {
  api.interceptors.response.use(
    (response) => response,
    (error) => {
      if (error.response?.status === 401) {
        // Redirect to login or refresh token
        localStorage.removeItem('auth_token')
        window.location.href = '/login'
      }
      return Promise.reject(error)
    }
  )
}

// Add error interceptors to all APIs
addErrorInterceptor(chatApi)
addErrorInterceptor(orderApi)
addErrorInterceptor(inventoryApi)
addErrorInterceptor(deliveryApi)
addErrorInterceptor(financeApi)
