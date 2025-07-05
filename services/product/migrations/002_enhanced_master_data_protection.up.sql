-- Enhanced Products Table with Master Data Protection Pattern
-- Following PRODUCT_SERVICE_IMPLEMENT_PLAN.md

-- Create enhanced products table with proper field separation
CREATE TABLE IF NOT EXISTS products_enhanced (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- ✅ Loyverse-controlled fields (sync updates these)
    external_id VARCHAR(100) UNIQUE,        -- Loyverse ID
    source_system VARCHAR(50) DEFAULT 'loyverse',
    name VARCHAR(200) NOT NULL,
    description TEXT,
    sku VARCHAR(100),
    barcode VARCHAR(100),
    category_id UUID,
    supplier_id UUID,
    cost_price DECIMAL(10,2),
    selling_price DECIMAL(10,2),
    status VARCHAR(20) DEFAULT 'active',
    last_sync_from_loyverse TIMESTAMP WITH TIME ZONE,
    
    -- 🔒 Admin-controlled fields (sync never touches these)
    display_name VARCHAR(200),
    internal_category VARCHAR(100),
    internal_notes TEXT,
    is_featured BOOLEAN DEFAULT false,
    profit_margin_target DECIMAL(5,2),
    sales_tags JSONB,
    
    -- Product Specifications
    weight_grams DECIMAL(8,2),
    units_per_pack INT DEFAULT 1,
    unit_type VARCHAR(20) DEFAULT 'piece',
    
    -- Advanced Availability Control
    is_admin_active BOOLEAN DEFAULT true,
    inactive_reason VARCHAR(200),
    inactive_until TIMESTAMP WITH TIME ZONE,
    auto_reactivate BOOLEAN DEFAULT false,
    inactive_schedule JSONB,
    
    -- VIP Access Control
    vip_only BOOLEAN DEFAULT false,
    min_vip_level VARCHAR(20),
    vip_early_access BOOLEAN DEFAULT false,
    early_access_until TIMESTAMP WITH TIME ZONE,
    
    -- System fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes for performance
    CONSTRAINT products_enhanced_sku_key UNIQUE (sku),
    CONSTRAINT products_enhanced_external_id_key UNIQUE (external_id)
);

-- Product pricing tiers สำหรับ quantity-based pricing
CREATE TABLE IF NOT EXISTS product_pricing_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products_enhanced(id) ON DELETE CASCADE,
    min_quantity INT NOT NULL,
    max_quantity INT,
    price DECIMAL(10,2) NOT NULL,
    tier_name VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    valid_from DATE DEFAULT CURRENT_DATE,
    valid_until DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Customer group pricing
CREATE TABLE IF NOT EXISTS customer_group_pricing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products_enhanced(id) ON DELETE CASCADE,
    customer_group VARCHAR(50) NOT NULL,
    base_price DECIMAL(10,2),
    discount_percentage DECIMAL(5,2),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- VIP pricing benefits
CREATE TABLE IF NOT EXISTS vip_pricing_benefits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vip_level VARCHAR(20) NOT NULL,
    global_discount_percentage DECIMAL(5,2),
    quantity_multiplier DECIMAL(5,2) DEFAULT 1.0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Product availability audit log
CREATE TABLE IF NOT EXISTS product_availability_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products_enhanced(id) ON DELETE CASCADE,
    changed_by UUID,
    change_type VARCHAR(50) NOT NULL, -- 'activated', 'deactivated', 'scheduled', 'auto_reactivated'
    old_status BOOLEAN,
    new_status BOOLEAN,
    reason VARCHAR(200),
    scheduled_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Enhanced indexes for performance
CREATE INDEX IF NOT EXISTS idx_products_enhanced_external_id ON products_enhanced(external_id);
CREATE INDEX IF NOT EXISTS idx_products_enhanced_sku ON products_enhanced(sku);
CREATE INDEX IF NOT EXISTS idx_products_enhanced_name ON products_enhanced USING gin(to_tsvector('english', name));
CREATE INDEX IF NOT EXISTS idx_products_enhanced_status ON products_enhanced(status);
CREATE INDEX IF NOT EXISTS idx_products_enhanced_vip_only ON products_enhanced(vip_only);
CREATE INDEX IF NOT EXISTS idx_products_enhanced_is_admin_active ON products_enhanced(is_admin_active);
CREATE INDEX IF NOT EXISTS idx_products_enhanced_category ON products_enhanced(category_id);
CREATE INDEX IF NOT EXISTS idx_products_enhanced_sales_tags ON products_enhanced USING gin(sales_tags);

CREATE INDEX IF NOT EXISTS idx_pricing_tiers_product_id ON product_pricing_tiers(product_id);
CREATE INDEX IF NOT EXISTS idx_pricing_tiers_quantity ON product_pricing_tiers(min_quantity, max_quantity);
CREATE INDEX IF NOT EXISTS idx_pricing_tiers_active ON product_pricing_tiers(is_active);

CREATE INDEX IF NOT EXISTS idx_customer_group_pricing_product_id ON customer_group_pricing(product_id);
CREATE INDEX IF NOT EXISTS idx_customer_group_pricing_group ON customer_group_pricing(customer_group);

CREATE INDEX IF NOT EXISTS idx_vip_benefits_level ON vip_pricing_benefits(vip_level);
CREATE INDEX IF NOT EXISTS idx_vip_benefits_active ON vip_pricing_benefits(is_active);

-- Triggers for automatic updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_products_enhanced_updated_at BEFORE UPDATE ON products_enhanced FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_pricing_tiers_updated_at BEFORE UPDATE ON product_pricing_tiers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_customer_group_pricing_updated_at BEFORE UPDATE ON customer_group_pricing FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_vip_benefits_updated_at BEFORE UPDATE ON vip_pricing_benefits FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
