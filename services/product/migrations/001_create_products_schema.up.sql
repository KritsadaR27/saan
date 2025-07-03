-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    loyverse_id VARCHAR(255) UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    sku VARCHAR(255) UNIQUE NOT NULL,
    barcode VARCHAR(255) UNIQUE,
    category_id UUID,
    base_price DECIMAL(10,2) NOT NULL,
    unit VARCHAR(50) NOT NULL,
    weight DECIMAL(10,3),
    
    -- Dimensions
    length DECIMAL(10,2),
    width DECIMAL(10,2),
    height DECIMAL(10,2),
    dimension_unit VARCHAR(10),
    
    -- Flags
    is_active BOOLEAN DEFAULT TRUE,
    is_vip_only BOOLEAN DEFAULT FALSE,
    
    -- Tags
    tags TEXT[],
    
    -- Master Data Protection
    data_source_type VARCHAR(50) NOT NULL,
    data_source_id VARCHAR(255),
    last_synced_at TIMESTAMP WITH TIME ZONE,
    is_manual_override BOOLEAN DEFAULT FALSE,
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_by UUID,
    version INTEGER DEFAULT 1
);

-- Create categories table
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    loyverse_id VARCHAR(255) UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    parent_id UUID,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    
    -- Master Data Protection
    data_source_type VARCHAR(50) NOT NULL,
    data_source_id VARCHAR(255),
    last_synced_at TIMESTAMP WITH TIME ZONE,
    is_manual_override BOOLEAN DEFAULT FALSE,
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_by UUID,
    version INTEGER DEFAULT 1
);

-- Create prices table
CREATE TABLE IF NOT EXISTS prices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    price_type VARCHAR(50) NOT NULL, -- 'base', 'vip', 'bulk', 'promotional'
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'THB',
    
    -- Bulk pricing
    min_quantity INTEGER,
    max_quantity INTEGER,
    
    -- VIP pricing
    vip_tier_id UUID,
    
    -- Promotional pricing
    valid_from TIMESTAMP WITH TIME ZONE,
    valid_to TIMESTAMP WITH TIME ZONE,
    promotion_name VARCHAR(255),
    discount_percent DECIMAL(5,2),
    
    -- Conditions
    location_ids UUID[],
    customer_group_ids UUID[],
    
    -- Flags
    is_active BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 0,
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_by UUID,
    version INTEGER DEFAULT 1
);

-- Create inventory table
CREATE TABLE IF NOT EXISTS inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    location_id UUID NOT NULL,
    
    -- Stock levels
    stock_level DECIMAL(10,3) NOT NULL DEFAULT 0,
    reserved_level DECIMAL(10,3) NOT NULL DEFAULT 0,
    available_level DECIMAL(10,3) NOT NULL DEFAULT 0,
    
    -- Thresholds
    low_stock_threshold DECIMAL(10,3),
    reorder_point DECIMAL(10,3),
    max_stock_level DECIMAL(10,3),
    
    -- Availability
    is_available BOOLEAN DEFAULT TRUE,
    availability_notes TEXT,
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    version INTEGER DEFAULT 1,
    
    -- Unique constraint for product-location combination
    UNIQUE(product_id, location_id)
);

-- Create foreign key constraints
ALTER TABLE products 
ADD CONSTRAINT fk_products_category 
FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;

ALTER TABLE categories 
ADD CONSTRAINT fk_categories_parent 
FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE SET NULL;

ALTER TABLE prices 
ADD CONSTRAINT fk_prices_product 
FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE;

ALTER TABLE inventory 
ADD CONSTRAINT fk_inventory_product 
FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_products_loyverse_id ON products(loyverse_id);
CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
CREATE INDEX IF NOT EXISTS idx_products_barcode ON products(barcode);
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_is_active ON products(is_active);
CREATE INDEX IF NOT EXISTS idx_products_is_vip_only ON products(is_vip_only);
CREATE INDEX IF NOT EXISTS idx_products_data_source ON products(data_source_type, data_source_id);
CREATE INDEX IF NOT EXISTS idx_products_last_synced ON products(last_synced_at);
CREATE INDEX IF NOT EXISTS idx_products_tags ON products USING GIN(tags);

CREATE INDEX IF NOT EXISTS idx_categories_loyverse_id ON categories(loyverse_id);
CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_categories_is_active ON categories(is_active);
CREATE INDEX IF NOT EXISTS idx_categories_data_source ON categories(data_source_type, data_source_id);
CREATE INDEX IF NOT EXISTS idx_categories_sort_order ON categories(sort_order);

CREATE INDEX IF NOT EXISTS idx_prices_product_id ON prices(product_id);
CREATE INDEX IF NOT EXISTS idx_prices_price_type ON prices(price_type);
CREATE INDEX IF NOT EXISTS idx_prices_is_active ON prices(is_active);
CREATE INDEX IF NOT EXISTS idx_prices_valid_from ON prices(valid_from);
CREATE INDEX IF NOT EXISTS idx_prices_valid_to ON prices(valid_to);
CREATE INDEX IF NOT EXISTS idx_prices_vip_tier_id ON prices(vip_tier_id);
CREATE INDEX IF NOT EXISTS idx_prices_location_ids ON prices USING GIN(location_ids);
CREATE INDEX IF NOT EXISTS idx_prices_customer_group_ids ON prices USING GIN(customer_group_ids);
CREATE INDEX IF NOT EXISTS idx_prices_priority ON prices(priority);

CREATE INDEX IF NOT EXISTS idx_inventory_product_id ON inventory(product_id);
CREATE INDEX IF NOT EXISTS idx_inventory_location_id ON inventory(location_id);
CREATE INDEX IF NOT EXISTS idx_inventory_is_available ON inventory(is_available);
CREATE INDEX IF NOT EXISTS idx_inventory_stock_level ON inventory(stock_level);
CREATE INDEX IF NOT EXISTS idx_inventory_available_level ON inventory(available_level);
CREATE INDEX IF NOT EXISTS idx_inventory_low_stock ON inventory(low_stock_threshold) WHERE low_stock_threshold IS NOT NULL;

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update updated_at
CREATE TRIGGER update_products_updated_at 
    BEFORE UPDATE ON products 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_categories_updated_at 
    BEFORE UPDATE ON categories 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_prices_updated_at 
    BEFORE UPDATE ON prices 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_inventory_updated_at 
    BEFORE UPDATE ON inventory 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create function to update available_level when stock_level or reserved_level changes
CREATE OR REPLACE FUNCTION update_available_level()
RETURNS TRIGGER AS $$
BEGIN
    NEW.available_level = NEW.stock_level - NEW.reserved_level;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update available_level
CREATE TRIGGER update_inventory_available_level 
    BEFORE INSERT OR UPDATE ON inventory 
    FOR EACH ROW EXECUTE FUNCTION update_available_level();

-- Create function to increment version on update
CREATE OR REPLACE FUNCTION increment_version()
RETURNS TRIGGER AS $$
BEGIN
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically increment version
CREATE TRIGGER update_products_version 
    BEFORE UPDATE ON products 
    FOR EACH ROW EXECUTE FUNCTION increment_version();

CREATE TRIGGER update_categories_version 
    BEFORE UPDATE ON categories 
    FOR EACH ROW EXECUTE FUNCTION increment_version();

CREATE TRIGGER update_prices_version 
    BEFORE UPDATE ON prices 
    FOR EACH ROW EXECUTE FUNCTION increment_version();

CREATE TRIGGER update_inventory_version 
    BEFORE UPDATE ON inventory 
    FOR EACH ROW EXECUTE FUNCTION increment_version();
