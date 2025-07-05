-- Remove seed data for Finance Service
-- Migration: 002_seed_default_allocation_rules.down.sql

-- Remove all default allocation rules
DELETE FROM profit_allocation_rules WHERE updated_by_user_id = '00000000-0000-0000-0000-000000000001';
