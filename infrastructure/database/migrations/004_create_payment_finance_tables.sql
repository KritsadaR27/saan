-- SaaN Payment & Finance Service - Complete Database Schema
-- Migration: 004_create_payment_finance_tables.sql
-- Date: 2025-07-02

-- ================================================================
-- PAYMENT & FINANCE SERVICE TABLES
-- ================================================================

-- 1. Payment Methods
CREATE TABLE payment_methods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL, -- 'Cash', 'Bank Transfer', 'QR Code', 'Credit Card'
    code VARCHAR(20) UNIQUE NOT NULL, -- 'CASH', 'TRANSFER', 'QR', 'CARD'
    type VARCHAR(20) NOT NULL, -- 'cash', 'bank_transfer', 'digital_wallet', 'card'
    
    -- Settings
    is_active BOOLEAN DEFAULT true,
    requires_verification BOOLEAN DEFAULT false,
    processing_fee_percentage DECIMAL(5,2) DEFAULT 0,
    processing_fee_fixed DECIMAL(8,2) DEFAULT 0,
    
    -- Limits
    min_amount DECIMAL(10,2) DEFAULT 0,
    max_amount DECIMAL(10,2),
    daily_limit DECIMAL(10,2),
    
    -- Configuration
    config JSONB, -- {"api_key": "...", "webhook_url": "..."}
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 2. Payment Transactions
CREATE TABLE payment_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id VARCHAR(100) UNIQUE NOT NULL, -- TXN-20250702-001
    
    -- Reference
    reference_type VARCHAR(20) NOT NULL, -- 'order', 'invoice', 'refund'
    reference_id UUID NOT NULL,
    reference_number VARCHAR(50),
    
    -- Payment Details
    payment_method_id UUID REFERENCES payment_methods(id),
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'THB',
    
    -- Processing
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed', 'cancelled', 'refunded'
    gateway VARCHAR(50), -- 'promptpay', 'kbank', 'scb', 'manual'
    gateway_transaction_id VARCHAR(100),
    gateway_response JSONB,
    
    -- Fees
    processing_fee DECIMAL(8,2) DEFAULT 0,
    net_amount DECIMAL(12,2), -- amount - processing_fee
    
    -- Customer Info
    customer_id UUID,
    customer_name VARCHAR(200),
    customer_email VARCHAR(100),
    customer_phone VARCHAR(20),
    
    -- Timestamps
    initiated_at TIMESTAMP DEFAULT NOW(),
    processed_at TIMESTAMP,
    completed_at TIMESTAMP,
    failed_at TIMESTAMP,
    
    -- Notes
    notes TEXT,
    failure_reason TEXT,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 3. Payment Confirmations (for manual verification)
CREATE TABLE payment_confirmations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_transaction_id UUID REFERENCES payment_transactions(id),
    
    -- Confirmation Details
    confirmation_type VARCHAR(20), -- 'bank_slip', 'receipt', 'screenshot'
    confirmation_number VARCHAR(100),
    
    -- Evidence
    evidence_files JSONB, -- ["slip1.jpg", "receipt.pdf"]
    evidence_description TEXT,
    
    -- Bank Details (for transfers)
    bank_name VARCHAR(100),
    account_number VARCHAR(50),
    transfer_date DATE,
    transfer_time TIME,
    transfer_amount DECIMAL(12,2),
    
    -- Verification
    verification_status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'verified', 'rejected'
    verified_by_user_id UUID,
    verified_at TIMESTAMP,
    verification_notes TEXT,
    
    -- Submission
    submitted_by_user_id UUID,
    submitted_at TIMESTAMP DEFAULT NOW(),
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 4. Invoices
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number VARCHAR(20) UNIQUE NOT NULL, -- INV-20250702-001
    
    -- Reference
    reference_type VARCHAR(20), -- 'order', 'service', 'subscription'
    reference_id UUID,
    reference_number VARCHAR(50),
    
    -- Customer Info
    customer_id UUID,
    customer_name VARCHAR(200) NOT NULL,
    customer_email VARCHAR(100),
    customer_phone VARCHAR(20),
    
    -- Billing Address
    billing_address TEXT,
    billing_city VARCHAR(100),
    billing_province VARCHAR(100),
    billing_postal_code VARCHAR(10),
    
    -- Tax Info
    customer_tax_id VARCHAR(50),
    customer_business_name VARCHAR(200),
    
    -- Invoice Details
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    
    -- Financial
    subtotal DECIMAL(12,2) NOT NULL,
    tax_rate DECIMAL(5,2) DEFAULT 7.00, -- 7% VAT in Thailand
    tax_amount DECIMAL(10,2) DEFAULT 0,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    total_amount DECIMAL(12,2) NOT NULL,
    
    -- Payment
    payment_status VARCHAR(20) DEFAULT 'unpaid', -- 'unpaid', 'partial', 'paid', 'overdue', 'cancelled'
    paid_amount DECIMAL(12,2) DEFAULT 0,
    remaining_amount DECIMAL(12,2),
    
    -- Status
    status VARCHAR(20) DEFAULT 'draft', -- 'draft', 'sent', 'paid', 'cancelled', 'refunded'
    
    -- Timestamps
    sent_at TIMESTAMP,
    paid_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    
    -- Notes
    notes TEXT,
    terms_and_conditions TEXT,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 5. Invoice Items
