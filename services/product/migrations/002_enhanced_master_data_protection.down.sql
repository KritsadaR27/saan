-- Drop enhanced master data protection tables
DROP TRIGGER IF EXISTS update_vip_benefits_updated_at ON vip_pricing_benefits;
DROP TRIGGER IF EXISTS update_customer_group_pricing_updated_at ON customer_group_pricing;
DROP TRIGGER IF EXISTS update_pricing_tiers_updated_at ON product_pricing_tiers;
DROP TRIGGER IF EXISTS update_products_enhanced_updated_at ON products_enhanced;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_vip_benefits_active;
DROP INDEX IF EXISTS idx_vip_benefits_level;
DROP INDEX IF EXISTS idx_customer_group_pricing_group;
DROP INDEX IF EXISTS idx_customer_group_pricing_product_id;
DROP INDEX IF EXISTS idx_pricing_tiers_active;
DROP INDEX IF EXISTS idx_pricing_tiers_quantity;
DROP INDEX IF EXISTS idx_pricing_tiers_product_id;
DROP INDEX IF EXISTS idx_products_enhanced_sales_tags;
DROP INDEX IF EXISTS idx_products_enhanced_category;
DROP INDEX IF EXISTS idx_products_enhanced_is_admin_active;
DROP INDEX IF EXISTS idx_products_enhanced_vip_only;
DROP INDEX IF EXISTS idx_products_enhanced_status;
DROP INDEX IF EXISTS idx_products_enhanced_name;
DROP INDEX IF EXISTS idx_products_enhanced_sku;
DROP INDEX IF EXISTS idx_products_enhanced_external_id;

DROP TABLE IF EXISTS product_availability_log;
DROP TABLE IF EXISTS vip_pricing_benefits;
DROP TABLE IF EXISTS customer_group_pricing;
DROP TABLE IF EXISTS product_pricing_tiers;
DROP TABLE IF EXISTS products_enhanced;
