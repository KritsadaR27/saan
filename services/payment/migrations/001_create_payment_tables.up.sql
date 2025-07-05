-- Migration: 001_create_payment_tables.up.sql
-- Create payment-related tables for SAAN Payment Service

-- Create payment_transactions table
CREATE TABLE IF NOT EXISTS payment_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    customer_id UUID NOT NULL,
    
    -- Payment details
    payment_method VARCHAR(50) NOT NULL CHECK (payment_method IN ('cash', 'bank_transfer', 'cod_cash', 'cod_transfer', 'digital_wallet')),
    payment_channel VARCHAR(50) NOT NULL CHECK (payment_channel IN ('loyverse_pos', 'saan_app', 'saan_chat', 'delivery', 'web_portal')),
    payment_timing VARCHAR(20) NOT NULL CHECK (payment_timing IN ('prepaid', 'cod')),
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'THB',
    
    -- Payment status
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'refunded', 'cancelled')),
    paid_at TIMESTAMP WITH TIME ZONE,
    
    -- Loyverse integration
    loyverse_receipt_id VARCHAR(100),
    loyverse_payment_type VARCHAR(50),
    assigned_store_id VARCHAR(50),
    
    -- Metadata
    metadata JSONB,
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_by UUID
);

-- Create loyverse_stores table
CREATE TABLE IF NOT EXISTS loyverse_stores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_code VARCHAR(50) NOT NULL UNIQUE,
    store_name VARCHAR(255) NOT NULL,
    store_address TEXT,
    region VARCHAR(100),
    manager_id UUID,
    
    -- Store configuration
    is_active BOOLEAN NOT NULL DEFAULT true,
    max_concurrent_orders INTEGER DEFAULT 10,
    avg_processing_time_minutes INTEGER DEFAULT 30,
    business_hours JSONB,
    
    -- Contact information
    phone VARCHAR(20),
    email VARCHAR(255),
    
    -- Loyverse integration
    loyverse_store_id VARCHAR(100),
    loyverse_api_token VARCHAR(500),
    last_sync_at TIMESTAMP WITH TIME ZONE,
    
    -- Metadata
    metadata JSONB,
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_by UUID
);

