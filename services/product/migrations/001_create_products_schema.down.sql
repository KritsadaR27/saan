-- Drop triggers first
DROP TRIGGER IF EXISTS update_products_version ON products;
DROP TRIGGER IF EXISTS update_categories_version ON categories;
DROP TRIGGER IF EXISTS update_prices_version ON prices;
DROP TRIGGER IF EXISTS update_inventory_version ON inventory;

DROP TRIGGER IF EXISTS update_inventory_available_level ON inventory;

DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;
DROP TRIGGER IF EXISTS update_prices_updated_at ON prices;
DROP TRIGGER IF EXISTS update_inventory_updated_at ON inventory;

-- Drop functions
DROP FUNCTION IF EXISTS increment_version();
DROP FUNCTION IF EXISTS update_available_level();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_products_loyverse_id;
DROP INDEX IF EXISTS idx_products_sku;
DROP INDEX IF EXISTS idx_products_barcode;
DROP INDEX IF EXISTS idx_products_category_id;
DROP INDEX IF EXISTS idx_products_is_active;
DROP INDEX IF EXISTS idx_products_is_vip_only;
DROP INDEX IF EXISTS idx_products_data_source;
DROP INDEX IF EXISTS idx_products_last_synced;
DROP INDEX IF EXISTS idx_products_tags;

DROP INDEX IF EXISTS idx_categories_loyverse_id;
DROP INDEX IF EXISTS idx_categories_parent_id;
DROP INDEX IF EXISTS idx_categories_is_active;
DROP INDEX IF EXISTS idx_categories_data_source;
DROP INDEX IF EXISTS idx_categories_sort_order;

DROP INDEX IF EXISTS idx_prices_product_id;
DROP INDEX IF EXISTS idx_prices_price_type;
DROP INDEX IF EXISTS idx_prices_is_active;
DROP INDEX IF EXISTS idx_prices_valid_from;
DROP INDEX IF EXISTS idx_prices_valid_to;
DROP INDEX IF EXISTS idx_prices_vip_tier_id;
DROP INDEX IF EXISTS idx_prices_location_ids;
DROP INDEX IF EXISTS idx_prices_customer_group_ids;
DROP INDEX IF EXISTS idx_prices_priority;

DROP INDEX IF EXISTS idx_inventory_product_id;
DROP INDEX IF EXISTS idx_inventory_location_id;
DROP INDEX IF EXISTS idx_inventory_is_available;
DROP INDEX IF EXISTS idx_inventory_stock_level;
DROP INDEX IF EXISTS idx_inventory_available_level;
DROP INDEX IF EXISTS idx_inventory_low_stock;

-- Drop foreign key constraints
ALTER TABLE products DROP CONSTRAINT IF EXISTS fk_products_category;
ALTER TABLE categories DROP CONSTRAINT IF EXISTS fk_categories_parent;
ALTER TABLE prices DROP CONSTRAINT IF EXISTS fk_prices_product;
ALTER TABLE inventory DROP CONSTRAINT IF EXISTS fk_inventory_product;

-- Drop tables
DROP TABLE IF EXISTS inventory;
DROP TABLE IF EXISTS prices;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
