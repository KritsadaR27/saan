-- Rollback schema upgrade
-- Drop new tables
DROP TABLE IF EXISTS customer_points_transactions;
DROP TABLE IF EXISTS vip_tier_benefits;

-- Drop sequence
DROP SEQUENCE IF EXISTS customer_code_seq;

-- Remove new columns from customers table
ALTER TABLE customers 
DROP COLUMN IF EXISTS customer_code,
DROP COLUMN IF EXISTS points_balance,
DROP COLUMN IF EXISTS tier_achieved_date,
DROP COLUMN IF EXISTS loyverse_total_visits,
DROP COLUMN IF EXISTS loyverse_total_spent,
DROP COLUMN IF EXISTS loyverse_points,
DROP COLUMN IF EXISTS first_visit,
DROP COLUMN IF EXISTS last_visit,
DROP COLUMN IF EXISTS last_sync_at,
DROP COLUMN IF EXISTS line_user_id,
DROP COLUMN IF EXISTS line_display_name,
DROP COLUMN IF EXISTS digital_card_issued_at,
DROP COLUMN IF EXISTS last_card_scan,
DROP COLUMN IF EXISTS average_order_value,
DROP COLUMN IF EXISTS purchase_frequency;

-- Revert tier column to VARCHAR
ALTER TABLE customers 
ALTER COLUMN tier TYPE VARCHAR(20) USING (
    CASE tier
        WHEN 1 THEN 'bronze'
        WHEN 2 THEN 'silver'
        WHEN 3 THEN 'gold'
        WHEN 4 THEN 'platinum'
        WHEN 5 THEN 'diamond'
        ELSE 'bronze'
    END
);

-- Remove new columns from thai_addresses
ALTER TABLE thai_addresses 
DROP COLUMN IF EXISTS is_self_delivery_area,
DROP COLUMN IF EXISTS delivery_route;

-- Remove new columns from customer_addresses
ALTER TABLE customer_addresses 
DROP COLUMN IF EXISTS house_number,
DROP COLUMN IF EXISTS delivery_route,
DROP COLUMN IF EXISTS label;

-- Drop new indexes
DROP INDEX IF EXISTS idx_customers_code;
DROP INDEX IF EXISTS idx_customers_tier_active;
DROP INDEX IF EXISTS idx_points_customer;
DROP INDEX IF EXISTS idx_points_reference;
DROP INDEX IF EXISTS idx_points_loyverse;
DROP INDEX IF EXISTS idx_thai_subdistrict;
DROP INDEX IF EXISTS idx_thai_postal;
DROP INDEX IF EXISTS idx_addresses_route;
