-- Add stock override support to order_items table
-- Migration: 003_add_stock_override_fields.sql

ALTER TABLE order_items 
ADD COLUMN is_override BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN override_reason TEXT;

-- Create index for querying override items
CREATE INDEX idx_order_items_is_override ON order_items(is_override);

-- Add comments for documentation
COMMENT ON COLUMN order_items.is_override IS 'Indicates if this item was added with stock override';
COMMENT ON COLUMN order_items.override_reason IS 'Reason for stock override, required when is_override is true';
