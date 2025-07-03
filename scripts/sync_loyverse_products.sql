-- Sync Products from Loyverse Integration to SaaN Database
-- This script creates a stored procedure to sync Loyverse products

-- First, ensure product_categories table has the required data
INSERT INTO product_categories (id, name, description, created_at, updated_at) 
VALUES 
    ('355ffd25-ea53-4d68-825c-bd88281cff1e', 'อาหารแช่แข็ง', 'สินค้าอาหารแช่แข็ง', NOW(), NOW()),
    ('default-category-id', 'ทั่วไป', 'หมวดหมู่ทั่วไป', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Sample sync function for demonstration
-- In production, this would be called by the Loyverse integration service
CREATE OR REPLACE FUNCTION sync_loyverse_product(
    p_external_id VARCHAR,
    p_name VARCHAR,
    p_description TEXT,
    p_category_id UUID,
    p_sku VARCHAR,
    p_barcode VARCHAR,
    p_cost_price DECIMAL,
    p_selling_price DECIMAL
) RETURNS UUID AS $$
DECLARE
    product_id UUID;
BEGIN
    -- Insert or update product
    INSERT INTO products (
        external_id,
        source_system,
        name,
        description,
        category_id,
        sku,
        barcode,
        cost_price,
        selling_price,
        status,
        created_at,
        updated_at
    ) VALUES (
        p_external_id,
        'loyverse',
        p_name,
        p_description,
        COALESCE(p_category_id, (SELECT id FROM product_categories WHERE name = 'ทั่วไป' LIMIT 1)),
        COALESCE(p_sku, p_external_id),
        p_barcode,
        COALESCE(p_cost_price, 0),
        COALESCE(p_selling_price, 0),
        'active',
        NOW(),
        NOW()
    )
    ON CONFLICT (external_id, source_system) 
    DO UPDATE SET
        name = EXCLUDED.name,
        description = EXCLUDED.description,
        category_id = EXCLUDED.category_id,
        sku = EXCLUDED.sku,
        barcode = EXCLUDED.barcode,
        cost_price = EXCLUDED.cost_price,
        selling_price = EXCLUDED.selling_price,
        updated_at = NOW()
    RETURNING id INTO product_id;
    
    RETURN product_id;
END;
$$ LANGUAGE plpgsql;

-- Example usage (this would be called by Loyverse integration):
/*
SELECT sync_loyverse_product(
    '84d70ff3-ea81-4382-a3f9-0448d1f2c339',
    'P822 น้ำจิ้มขนมจีบ',
    '',
    '355ffd25-ea53-4d68-825c-bd88281cff1e',
    'P822',
    '',
    0.00,
    15.00
);
*/

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_products_loyverse_external 
ON products(external_id) WHERE source_system = 'loyverse';

-- Display current sync status
SELECT 
    COUNT(*) as total_products,
    COUNT(*) FILTER (WHERE source_system = 'loyverse') as loyverse_products,
    COUNT(*) FILTER (WHERE source_system = 'internal') as internal_products
FROM products;

NOTICE 'Product sync function created successfully!';
NOTICE 'Use sync_loyverse_product() to sync individual products from Loyverse';
NOTICE 'Products table is ready for external system integration';
