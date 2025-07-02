-- Drop indexes
DROP INDEX IF EXISTS idx_customer_addresses_default_unique;
DROP INDEX IF EXISTS idx_delivery_routes_is_active;
DROP INDEX IF EXISTS idx_delivery_routes_name;
DROP INDEX IF EXISTS idx_thai_addresses_subdistrict;
DROP INDEX IF EXISTS idx_thai_addresses_district;
DROP INDEX IF EXISTS idx_thai_addresses_province;
DROP INDEX IF EXISTS idx_thai_addresses_postal_code;
DROP INDEX IF EXISTS idx_customer_addresses_postal_code;
DROP INDEX IF EXISTS idx_customer_addresses_thai_address_id;
DROP INDEX IF EXISTS idx_customer_addresses_is_default;
DROP INDEX IF EXISTS idx_customer_addresses_type;
DROP INDEX IF EXISTS idx_customer_addresses_customer_id;
DROP INDEX IF EXISTS idx_customers_total_spent;
DROP INDEX IF EXISTS idx_customers_created_at;
DROP INDEX IF EXISTS idx_customers_delivery_route_id;
DROP INDEX IF EXISTS idx_customers_is_active;
DROP INDEX IF EXISTS idx_customers_tier;
DROP INDEX IF EXISTS idx_customers_loyverse_id;
DROP INDEX IF EXISTS idx_customers_phone;
DROP INDEX IF EXISTS idx_customers_email;

-- Drop triggers
DROP TRIGGER IF EXISTS update_delivery_routes_updated_at ON delivery_routes;
DROP TRIGGER IF EXISTS update_thai_addresses_updated_at ON thai_addresses;
DROP TRIGGER IF EXISTS update_customer_addresses_updated_at ON customer_addresses;
DROP TRIGGER IF EXISTS update_customers_updated_at ON customers;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop foreign key constraints
ALTER TABLE customers DROP CONSTRAINT IF EXISTS fk_customers_delivery_route;

-- Drop tables in reverse order
DROP TABLE IF EXISTS customer_addresses;
DROP TABLE IF EXISTS thai_addresses;
DROP TABLE IF EXISTS delivery_routes;
DROP TABLE IF EXISTS customers;
