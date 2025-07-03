-- SaaN Inventory Service - Complete Database Schema
-- Migration: 002_create_inventory_service_tables.sql
-- Date: 2025-07-02

-- ================================================================
-- INVENTORY SERVICE TABLES
-- ================================================================

-- 1. Product Categories
CREATE TABLE product_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_category_id UUID REFERENCES product_categories(id),
    
    -- Display Options
    display_order INT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    image_url TEXT,
    
    -- SEO & Marketing
    slug VARCHAR(100) UNIQUE,
    meta_description TEXT,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 2. Suppliers
CREATE TABLE suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    contact_person VARCHAR(100),
    email VARCHAR(100),
    phone VARCHAR(20),
    
    -- Address
    address_line1 TEXT,
    address_line2 TEXT,
    subdistrict VARCHAR(100),
    district VARCHAR(100),
    province VARCHAR(100),
    postal_code VARCHAR(10),
    
    -- Business Info
    tax_id VARCHAR(50),
    business_type VARCHAR(50), -- 'manufacturer', 'distributor', 'wholesaler'
    payment_terms VARCHAR(100), -- 'NET 30', 'COD', 'Credit'
    credit_limit DECIMAL(12,2),
    
    -- Performance Metrics
    total_orders INT DEFAULT 0,
    total_spent DECIMAL(12,2) DEFAULT 0,
    average_delivery_days INT,
    quality_rating DECIMAL(3,2) DEFAULT 5.00,
    
    -- Status
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'inactive', 'blocked'
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 3. Products
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES product_categories(id),
    supplier_id UUID REFERENCES suppliers(id),
    
    -- Product Details
    brand VARCHAR(100),
    model VARCHAR(100),
    barcode VARCHAR(50),
    weight DECIMAL(8,3), -- กิโลกรัม
    dimensions JSONB, -- {"length": 10, "width": 5, "height": 3}
    
    -- Pricing
    cost_price DECIMAL(10,2) NOT NULL,
    selling_price DECIMAL(10,2) NOT NULL,
    wholesale_price DECIMAL(10,2),
    markup_percentage DECIMAL(5,2),
    
    -- Inventory Settings
    unit_of_measure VARCHAR(20), -- 'piece', 'box', 'kg', 'liter'
    min_stock_level INT DEFAULT 0,
    max_stock_level INT,
    reorder_point INT DEFAULT 0,
    safety_stock INT DEFAULT 0,
    
    -- Product Status
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'inactive', 'discontinued'
    is_serialized BOOLEAN DEFAULT false,
    requires_expiry_tracking BOOLEAN DEFAULT false,
    
    -- Images & Media
    primary_image_url TEXT,
    gallery_images JSONB,
    
    -- SEO & Marketing
    tags JSONB, -- ["electronics", "smartphone", "mobile"]
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 4. Product Variants (for products with size/color variations)
CREATE TABLE product_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id),
    variant_sku VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    
    -- Variant Attributes
    attributes JSONB, -- {"size": "L", "color": "Red", "material": "Cotton"}
    
    -- Pricing (can override parent product)
    cost_price DECIMAL(10,2),
    selling_price DECIMAL(10,2),
    
    -- Inventory
    current_stock INT DEFAULT 0,
    reserved_stock INT DEFAULT 0,
    available_stock INT GENERATED ALWAYS AS (current_stock - reserved_stock) STORED,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 5. Inventory Movements
CREATE TABLE inventory_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id),
    product_variant_id UUID REFERENCES product_variants(id),
    
    -- Movement Details
    movement_type VARCHAR(20) NOT NULL, -- 'in', 'out', 'adjustment', 'transfer'
    movement_reason VARCHAR(50), -- 'purchase', 'sale', 'return', 'damage', 'expired', 'count_adjustment'
    quantity INT NOT NULL,
    unit_cost DECIMAL(10,2),
    total_cost DECIMAL(12,2),
    
    -- Stock Levels (before and after)
    stock_before INT,
    stock_after INT,
    
    -- Reference Documents
    reference_type VARCHAR(20), -- 'purchase_order', 'sales_order', 'adjustment', 'transfer'
    reference_id UUID,
    reference_number VARCHAR(50),
    
    -- Location & Batch
    warehouse_location VARCHAR(100),
    batch_number VARCHAR(50),
    expiry_date DATE,
    serial_numbers JSONB, -- for serialized products
    
    -- User & Approval
    created_by_user_id UUID,
    approved_by_user_id UUID,
    approved_at TIMESTAMP,
    
    -- Notes
    notes TEXT,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 6. Purchase Orders
CREATE TABLE purchase_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    po_number VARCHAR(20) UNIQUE NOT NULL, -- PO-20250702-001
    supplier_id UUID REFERENCES suppliers(id),
    
    -- Order Details
    status VARCHAR(20) DEFAULT 'draft', -- 'draft', 'sent', 'confirmed', 'partial', 'completed', 'cancelled'
    order_date DATE NOT NULL,
    expected_delivery_date DATE,
    actual_delivery_date DATE,
    
    -- Financial
    subtotal DECIMAL(12,2) DEFAULT 0,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    shipping_cost DECIMAL(10,2) DEFAULT 0,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    total_amount DECIMAL(12,2) DEFAULT 0,
    
    -- Payment
    payment_status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'partial', 'paid'
    payment_terms VARCHAR(100),
    payment_due_date DATE,
    
    -- Delivery
    delivery_address TEXT,
    delivery_notes TEXT,
    
    -- Staff
    created_by_user_id UUID,
    approved_by_user_id UUID,
    received_by_user_id UUID,
    
    -- Timestamps
    approved_at TIMESTAMP,
    received_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 7. Purchase Order Items
