-- Create customers table
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE NOT NULL,
    date_of_birth DATE,
    gender VARCHAR(10),
    tier VARCHAR(20) NOT NULL DEFAULT 'bronze' CHECK (tier IN ('bronze', 'silver', 'gold', 'platinum', 'diamond')),
    loyverse_id VARCHAR(100) UNIQUE,
    total_spent DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    order_count INTEGER NOT NULL DEFAULT 0,
    last_order_date TIMESTAMP WITH TIME ZONE,
    delivery_route_id UUID,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create delivery_routes table
CREATE TABLE IF NOT EXISTS delivery_routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create thai_addresses table
CREATE TABLE IF NOT EXISTS thai_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    province VARCHAR(100) NOT NULL,
    district VARCHAR(100) NOT NULL,
    subdistrict VARCHAR(100) NOT NULL,
    postal_code VARCHAR(10) NOT NULL,
    province_code VARCHAR(10) NOT NULL,
    district_code VARCHAR(10) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(province, district, subdistrict, postal_code)
);

-- Create customer_addresses table
CREATE TABLE IF NOT EXISTS customer_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('home', 'work', 'billing', 'shipping')),
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    thai_address_id UUID REFERENCES thai_addresses(id),
    postal_code VARCHAR(10) NOT NULL,
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    is_default BOOLEAN NOT NULL DEFAULT false,
    delivery_notes TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add foreign key constraint for delivery_route_id
ALTER TABLE customers 
ADD CONSTRAINT fk_customers_delivery_route 
FOREIGN KEY (delivery_route_id) REFERENCES delivery_routes(id);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);
CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone);
CREATE INDEX IF NOT EXISTS idx_customers_loyverse_id ON customers(loyverse_id);
CREATE INDEX IF NOT EXISTS idx_customers_tier ON customers(tier);
CREATE INDEX IF NOT EXISTS idx_customers_is_active ON customers(is_active);
CREATE INDEX IF NOT EXISTS idx_customers_delivery_route_id ON customers(delivery_route_id);
CREATE INDEX IF NOT EXISTS idx_customers_created_at ON customers(created_at);
CREATE INDEX IF NOT EXISTS idx_customers_total_spent ON customers(total_spent);

CREATE INDEX IF NOT EXISTS idx_customer_addresses_customer_id ON customer_addresses(customer_id);
CREATE INDEX IF NOT EXISTS idx_customer_addresses_type ON customer_addresses(type);
CREATE INDEX IF NOT EXISTS idx_customer_addresses_is_default ON customer_addresses(is_default);
CREATE INDEX IF NOT EXISTS idx_customer_addresses_thai_address_id ON customer_addresses(thai_address_id);
CREATE INDEX IF NOT EXISTS idx_customer_addresses_postal_code ON customer_addresses(postal_code);

CREATE INDEX IF NOT EXISTS idx_thai_addresses_postal_code ON thai_addresses(postal_code);
CREATE INDEX IF NOT EXISTS idx_thai_addresses_province ON thai_addresses(province);
CREATE INDEX IF NOT EXISTS idx_thai_addresses_district ON thai_addresses(district);
CREATE INDEX IF NOT EXISTS idx_thai_addresses_subdistrict ON thai_addresses(subdistrict);

CREATE INDEX IF NOT EXISTS idx_delivery_routes_name ON delivery_routes(name);
CREATE INDEX IF NOT EXISTS idx_delivery_routes_is_active ON delivery_routes(is_active);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_customers_updated_at BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_customer_addresses_updated_at BEFORE UPDATE ON customer_addresses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_thai_addresses_updated_at BEFORE UPDATE ON thai_addresses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_delivery_routes_updated_at BEFORE UPDATE ON delivery_routes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Ensure only one default address per customer per type
CREATE UNIQUE INDEX IF NOT EXISTS idx_customer_addresses_default_unique 
ON customer_addresses(customer_id, type) 
WHERE is_default = true;
