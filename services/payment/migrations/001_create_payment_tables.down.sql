-- Migration: 001_create_payment_tables.down.sql
-- Rollback payment service database schema

-- Drop views
DROP VIEW IF EXISTS order_payment_summary;
DROP VIEW IF EXISTS store_workload;

-- Drop triggers
DROP TRIGGER IF EXISTS update_payment_transactions_updated_at ON payment_transactions;
DROP TRIGGER IF EXISTS update_loyverse_stores_updated_at ON loyverse_stores;
DROP TRIGGER IF EXISTS update_payment_delivery_contexts_updated_at ON payment_delivery_contexts;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes explicitly (though they would be dropped with tables)
DROP INDEX IF EXISTS idx_payment_events_customer_id;
DROP INDEX IF EXISTS idx_payment_events_order_id;
DROP INDEX IF EXISTS idx_payment_events_occurred_at;
DROP INDEX IF EXISTS idx_payment_events_event_type;
DROP INDEX IF EXISTS idx_payment_events_payment_id;

DROP INDEX IF EXISTS idx_payment_delivery_contexts_created_at;
DROP INDEX IF EXISTS idx_payment_delivery_contexts_delivery_status;
DROP INDEX IF EXISTS idx_payment_delivery_contexts_driver_id;
DROP INDEX IF EXISTS idx_payment_delivery_contexts_delivery_id;

DROP INDEX IF EXISTS idx_loyverse_stores_is_active;
DROP INDEX IF EXISTS idx_loyverse_stores_manager_id;
DROP INDEX IF EXISTS idx_loyverse_stores_region;
DROP INDEX IF EXISTS idx_loyverse_stores_store_code;

DROP INDEX IF EXISTS idx_payment_transactions_order_status;
DROP INDEX IF EXISTS idx_payment_transactions_customer_date;
DROP INDEX IF EXISTS idx_payment_transactions_store_date;
DROP INDEX IF EXISTS idx_payment_transactions_loyverse_receipt_id;
DROP INDEX IF EXISTS idx_payment_transactions_created_at;
DROP INDEX IF EXISTS idx_payment_transactions_payment_timing;
DROP INDEX IF EXISTS idx_payment_transactions_payment_channel;
DROP INDEX IF EXISTS idx_payment_transactions_assigned_store_id;
DROP INDEX IF EXISTS idx_payment_transactions_status;
DROP INDEX IF EXISTS idx_payment_transactions_customer_id;
DROP INDEX IF EXISTS idx_payment_transactions_order_id;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS payment_events;
DROP TABLE IF EXISTS payment_delivery_contexts;
DROP TABLE IF EXISTS loyverse_stores;
DROP TABLE IF EXISTS payment_transactions;
