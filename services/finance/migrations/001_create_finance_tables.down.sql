-- Drop Finance Service Database Schema
-- Migration: 001_create_finance_tables.down.sql

-- Drop triggers first
DROP TRIGGER IF EXISTS update_cash_transfers_updated_at ON cash_transfers;
DROP TRIGGER IF EXISTS update_cash_transfer_batches_updated_at ON cash_transfer_batches;
DROP TRIGGER IF EXISTS update_profit_allocation_rules_updated_at ON profit_allocation_rules;
DROP TRIGGER IF EXISTS update_daily_cash_summaries_updated_at ON daily_cash_summaries;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_cash_flow_type;
DROP INDEX IF EXISTS idx_cash_flow_date;
DROP INDEX IF EXISTS idx_cash_flow_entity;

DROP INDEX IF EXISTS idx_expenses_date;
DROP INDEX IF EXISTS idx_expenses_category;
DROP INDEX IF EXISTS idx_expenses_summary;

DROP INDEX IF EXISTS idx_transfers_type;
DROP INDEX IF EXISTS idx_transfers_status;
DROP INDEX IF EXISTS idx_transfers_batch;

DROP INDEX IF EXISTS idx_transfer_batches_vehicle;
DROP INDEX IF EXISTS idx_transfer_batches_branch;
DROP INDEX IF EXISTS idx_transfer_batches_date;
DROP INDEX IF EXISTS idx_transfer_batches_status;

DROP INDEX IF EXISTS idx_allocation_rules_effective;
DROP INDEX IF EXISTS idx_allocation_rules_vehicle;
DROP INDEX IF EXISTS idx_allocation_rules_branch;
DROP INDEX IF EXISTS idx_allocation_rules_active;

DROP INDEX IF EXISTS idx_daily_summaries_reconciled;
DROP INDEX IF EXISTS idx_daily_summaries_vehicle;
DROP INDEX IF EXISTS idx_daily_summaries_branch;
DROP INDEX IF EXISTS idx_daily_summaries_date;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS cash_flow_records;
DROP TABLE IF EXISTS expense_entries;
DROP TABLE IF EXISTS cash_transfers;
DROP TABLE IF EXISTS cash_transfer_batches;
DROP TABLE IF EXISTS profit_allocation_rules;
DROP TABLE IF EXISTS daily_cash_summaries;
