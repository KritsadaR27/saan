-- Seed data for Finance Service
-- Migration: 002_seed_default_allocation_rules.up.sql

-- Insert default profit allocation rules for different entity types

-- Global default rule (applies when no specific entity rule exists)
INSERT INTO profit_allocation_rules (
    id,
    branch_id,
    vehicle_id,
    profit_percentage,
    owner_pay_percentage,
    tax_percentage,
    effective_from,
    is_active,
    updated_by_user_id,
    created_at,
    updated_at
) VALUES (
    uuid_generate_v4(),
    NULL,
    NULL,
    5.00,   -- 5% profit
    50.00,  -- 50% owner pay
    15.00,  -- 15% tax
    CURRENT_DATE,
    TRUE,
    '00000000-0000-0000-0000-000000000001', -- System user
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
) ON CONFLICT DO NOTHING;

-- Example branch-specific rule (can be customized per branch)
-- INSERT INTO profit_allocation_rules (
--     id,
--     branch_id,
--     vehicle_id,
--     profit_percentage,
--     owner_pay_percentage,
--     tax_percentage,
--     effective_from,
--     is_active,
--     updated_by_user_id,
--     created_at,
--     updated_at
-- ) VALUES (
--     uuid_generate_v4(),
--     '11111111-1111-1111-1111-111111111111', -- Example branch ID
--     NULL,
--     7.00,   -- 7% profit for this branch
--     48.00,  -- 48% owner pay
--     15.00,  -- 15% tax
--     CURRENT_DATE,
--     TRUE,
--     '00000000-0000-0000-0000-000000000001',
--     CURRENT_TIMESTAMP,
--     CURRENT_TIMESTAMP
-- ) ON CONFLICT DO NOTHING;

-- Example vehicle-specific rule (can be customized per vehicle)
-- INSERT INTO profit_allocation_rules (
--     id,
--     branch_id,
--     vehicle_id,
--     profit_percentage,
--     owner_pay_percentage,
--     tax_percentage,
--     effective_from,
--     is_active,
--     updated_by_user_id,
--     created_at,
--     updated_at
-- ) VALUES (
--     uuid_generate_v4(),
--     NULL,
--     '22222222-2222-2222-2222-222222222222', -- Example vehicle ID
--     NULL,
--     6.00,   -- 6% profit for this vehicle
--     49.00,  -- 49% owner pay
--     15.00,  -- 15% tax
--     CURRENT_DATE,
--     TRUE,
--     '00000000-0000-0000-0000-000000000001',
--     CURRENT_TIMESTAMP,
--     CURRENT_TIMESTAMP
-- ) ON CONFLICT DO NOTHING;