CREATE TABLE invoice_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID REFERENCES invoices(id),
    
    -- Item Details
    product_id UUID, -- Reference to products table
    description TEXT NOT NULL,
    quantity DECIMAL(10,2) NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(12,2) GENERATED ALWAYS AS (quantity * unit_price) STORED,
    
    -- Tax
    tax_rate DECIMAL(5,2) DEFAULT 7.00,
    tax_amount DECIMAL(10,2),
    
    -- Metadata
    sku VARCHAR(50),
    unit_of_measure VARCHAR(20),
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 6. Refunds
CREATE TABLE refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    refund_number VARCHAR(20) UNIQUE NOT NULL, -- REF-20250702-001
    
    -- Original Transaction
    original_transaction_id UUID REFERENCES payment_transactions(id),
    original_amount DECIMAL(12,2),
    
    -- Refund Details
    refund_amount DECIMAL(12,2) NOT NULL,
    refund_reason VARCHAR(100) NOT NULL,
    refund_type VARCHAR(20), -- 'full', 'partial', 'processing_fee'
    
    -- Processing
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed', 'cancelled'
    gateway VARCHAR(50),
    gateway_refund_id VARCHAR(100),
    gateway_response JSONB,
    
    -- Approval
    requested_by_user_id UUID,
    approved_by_user_id UUID,
    approved_at TIMESTAMP,
    
    -- Processing Times
    processed_at TIMESTAMP,
    completed_at TIMESTAMP,
    
    -- Notes
    notes TEXT,
    failure_reason TEXT,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 7. Accounting Entries (Double-entry bookkeeping)
CREATE TABLE accounting_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_number VARCHAR(20) UNIQUE NOT NULL, -- ACC-20250702-001
    
    -- Transaction Reference
    reference_type VARCHAR(20), -- 'payment', 'invoice', 'refund', 'expense', 'adjustment'
    reference_id UUID,
    reference_number VARCHAR(50),
    
    -- Entry Details
    entry_date DATE NOT NULL,
    description TEXT NOT NULL,
    total_amount DECIMAL(12,2) NOT NULL,
    
    -- Status
    status VARCHAR(20) DEFAULT 'draft', -- 'draft', 'posted', 'cancelled'
    
    -- Approval
    created_by_user_id UUID,
    approved_by_user_id UUID,
    approved_at TIMESTAMP,
    
    -- Notes
    notes TEXT,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 8. Accounting Entry Lines (Debit/Credit lines)
CREATE TABLE accounting_entry_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    accounting_entry_id UUID REFERENCES accounting_entries(id),
    
    -- Account Details
    account_code VARCHAR(20) NOT NULL, -- '1001', '2001', '3001', etc.
    account_name VARCHAR(100) NOT NULL, -- 'Cash', 'Accounts Receivable', 'Revenue'
    account_type VARCHAR(20) NOT NULL, -- 'asset', 'liability', 'equity', 'revenue', 'expense'
    
    -- Amount (positive for debit, negative for credit)
    debit_amount DECIMAL(12,2) DEFAULT 0,
    credit_amount DECIMAL(12,2) DEFAULT 0,
    
    -- Description
    description TEXT,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 9. Financial Reports Cache
