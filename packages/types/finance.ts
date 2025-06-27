import { BaseEntity } from './index'

// Finance Types
export interface Payment extends BaseEntity {
  order_id: string
  amount: number
  currency: string
  method: PaymentMethod
  status: PaymentStatus
  provider: PaymentProvider
  provider_transaction_id?: string
  reference?: string
  notes?: string
  processed_at?: string
  refunded_amount?: number
}

export interface Invoice extends BaseEntity {
  order_id: string
  invoice_number: string
  amount: number
  currency: string
  tax_amount: number
  discount_amount: number
  total_amount: number
  status: InvoiceStatus
  due_date: string
  paid_date?: string
  notes?: string
}

export interface Transaction extends BaseEntity {
  type: TransactionType
  amount: number
  currency: string
  description: string
  reference?: string
  order_id?: string
  payment_id?: string
  category: TransactionCategory
  account: string
}

export type PaymentMethod = 'credit_card' | 'debit_card' | 'bank_transfer' | 'e_wallet' | 'cash' | 'crypto'
export type PaymentStatus = 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled' | 'refunded'
export type PaymentProvider = 'stripe' | 'paypal' | 'omise' | 'promptpay' | 'bank' | 'cash'

export type InvoiceStatus = 'draft' | 'sent' | 'paid' | 'overdue' | 'cancelled'
export type TransactionType = 'income' | 'expense' | 'transfer'
export type TransactionCategory = 'sales' | 'refund' | 'fee' | 'tax' | 'other'

// Finance API Request/Response Types
export interface CreatePaymentRequest {
  order_id: string
  amount: number
  currency?: string
  method: PaymentMethod
  provider: PaymentProvider
  reference?: string
  notes?: string
}

export interface ProcessPaymentRequest {
  payment_id: string
  provider_transaction_id?: string
  notes?: string
}

export interface RefundPaymentRequest {
  payment_id: string
  amount?: number
  reason: string
}

export interface CreateInvoiceRequest {
  order_id: string
  tax_amount?: number
  discount_amount?: number
  due_date: string
  notes?: string
}

export interface FinancialStatsResponse {
  total_revenue: number
  total_expenses: number
  net_profit: number
  pending_payments: number
  overdue_invoices: number
  revenue_by_period: {
    period: string
    revenue: number
    expenses: number
    profit: number
  }[]
  payment_methods: {
    method: PaymentMethod
    count: number
    amount: number
  }[]
}