CREATE TABLE purchase_order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_order_id UUID REFERENCES purchase_orders(id),
    product_id UUID REFERENCES products(id),
    product_variant_id UUID REFERENCES product_variants(id),
    
    -- Order Details
    quantity_ordered INT NOT NULL,
    quantity_received INT DEFAULT 0,
    quantity_remaining INT GENERATED ALWAYS AS (quantity_ordered - quantity_received) STORED,
    
    -- Pricing
    unit_cost DECIMAL(10,2) NOT NULL,
    total_cost DECIMAL(12,2) GENERATED ALWAYS AS (quantity_ordered * unit_cost) STORED,
    
    -- Delivery Tracking
    expected_date DATE,
    actual_date DATE,
    
    -- Quality Control
    quality_check_status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'passed', 'failed', 'partial'
    quality_notes TEXT,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 8. Stock Adjustments
CREATE TABLE stock_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    adjustment_number VARCHAR(20) UNIQUE NOT NULL, -- ADJ-20250702-001
    
    -- Adjustment Details
    adjustment_type VARCHAR(20) NOT NULL, -- 'count', 'damage', 'expired', 'theft', 'gift'
    reason TEXT NOT NULL,
    adjustment_date DATE NOT NULL,
    
    -- Financial Impact
    total_cost_impact DECIMAL(12,2) DEFAULT 0,
    
    -- Approval
    status VARCHAR(20) DEFAULT 'draft', -- 'draft', 'submitted', 'approved', 'rejected'
    created_by_user_id UUID,
    approved_by_user_id UUID,
    approved_at TIMESTAMP,
    
    -- Notes
    notes TEXT,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 9. Stock Adjustment Items
CREATE TABLE stock_adjustment_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stock_adjustment_id UUID REFERENCES stock_adjustments(id),
    product_id UUID REFERENCES products(id),
    product_variant_id UUID REFERENCES product_variants(id),
    
    -- Adjustment Details
    quantity_counted INT,
    quantity_system INT,
    quantity_difference INT GENERATED ALWAYS AS (quantity_counted - quantity_system) STORED,
    
    -- Financial Impact
    unit_cost DECIMAL(10,2),
    cost_impact DECIMAL(12,2) GENERATED ALWAYS AS (quantity_difference * unit_cost) STORED,
    
    -- Notes
    notes TEXT,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 10. Low Stock Alerts
CREATE TABLE low_stock_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id),
    product_variant_id UUID REFERENCES product_variants(id),
    
    -- Alert Details
    current_stock INT,
    reorder_point INT,
    suggested_order_quantity INT,
    
    -- Status
    alert_status VARCHAR(20) DEFAULT 'active', -- 'active', 'acknowledged', 'resolved'
    acknowledged_by_user_id UUID,
    acknowledged_at TIMESTAMP,
    
    -- Priority
    priority VARCHAR(20) DEFAULT 'normal', -- 'critical', 'high', 'normal', 'low'
    
    created_at TIMESTAMP DEFAULT NOW(),
    resolved_at TIMESTAMP
);

-- ================================================================
-- INDEXES FOR INVENTORY SERVICE
-- ================================================================

-- Product Categories indexes
CREATE INDEX idx_product_categories_parent ON product_categories(parent_category_id);
CREATE INDEX idx_product_categories_active ON product_categories(is_active);

-- Suppliers indexes
CREATE INDEX idx_suppliers_status ON suppliers(status);
CREATE INDEX idx_suppliers_name ON suppliers(name);

-- Products indexes
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_supplier ON products(supplier_id);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_name ON products(name);

-- Product Variants indexes
CREATE INDEX idx_product_variants_product ON product_variants(product_id);
CREATE INDEX idx_product_variants_sku ON product_variants(variant_sku);
CREATE INDEX idx_product_variants_active ON product_variants(is_active);

-- Inventory Movements indexes
CREATE INDEX idx_inventory_movements_product ON inventory_movements(product_id);
CREATE INDEX idx_inventory_movements_variant ON inventory_movements(product_variant_id);
CREATE INDEX idx_inventory_movements_type ON inventory_movements(movement_type);
CREATE INDEX idx_inventory_movements_date ON inventory_movements(created_at);
CREATE INDEX idx_inventory_movements_reference ON inventory_movements(reference_type, reference_id);

-- Purchase Orders indexes
CREATE INDEX idx_purchase_orders_number ON purchase_orders(po_number);
CREATE INDEX idx_purchase_orders_supplier ON purchase_orders(supplier_id);
CREATE INDEX idx_purchase_orders_status ON purchase_orders(status);
CREATE INDEX idx_purchase_orders_date ON purchase_orders(order_date);

-- Purchase Order Items indexes
CREATE INDEX idx_po_items_order ON purchase_order_items(purchase_order_id);
CREATE INDEX idx_po_items_product ON purchase_order_items(product_id);

-- Stock Adjustments indexes
CREATE INDEX idx_stock_adjustments_number ON stock_adjustments(adjustment_number);
CREATE INDEX idx_stock_adjustments_status ON stock_adjustments(status);
CREATE INDEX idx_stock_adjustments_date ON stock_adjustments(adjustment_date);

-- Stock Adjustment Items indexes
CREATE INDEX idx_adjustment_items_adjustment ON stock_adjustment_items(stock_adjustment_id);
CREATE INDEX idx_adjustment_items_product ON stock_adjustment_items(product_id);

-- Low Stock Alerts indexes
CREATE INDEX idx_low_stock_alerts_product ON low_stock_alerts(product_id);
CREATE INDEX idx_low_stock_alerts_status ON low_stock_alerts(alert_status);
CREATE INDEX idx_low_stock_alerts_priority ON low_stock_alerts(priority);