CREATE TABLE financial_reports_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_type VARCHAR(50) NOT NULL, -- 'income_statement', 'balance_sheet', 'cash_flow'
    period_type VARCHAR(20) NOT NULL, -- 'daily', 'monthly', 'quarterly', 'yearly'
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    
    -- Report Data
    report_data JSONB NOT NULL,
    summary JSONB,
    
    -- Metadata
    generated_at TIMESTAMP DEFAULT NOW(),
    generated_by_user_id UUID,
    is_final BOOLEAN DEFAULT false, -- true if period is closed
    
    -- Cache Control
    expires_at TIMESTAMP,
    last_updated TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(report_type, period_type, period_start, period_end)
);

-- 10. Bank Accounts (Company bank accounts)
CREATE TABLE bank_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_name VARCHAR(100) NOT NULL,
    bank_code VARCHAR(10),
    account_number VARCHAR(50) NOT NULL,
    account_name VARCHAR(200) NOT NULL,
    account_type VARCHAR(20), -- 'current', 'savings', 'fixed_deposit'
    
    -- Balance
    current_balance DECIMAL(12,2) DEFAULT 0,
    available_balance DECIMAL(12,2) DEFAULT 0,
    last_reconciled_at TIMESTAMP,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    is_primary BOOLEAN DEFAULT false,
    
    -- API Integration
    bank_api_config JSONB,
    auto_sync_enabled BOOLEAN DEFAULT false,
    last_sync_at TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- ================================================================
-- INDEXES FOR PAYMENT & FINANCE
-- ================================================================

-- Payment Methods indexes
CREATE INDEX idx_payment_methods_code ON payment_methods(code);
CREATE INDEX idx_payment_methods_active ON payment_methods(is_active);

-- Payment Transactions indexes
CREATE INDEX idx_payment_transactions_id ON payment_transactions(transaction_id);
CREATE INDEX idx_payment_transactions_reference ON payment_transactions(reference_type, reference_id);
CREATE INDEX idx_payment_transactions_status ON payment_transactions(status);
CREATE INDEX idx_payment_transactions_gateway ON payment_transactions(gateway, gateway_transaction_id);
CREATE INDEX idx_payment_transactions_customer ON payment_transactions(customer_id);
CREATE INDEX idx_payment_transactions_date ON payment_transactions(created_at);

-- Payment Confirmations indexes
CREATE INDEX idx_payment_confirmations_transaction ON payment_confirmations(payment_transaction_id);
CREATE INDEX idx_payment_confirmations_status ON payment_confirmations(verification_status);

-- Invoices indexes
CREATE INDEX idx_invoices_number ON invoices(invoice_number);
CREATE INDEX idx_invoices_customer ON invoices(customer_id);
CREATE INDEX idx_invoices_status ON invoices(status, payment_status);
CREATE INDEX idx_invoices_date ON invoices(invoice_date);
CREATE INDEX idx_invoices_due_date ON invoices(due_date);

-- Invoice Items indexes
CREATE INDEX idx_invoice_items_invoice ON invoice_items(invoice_id);
CREATE INDEX idx_invoice_items_product ON invoice_items(product_id);

-- Refunds indexes
CREATE INDEX idx_refunds_number ON refunds(refund_number);
CREATE INDEX idx_refunds_original_transaction ON refunds(original_transaction_id);
CREATE INDEX idx_refunds_status ON refunds(status);

-- Accounting Entries indexes
CREATE INDEX idx_accounting_entries_number ON accounting_entries(entry_number);
CREATE INDEX idx_accounting_entries_reference ON accounting_entries(reference_type, reference_id);
CREATE INDEX idx_accounting_entries_date ON accounting_entries(entry_date);
CREATE INDEX idx_accounting_entries_status ON accounting_entries(status);

-- Accounting Entry Lines indexes
CREATE INDEX idx_entry_lines_entry ON accounting_entry_lines(accounting_entry_id);
CREATE INDEX idx_entry_lines_account ON accounting_entry_lines(account_code);
CREATE INDEX idx_entry_lines_type ON accounting_entry_lines(account_type);

-- Financial Reports Cache indexes
CREATE INDEX idx_financial_reports_type_period ON financial_reports_cache(report_type, period_type, period_start, period_end);
CREATE INDEX idx_financial_reports_expires ON financial_reports_cache(expires_at);

-- Bank Accounts indexes
CREATE INDEX idx_bank_accounts_number ON bank_accounts(account_number);
CREATE INDEX idx_bank_accounts_active ON bank_accounts(is_active);
CREATE INDEX idx_bank_accounts_primary ON bank_accounts(is_primary);