-- Create payment_delivery_contexts table
CREATE TABLE IF NOT EXISTS payment_delivery_contexts (
    payment_id UUID PRIMARY KEY REFERENCES payment_transactions(id) ON DELETE CASCADE,
    delivery_id UUID NOT NULL,
    driver_id UUID,
    
    -- Delivery details
    delivery_address TEXT NOT NULL,
    delivery_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    estimated_arrival TIMESTAMP WITH TIME ZONE,
    actual_arrival TIMESTAMP WITH TIME ZONE,
    delivery_instructions TEXT,
    
    -- COD collection details
    cod_amount DECIMAL(15,2),
    cod_collected_at TIMESTAMP WITH TIME ZONE,
    cod_collection_method VARCHAR(50),
    
    -- GPS tracking
    pickup_lat DECIMAL(10,8),
    pickup_lng DECIMAL(11,8),
    delivery_lat DECIMAL(10,8),
    delivery_lng DECIMAL(11,8),
    
    -- Metadata
    metadata JSONB,
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create payment_events table for event sourcing
CREATE TABLE IF NOT EXISTS payment_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    payment_id UUID NOT NULL REFERENCES payment_transactions(id),
    order_id UUID,
    customer_id UUID,
    
    -- Event data
    event_data JSONB NOT NULL,
    event_version VARCHAR(10) NOT NULL DEFAULT '1.0',
    event_source VARCHAR(50) NOT NULL DEFAULT 'payment-service',
    
    -- Timing
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for better performance

-- Payment transactions indexes
CREATE INDEX IF NOT EXISTS idx_payment_transactions_order_id ON payment_transactions(order_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_customer_id ON payment_transactions(customer_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_status ON payment_transactions(status);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_assigned_store_id ON payment_transactions(assigned_store_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_payment_channel ON payment_transactions(payment_channel);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_payment_timing ON payment_transactions(payment_timing);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_created_at ON payment_transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_loyverse_receipt_id ON payment_transactions(loyverse_receipt_id);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_payment_transactions_store_date ON payment_transactions(assigned_store_id, created_at);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_customer_date ON payment_transactions(customer_id, created_at);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_order_status ON payment_transactions(order_id, status);

-- Loyverse stores indexes
CREATE INDEX IF NOT EXISTS idx_loyverse_stores_store_code ON loyverse_stores(store_code);
CREATE INDEX IF NOT EXISTS idx_loyverse_stores_region ON loyverse_stores(region);
CREATE INDEX IF NOT EXISTS idx_loyverse_stores_manager_id ON loyverse_stores(manager_id);
CREATE INDEX IF NOT EXISTS idx_loyverse_stores_is_active ON loyverse_stores(is_active);

-- Payment delivery contexts indexes
CREATE INDEX IF NOT EXISTS idx_payment_delivery_contexts_delivery_id ON payment_delivery_contexts(delivery_id);
CREATE INDEX IF NOT EXISTS idx_payment_delivery_contexts_driver_id ON payment_delivery_contexts(driver_id);
CREATE INDEX IF NOT EXISTS idx_payment_delivery_contexts_delivery_status ON payment_delivery_contexts(delivery_status);
CREATE INDEX IF NOT EXISTS idx_payment_delivery_contexts_created_at ON payment_delivery_contexts(created_at);

-- Payment events indexes
CREATE INDEX IF NOT EXISTS idx_payment_events_payment_id ON payment_events(payment_id);
CREATE INDEX IF NOT EXISTS idx_payment_events_event_type ON payment_events(event_type);
CREATE INDEX IF NOT EXISTS idx_payment_events_occurred_at ON payment_events(occurred_at);
CREATE INDEX IF NOT EXISTS idx_payment_events_order_id ON payment_events(order_id);
CREATE INDEX IF NOT EXISTS idx_payment_events_customer_id ON payment_events(customer_id);

-- Create triggers for updating updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers
CREATE TRIGGER update_payment_transactions_updated_at 
    BEFORE UPDATE ON payment_transactions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_loyverse_stores_updated_at 
    BEFORE UPDATE ON loyverse_stores 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payment_delivery_contexts_updated_at 
    BEFORE UPDATE ON payment_delivery_contexts 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create views for common queries

-- View for payment summary by order
CREATE OR REPLACE VIEW order_payment_summary AS
SELECT 
    order_id,
    COUNT(*) as transaction_count,
    SUM(amount) as total_amount,
    SUM(CASE WHEN status = 'completed' THEN amount ELSE 0 END) as paid_amount,
    SUM(CASE WHEN status = 'pending' THEN amount ELSE 0 END) as pending_amount,
    SUM(CASE WHEN status = 'refunded' THEN amount ELSE 0 END) as refunded_amount,
    currency,
    MAX(CASE WHEN status = 'completed' THEN paid_at END) as last_payment_at,
    CASE 
        WHEN SUM(CASE WHEN status = 'completed' THEN amount ELSE 0 END) >= SUM(amount) THEN 'fully_paid'
        WHEN SUM(CASE WHEN status = 'completed' THEN amount ELSE 0 END) > 0 THEN 'partially_paid'
        ELSE 'unpaid'
    END as payment_status,
    ARRAY_AGG(DISTINCT payment_method) as payment_methods
FROM payment_transactions
GROUP BY order_id, currency;

-- View for store workload
CREATE OR REPLACE VIEW store_workload AS
SELECT 
    assigned_store_id as store_code,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_orders,
    COUNT(CASE WHEN status = 'processing' THEN 1 END) as processing_orders,
    COUNT(CASE WHEN DATE(created_at) = CURRENT_DATE THEN 1 END) as total_orders_today,
    NOW() as last_updated
FROM payment_transactions
WHERE assigned_store_id IS NOT NULL
GROUP BY assigned_store_id;

-- Add comments for documentation
COMMENT ON TABLE payment_transactions IS 'Main payment transactions table storing all payment records';
COMMENT ON TABLE loyverse_stores IS 'Loyverse store configurations and integration details';
COMMENT ON TABLE payment_delivery_contexts IS 'Delivery context for COD payments and GPS tracking';
COMMENT ON TABLE payment_events IS 'Event sourcing table for payment-related events';

COMMENT ON COLUMN payment_transactions.payment_timing IS 'When payment occurs: prepaid or cod (cash on delivery)';
COMMENT ON COLUMN payment_transactions.assigned_store_id IS 'Store assigned to handle this payment (for Loyverse integration)';
COMMENT ON COLUMN loyverse_stores.max_concurrent_orders IS 'Maximum concurrent orders this store can handle';
COMMENT ON COLUMN payment_delivery_contexts.cod_amount IS 'Amount to be collected on delivery (may differ from payment amount due to tips)';
