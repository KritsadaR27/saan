-- Update customers table to match CUSTOMER_SERVICE_IMPLEMENT.md spec
-- Update tier column to use INTEGER instead of VARCHAR
ALTER TABLE customers 
ALTER COLUMN tier TYPE INTEGER USING (
    CASE tier
        WHEN 'bronze' THEN 1
        WHEN 'silver' THEN 2
        WHEN 'gold' THEN 3
        WHEN 'platinum' THEN 4
        WHEN 'diamond' THEN 5
        ELSE 1
    END
);

-- Add new SAAN-specific fields
ALTER TABLE customers 
ADD COLUMN IF NOT EXISTS customer_code VARCHAR(20) UNIQUE,
ADD COLUMN IF NOT EXISTS points_balance INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS tier_achieved_date TIMESTAMP WITH TIME ZONE;

-- Add Loyverse integration fields
ALTER TABLE customers 
ADD COLUMN IF NOT EXISTS loyverse_total_visits INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS loyverse_total_spent DECIMAL(12,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS loyverse_points INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS first_visit TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS last_visit TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS last_sync_at TIMESTAMP WITH TIME ZONE;

-- Add LINE integration fields
ALTER TABLE customers 
ADD COLUMN IF NOT EXISTS line_user_id VARCHAR(100) UNIQUE,
ADD COLUMN IF NOT EXISTS line_display_name VARCHAR(100),
ADD COLUMN IF NOT EXISTS digital_card_issued_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS last_card_scan TIMESTAMP WITH TIME ZONE;

-- Add purchase analytics fields
ALTER TABLE customers 
ADD COLUMN IF NOT EXISTS average_order_value DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS purchase_frequency DECIMAL(5,2);

-- Update tier constraint to use integers
ALTER TABLE customers 
DROP CONSTRAINT IF EXISTS customers_tier_check,
ADD CONSTRAINT customers_tier_check CHECK (tier >= 1 AND tier <= 5);

-- Create VIP tier benefits table
CREATE TABLE IF NOT EXISTS vip_tier_benefits (
    tier INTEGER PRIMARY KEY CHECK (tier >= 1 AND tier <= 5),
    tier_name VARCHAR(20) NOT NULL,
    tier_icon VARCHAR(10) NOT NULL,
    min_spent DECIMAL(12,2) NOT NULL,
    discount_percentage DECIMAL(5,2) DEFAULT 0,
    points_multiplier DECIMAL(5,2) DEFAULT 1.0,
    free_delivery_threshold DECIMAL(10,2),
    early_access BOOLEAN DEFAULT false,
    personal_shopper BOOLEAN DEFAULT false,
    vip_hotline BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create customer points transactions table
CREATE TABLE IF NOT EXISTS customer_points_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID REFERENCES customers(id) ON DELETE CASCADE,
    transaction_type VARCHAR(20) NOT NULL CHECK (transaction_type IN ('earned', 'redeemed', 'expired', 'bonus')),
    points_amount INTEGER NOT NULL,
    reference_id UUID,
    reference_type VARCHAR(50),
    loyverse_receipt_id VARCHAR(100),
    source VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Update thai_addresses table for better address lookup
ALTER TABLE thai_addresses 
ADD COLUMN IF NOT EXISTS is_self_delivery_area BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS delivery_route VARCHAR(50);

-- Create customer code sequence
CREATE SEQUENCE IF NOT EXISTS customer_code_seq START 1;

-- Insert VIP tier configuration
INSERT INTO vip_tier_benefits (tier, tier_name, tier_icon, min_spent, discount_percentage, points_multiplier, free_delivery_threshold, early_access, personal_shopper, vip_hotline) 
VALUES
(1, 'Bronze', '🥉', 0, 0, 1.0, 500, false, false, false),
(2, 'Silver', '🥈', 10000, 3, 1.2, 400, false, false, false),
(3, 'Gold', '🥇', 50000, 5, 1.5, 300, true, false, false),
(4, 'Platinum', '💎', 100000, 8, 2.0, 200, true, true, false),
(5, 'Diamond', '💍', 250000, 12, 3.0, 0, true, true, true)
ON CONFLICT (tier) DO NOTHING;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone);
CREATE INDEX IF NOT EXISTS idx_customers_code ON customers(customer_code);
CREATE INDEX IF NOT EXISTS idx_customers_loyverse ON customers(loyverse_id);
CREATE INDEX IF NOT EXISTS idx_customers_line ON customers(line_user_id);
CREATE INDEX IF NOT EXISTS idx_customers_tier ON customers(tier);
CREATE INDEX IF NOT EXISTS idx_customers_tier_active ON customers(tier, is_active) WHERE is_active = true;

CREATE INDEX IF NOT EXISTS idx_points_customer ON customer_points_transactions(customer_id, created_at);
CREATE INDEX IF NOT EXISTS idx_points_reference ON customer_points_transactions(reference_id, reference_type);
CREATE INDEX IF NOT EXISTS idx_points_loyverse ON customer_points_transactions(loyverse_receipt_id);

CREATE INDEX IF NOT EXISTS idx_thai_subdistrict ON thai_addresses(subdistrict);
CREATE INDEX IF NOT EXISTS idx_thai_province ON thai_addresses(province);
CREATE INDEX IF NOT EXISTS idx_thai_postal ON thai_addresses(postal_code);

-- Update customer_addresses table for better Thai address support
ALTER TABLE customer_addresses 
ADD COLUMN IF NOT EXISTS house_number VARCHAR(20),
ADD COLUMN IF NOT EXISTS delivery_route VARCHAR(50),
ADD COLUMN IF NOT EXISTS label VARCHAR(100),
RENAME COLUMN address_line1 TO address_line1,
RENAME COLUMN sub_district TO subdistrict,
RENAME COLUMN thai_address_id TO thai_address_id;

-- Change column name if needed
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'customer_addresses' AND column_name = 'sub_district') THEN
        ALTER TABLE customer_addresses RENAME COLUMN sub_district TO subdistrict;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_customer_addresses ON customer_addresses(customer_id);
CREATE INDEX IF NOT EXISTS idx_addresses_province ON customer_addresses(province);
CREATE INDEX IF NOT EXISTS idx_addresses_route ON customer_addresses(delivery_route);
