-- Migration: 002_add_audit_and_events.sql
-- Description: Add audit log table, outbox pattern table, and additional columns to orders table
-- Created: 2025-06-28

-- Create order_audit_log table for tracking all order changes
CREATE TABLE order_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    user_id VARCHAR(255),
    action VARCHAR(50) NOT NULL,
    details JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Foreign key constraint
    CONSTRAINT fk_audit_order_id FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);

-- Create indexes for order_audit_log
CREATE INDEX idx_audit_order_id ON order_audit_log(order_id);
CREATE INDEX idx_audit_timestamp ON order_audit_log(timestamp);

-- Create order_events_outbox table for outbox pattern implementation
CREATE TABLE order_events_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    sent_at TIMESTAMP WITH TIME ZONE,
    retry_count INTEGER DEFAULT 0,
    
    -- Foreign key constraint
    CONSTRAINT fk_events_order_id FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    
    -- Check constraints
    CONSTRAINT chk_event_status CHECK (status IN ('pending', 'sent', 'failed', 'cancelled')),
    CONSTRAINT chk_retry_count CHECK (retry_count >= 0)
);

-- Create indexes for order_events_outbox
CREATE INDEX idx_events_status ON order_events_outbox(status);
CREATE INDEX idx_events_created_at ON order_events_outbox(created_at);
CREATE INDEX idx_events_order_id ON order_events_outbox(order_id);

-- Add new columns to orders table
ALTER TABLE orders 
ADD COLUMN code VARCHAR(50) UNIQUE,
ADD COLUMN source VARCHAR(20) DEFAULT 'online',
ADD COLUMN paid_status VARCHAR(20) DEFAULT 'unpaid',
ADD COLUMN discount DECIMAL(10,2) DEFAULT 0,
ADD COLUMN shipping_fee DECIMAL(10,2) DEFAULT 0,
ADD COLUMN tax DECIMAL(10,2) DEFAULT 0,
ADD COLUMN tax_enabled BOOLEAN DEFAULT true,
ADD COLUMN payment_method VARCHAR(50),
ADD COLUMN promo_code VARCHAR(50),
ADD COLUMN confirmed_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN cancelled_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN cancelled_reason TEXT;

-- Add check constraints for new columns
ALTER TABLE orders 
ADD CONSTRAINT chk_source CHECK (source IN ('online', 'offline', 'mobile', 'api', 'pos')),
ADD CONSTRAINT chk_paid_status CHECK (paid_status IN ('unpaid', 'partial', 'paid', 'refunded', 'cancelled')),
ADD CONSTRAINT chk_discount CHECK (discount >= 0),
ADD CONSTRAINT chk_shipping_fee CHECK (shipping_fee >= 0),
ADD CONSTRAINT chk_tax CHECK (tax >= 0),
ADD CONSTRAINT chk_payment_method CHECK (payment_method IN ('cash', 'credit_card', 'bank_transfer', 'qr_code', 'wallet', 'installment'));

-- Create index for order code (for faster lookups)
CREATE INDEX idx_orders_code ON orders(code);
CREATE INDEX idx_orders_source ON orders(source);
CREATE INDEX idx_orders_paid_status ON orders(paid_status);
CREATE INDEX idx_orders_confirmed_at ON orders(confirmed_at);
CREATE INDEX idx_orders_cancelled_at ON orders(cancelled_at);

-- Create composite indexes for common queries
CREATE INDEX idx_orders_status_created ON orders(status, created_at);
CREATE INDEX idx_orders_source_status ON orders(source, status);
CREATE INDEX idx_orders_paid_status_total ON orders(paid_status, total_amount);

-- Add comments for documentation
COMMENT ON TABLE order_audit_log IS 'Audit log table for tracking all order changes and actions';
COMMENT ON TABLE order_events_outbox IS 'Outbox pattern table for reliable event publishing';
COMMENT ON COLUMN orders.code IS 'Unique order code for customer reference (e.g., ORD20250001)';
COMMENT ON COLUMN orders.source IS 'Order source channel (online, offline, mobile, api, pos)';
COMMENT ON COLUMN orders.paid_status IS 'Payment status independent of order status';
COMMENT ON COLUMN orders.discount IS 'Total discount amount applied to the order';
COMMENT ON COLUMN orders.shipping_fee IS 'Shipping/delivery fee for the order';
COMMENT ON COLUMN orders.tax IS 'Tax amount calculated for the order';
COMMENT ON COLUMN orders.tax_enabled IS 'Whether tax calculation is enabled for this order';
COMMENT ON COLUMN orders.payment_method IS 'Payment method used for the order';
COMMENT ON COLUMN orders.promo_code IS 'Promotional code applied to the order';
COMMENT ON COLUMN orders.confirmed_at IS 'Timestamp when order was confirmed';
COMMENT ON COLUMN orders.cancelled_at IS 'Timestamp when order was cancelled';
COMMENT ON COLUMN orders.cancelled_reason IS 'Reason for order cancellation';
