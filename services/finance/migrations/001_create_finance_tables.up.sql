-- Create Finance Service Database Schema
-- Migration: 001_create_finance_tables.up.sql

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Daily Cash Summaries Table
CREATE TABLE daily_cash_summaries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_date DATE NOT NULL,
    branch_id UUID,
    vehicle_id UUID,
    opening_cash DECIMAL(15,2) DEFAULT 0.00,
    total_sales DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    cod_collections DECIMAL(15,2) DEFAULT 0.00,
    
    -- Profit First allocations
    profit_allocation DECIMAL(15,2) DEFAULT 0.00,
    owner_pay_allocation DECIMAL(15,2) DEFAULT 0.00,
    tax_allocation DECIMAL(15,2) DEFAULT 0.00,
    available_for_expenses DECIMAL(15,2) DEFAULT 0.00,
    
    -- Manual entries
    manual_expenses DECIMAL(15,2) DEFAULT 0.00,
    supplier_transfers DECIMAL(15,2) DEFAULT 0.00,
    other_transfers DECIMAL(15,2) DEFAULT 0.00,
    
    closing_cash DECIMAL(15,2) DEFAULT 0.00,
    reconciled BOOLEAN DEFAULT FALSE,
    reconciled_by_user_id UUID,
    reconciled_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT unique_daily_summary_entity_date UNIQUE (business_date, branch_id, vehicle_id),
    CONSTRAINT check_entity_specified CHECK (
        (branch_id IS NOT NULL AND vehicle_id IS NULL) OR 
        (branch_id IS NULL AND vehicle_id IS NOT NULL) OR
        (branch_id IS NULL AND vehicle_id IS NULL)
    )
);

-- Profit Allocation Rules Table
CREATE TABLE profit_allocation_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_id UUID,
    vehicle_id UUID,
    profit_percentage DECIMAL(5,2) NOT NULL DEFAULT 5.00,
    owner_pay_percentage DECIMAL(5,2) NOT NULL DEFAULT 50.00,
    tax_percentage DECIMAL(5,2) NOT NULL DEFAULT 15.00,
    effective_from DATE NOT NULL DEFAULT CURRENT_DATE,
    effective_to DATE,
    is_active BOOLEAN DEFAULT TRUE,
    updated_by_user_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT check_allocation_entity CHECK (
        (branch_id IS NOT NULL AND vehicle_id IS NULL) OR 
        (branch_id IS NULL AND vehicle_id IS NOT NULL) OR
        (branch_id IS NULL AND vehicle_id IS NULL)
    ),
    CONSTRAINT check_percentage_sum CHECK (
        profit_percentage + owner_pay_percentage + tax_percentage <= 100.00
    ),
    CONSTRAINT check_effective_dates CHECK (
        effective_to IS NULL OR effective_to > effective_from
    )
);

-- Cash Transfer Batches Table
CREATE TABLE cash_transfer_batches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    batch_reference VARCHAR(50) UNIQUE NOT NULL,
    branch_id UUID,
    vehicle_id UUID,
    total_amount DECIMAL(15,2) NOT NULL,
    transfer_count INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    scheduled_at TIMESTAMP WITH TIME ZONE,
    processed_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    authorized_by UUID NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT check_batch_status CHECK (
        status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')
    ),
    CONSTRAINT check_batch_entity CHECK (
        (branch_id IS NOT NULL AND vehicle_id IS NULL) OR 
        (branch_id IS NULL AND vehicle_id IS NOT NULL) OR
        (branch_id IS NULL AND vehicle_id IS NULL)
    )
);

-- Cash Transfers Table
CREATE TABLE cash_transfers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    batch_id UUID REFERENCES cash_transfer_batches(id) ON DELETE CASCADE,
    transfer_type VARCHAR(30) NOT NULL,
    recipient_name VARCHAR(255) NOT NULL,
    recipient_account VARCHAR(100) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'THB',
    reference VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    
    -- Bank transfer details
    bank_name VARCHAR(255),
    account_number VARCHAR(50),
    transaction_ref VARCHAR(100),
    
    scheduled_at TIMESTAMP WITH TIME ZONE,
    executed_at TIMESTAMP WITH TIME ZONE,
    confirmed_at TIMESTAMP WITH TIME ZONE,
    failure_reason TEXT,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT check_transfer_type CHECK (
        transfer_type IN ('supplier_payment', 'expense', 'central_transfer', 'bank_transfer')
    ),
    CONSTRAINT check_transfer_status CHECK (
        status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')
    ),
    CONSTRAINT check_positive_amount CHECK (amount > 0)
);

-- Expense Entries Table
CREATE TABLE expense_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    summary_id UUID NOT NULL REFERENCES daily_cash_summaries(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    receipt TEXT, -- File path or URL
    entered_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT check_expense_amount CHECK (amount > 0),
    CONSTRAINT check_expense_category CHECK (
        category IN ('fuel', 'meals', 'utilities', 'maintenance', 'supplies', 'other')
    )
);

-- Cash Flow Records Table
CREATE TABLE cash_flow_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(20) NOT NULL,
    entity_id UUID NOT NULL,
    transaction_type VARCHAR(10) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    description TEXT NOT NULL,
    reference VARCHAR(100) NOT NULL,
    running_balance DECIMAL(15,2) NOT NULL,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT check_entity_type CHECK (
        entity_type IN ('branch', 'vehicle', 'central')
    ),
    CONSTRAINT check_transaction_type CHECK (
        transaction_type IN ('inflow', 'outflow')
    ),
    CONSTRAINT check_flow_amount CHECK (amount > 0)
);

-- Indexes for performance
CREATE INDEX idx_daily_summaries_date ON daily_cash_summaries(business_date);
CREATE INDEX idx_daily_summaries_branch ON daily_cash_summaries(branch_id);
CREATE INDEX idx_daily_summaries_vehicle ON daily_cash_summaries(vehicle_id);
CREATE INDEX idx_daily_summaries_reconciled ON daily_cash_summaries(reconciled);

CREATE INDEX idx_allocation_rules_active ON profit_allocation_rules(is_active);
CREATE INDEX idx_allocation_rules_branch ON profit_allocation_rules(branch_id);
CREATE INDEX idx_allocation_rules_vehicle ON profit_allocation_rules(vehicle_id);
CREATE INDEX idx_allocation_rules_effective ON profit_allocation_rules(effective_from, effective_to);

CREATE INDEX idx_transfer_batches_status ON cash_transfer_batches(status);
CREATE INDEX idx_transfer_batches_date ON cash_transfer_batches(created_at);
CREATE INDEX idx_transfer_batches_branch ON cash_transfer_batches(branch_id);
CREATE INDEX idx_transfer_batches_vehicle ON cash_transfer_batches(vehicle_id);

CREATE INDEX idx_transfers_batch ON cash_transfers(batch_id);
CREATE INDEX idx_transfers_status ON cash_transfers(status);
CREATE INDEX idx_transfers_type ON cash_transfers(transfer_type);

CREATE INDEX idx_expenses_summary ON expense_entries(summary_id);
CREATE INDEX idx_expenses_category ON expense_entries(category);
CREATE INDEX idx_expenses_date ON expense_entries(created_at);

CREATE INDEX idx_cash_flow_entity ON cash_flow_records(entity_type, entity_id);
CREATE INDEX idx_cash_flow_date ON cash_flow_records(created_at);
CREATE INDEX idx_cash_flow_type ON cash_flow_records(transaction_type);

-- Update timestamp triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_daily_cash_summaries_updated_at 
    BEFORE UPDATE ON daily_cash_summaries 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_profit_allocation_rules_updated_at 
    BEFORE UPDATE ON profit_allocation_rules 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cash_transfer_batches_updated_at 
    BEFORE UPDATE ON cash_transfer_batches 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cash_transfers_updated_at 
    BEFORE UPDATE ON cash_transfers 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
