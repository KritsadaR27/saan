-- SaaN Enhanced Orders & Order Management
-- Migration: 008_create_enhanced_orders_tables.sql
-- Date: 2025-07-02

-- ================================================================
-- ENHANCED ORDERS & ORDER MANAGEMENT TABLES
-- ================================================================

-- 1. Enhanced Orders Table (Updated from existing)
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number VARCHAR(20) UNIQUE NOT NULL, -- ORD-20250702-001
    customer_id UUID REFERENCES customers(id),
    customer_address_id UUID REFERENCES customer_addresses(id), -- ที่อยู่ส่งของ
    
    -- Order Classification
    order_type VARCHAR(20) DEFAULT 'standard', -- 'standard', 'urgent', 'bulk', 'sample', 'return'
    order_source VARCHAR(20) DEFAULT 'phone', -- 'phone', 'website', 'app', 'walk_in', 'sales_rep'
    
    -- Order Details
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'confirmed', 'processing', 'ready', 'shipped', 'delivered', 'cancelled', 'returned'
    priority VARCHAR(20) DEFAULT 'normal', -- 'low', 'normal', 'high', 'urgent'
    
    -- Financial Summary
    subtotal DECIMAL(12,2) NOT NULL DEFAULT 0,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    delivery_fee DECIMAL(8,2) DEFAULT 0,
    additional_fees DECIMAL(8,2) DEFAULT 0,
    total_amount DECIMAL(12,2) NOT NULL,
    
    -- Payment Information
    payment_status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'partial', 'paid', 'cod', 'failed', 'refunded'
    payment_method VARCHAR(20), -- 'cash', 'transfer', 'qr_code', 'cod', 'credit'
    cod_amount DECIMAL(10,2) DEFAULT 0,
    paid_amount DECIMAL(12,2) DEFAULT 0,
    remaining_amount DECIMAL(12,2) GENERATED ALWAYS AS (total_amount - paid_amount) STORED,
    
    -- Delivery Scheduling
    requested_delivery_date DATE, -- วันที่ลูกค้าต้องการ
    delivery_time_slot VARCHAR(20), -- 'morning', 'afternoon', 'evening', 'anytime'
    delivery_priority VARCHAR(20) DEFAULT 'normal', -- 'urgent', 'normal', 'flexible'
    special_requirements TEXT, -- "ของเปราะบาง", "เก็บเงินปลายทาง"
    
    -- Delivery Assignment & Tracking
    delivery_status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'scheduled', 'assigned', 'picked_up', 'in_transit', 'delivered', 'failed', 'returned', 'rescheduled'
    assigned_delivery_date DATE, -- วันที่ระบบจัดส่ง
    delivery_route_code VARCHAR(10),
    estimated_delivery_time TIMESTAMP,
    actual_delivery_time TIMESTAMP,
    
    -- Customer Information (denormalized for reporting)
    customer_name VARCHAR(200),
    customer_phone VARCHAR(20),
    customer_email VARCHAR(100),
    delivery_address TEXT,
    delivery_notes TEXT,
    
    -- Business Context
    sales_channel VARCHAR(20), -- 'direct', 'online', 'reseller', 'wholesale'
    campaign_code VARCHAR(50), -- Marketing campaign reference
    referral_source VARCHAR(100),
    
    -- Staff Assignment
    created_by_user_id UUID REFERENCES users(id),
    sales_rep_id UUID REFERENCES users(id),
    assigned_to_user_id UUID REFERENCES users(id),
    approved_by_user_id UUID REFERENCES users(id),
    
    -- Important Timestamps
    order_date DATE DEFAULT CURRENT_DATE,
    confirmed_at TIMESTAMP,
    shipped_at TIMESTAMP,
    delivered_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    
    -- Customer Communication
    customer_notes TEXT, -- Notes from customer
    internal_notes TEXT, -- Internal staff notes
    
    -- Quality & Feedback
    customer_rating INT, -- 1-5 stars
    customer_feedback TEXT,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 2. Order Items (Enhanced)
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES orders(id),
    product_id UUID, -- Will reference products table
    product_variant_id UUID, -- Will reference product_variants table
    
    -- Product Information (denormalized for order history)
    sku VARCHAR(50),
    product_name VARCHAR(200) NOT NULL,
    variant_name VARCHAR(200),
    description TEXT,
    
    -- Quantities
    quantity_ordered DECIMAL(10,2) NOT NULL,
    quantity_shipped DECIMAL(10,2) DEFAULT 0,
    quantity_delivered DECIMAL(10,2) DEFAULT 0,
    quantity_returned DECIMAL(10,2) DEFAULT 0,
    quantity_cancelled DECIMAL(10,2) DEFAULT 0,
    
    -- Pricing
    unit_price DECIMAL(10,2) NOT NULL,
    original_price DECIMAL(10,2), -- Before any discounts
    discount_percentage DECIMAL(5,2) DEFAULT 0,
    discount_amount DECIMAL(8,2) DEFAULT 0,
    line_total DECIMAL(12,2) GENERATED ALWAYS AS (quantity_ordered * unit_price - discount_amount) STORED,
    
    -- Tax
    tax_rate DECIMAL(5,2) DEFAULT 7.00,
    tax_amount DECIMAL(8,2),
    
    -- Product Attributes (at time of order)
    weight DECIMAL(8,3),
    dimensions JSONB,
    requires_special_handling BOOLEAN DEFAULT false,
    
    -- Inventory Tracking
    batch_number VARCHAR(50),
    expiry_date DATE,
    serial_numbers JSONB,
    
    -- Delivery Status
    item_status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'picked', 'shipped', 'delivered', 'returned', 'cancelled'
    
    -- Notes
    notes TEXT,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 3. Order Status History
CREATE TABLE order_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES orders(id),
    
    -- Status Change
    from_status VARCHAR(20),
    to_status VARCHAR(20) NOT NULL,
    status_type VARCHAR(20) DEFAULT 'order', -- 'order', 'payment', 'delivery'
    
    -- Change Details
    reason VARCHAR(100),
    notes TEXT,
    
    -- User & System Context
    changed_by_user_id UUID REFERENCES users(id),
    change_source VARCHAR(20) DEFAULT 'manual', -- 'manual', 'system', 'api', 'webhook'
    ip_address VARCHAR(45),
    
    -- Timestamp
    changed_at TIMESTAMP DEFAULT NOW()
);

-- 4. Order Promotions & Discounts Applied
CREATE TABLE order_promotions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES orders(id),
    promotion_id UUID REFERENCES promotions(id),
    
    -- Promotion Details (snapshot at time of application)
    promotion_code VARCHAR(50),
    promotion_name VARCHAR(200),
    promotion_type VARCHAR(20),
    
    -- Discount Applied
    discount_amount DECIMAL(10,2) NOT NULL,
    discount_percentage DECIMAL(5,2),
    
    -- Application Scope
    applies_to VARCHAR(20), -- 'order', 'items', 'delivery'
    affected_items JSONB, -- Order item IDs if item-specific
    
    -- Validation
    conditions_met JSONB, -- Conditions that were satisfied
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 5. Order Documents & Attachments
CREATE TABLE order_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES orders(id),
    
    -- Document Details
    document_type VARCHAR(20), -- 'invoice', 'receipt', 'delivery_note', 'customer_signature', 'photo'
    document_name VARCHAR(200),
    file_url TEXT NOT NULL,
    file_size INT, -- bytes
    mime_type VARCHAR(100),
    
    -- Document Metadata
    description TEXT,
    is_customer_visible BOOLEAN DEFAULT false,
    is_required_for_delivery BOOLEAN DEFAULT false,
    
    -- Upload Context
    uploaded_by_user_id UUID REFERENCES users(id),
    upload_source VARCHAR(20), -- 'admin', 'driver_app', 'customer', 'system'
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 6. Order Delivery Attempts
CREATE TABLE order_delivery_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES orders(id),
    delivery_round_id UUID, -- Will reference delivery_rounds table
    
    -- Attempt Details
    attempt_number INT NOT NULL,
    attempt_date DATE NOT NULL,
    attempt_time TIME,
    
    -- Delivery Assignment
    driver_id UUID, -- Will reference drivers table
    vehicle_id UUID, -- Will reference vehicles table
    
    -- Customer Location
    delivery_address TEXT,
    customer_latitude DECIMAL(10,8),
    customer_longitude DECIMAL(11,8),
    
    -- Attempt Result
    attempt_status VARCHAR(20) NOT NULL, -- 'successful', 'failed', 'rescheduled', 'customer_unavailable'
    completion_time TIMESTAMP,
    
    -- Failure Details
    failure_reason VARCHAR(100),
    failure_category VARCHAR(50), -- 'address_issue', 'customer_unavailable', 'payment_issue', 'product_issue'
    
    -- Customer Interaction
    customer_present BOOLEAN,
    received_by VARCHAR(100), -- Who received the order if not customer
    customer_signature_url TEXT,
    delivery_photo_url TEXT,
    
    -- COD Collection
    cod_amount_due DECIMAL(10,2),
    cod_amount_collected DECIMAL(10,2),
    cod_collection_method VARCHAR(20),
    
    -- Next Steps
    reschedule_requested BOOLEAN DEFAULT false,
    reschedule_date DATE,
    reschedule_notes TEXT,
    
    -- Notes
    driver_notes TEXT,
    customer_feedback TEXT,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 7. Order Returns & Exchanges
CREATE TABLE order_returns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    return_number VARCHAR(20) UNIQUE NOT NULL, -- RET-20250702-001
    original_order_id UUID REFERENCES orders(id),
    customer_id UUID REFERENCES customers(id),
    
    -- Return Type
    return_type VARCHAR(20) NOT NULL, -- 'return', 'exchange', 'refund'
    return_reason VARCHAR(50), -- 'defective', 'wrong_item', 'customer_change_mind', 'damaged_in_transit'
    
    -- Return Details
    return_status VARCHAR(20) DEFAULT 'requested', -- 'requested', 'approved', 'picked_up', 'received', 'processed', 'refunded', 'rejected'
    requested_date DATE DEFAULT CURRENT_DATE,
    approved_date DATE,
    pickup_date DATE,
    received_date DATE,
    processed_date DATE,
    
    -- Financial Impact
    refund_amount DECIMAL(12,2),
    restocking_fee DECIMAL(8,2) DEFAULT 0,
    return_shipping_cost DECIMAL(8,2) DEFAULT 0,
    
    -- Exchange Details (if applicable)
    exchange_order_id UUID REFERENCES orders(id),
    additional_payment_required DECIMAL(10,2) DEFAULT 0,
    
    -- Quality Assessment
    condition_on_return VARCHAR(20), -- 'new', 'like_new', 'good', 'fair', 'poor', 'damaged'
    quality_notes TEXT,
    
    -- Customer Service
    handled_by_user_id UUID REFERENCES users(id),
    customer_satisfaction INT, -- 1-5 rating
    
    -- Notes
    customer_reason TEXT,
    internal_notes TEXT,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 8. Order Return Items
CREATE TABLE order_return_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_return_id UUID REFERENCES order_returns(id),
    order_item_id UUID REFERENCES order_items(id),
    
    -- Return Quantities
    quantity_to_return DECIMAL(10,2) NOT NULL,
    quantity_received DECIMAL(10,2) DEFAULT 0,
    quantity_approved DECIMAL(10,2) DEFAULT 0,
    
    -- Item Condition
    condition_on_return VARCHAR(20),
    defect_description TEXT,
    
    -- Financial
    unit_refund_amount DECIMAL(10,2),
    total_refund_amount DECIMAL(12,2) GENERATED ALWAYS AS (quantity_approved * unit_refund_amount) STORED,
    
    -- Disposition
    disposition VARCHAR(20), -- 'restock', 'scrap', 'repair', 'sell_as_is'
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 9. Order Notifications & Communications
CREATE TABLE order_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES orders(id),
    customer_id UUID REFERENCES customers(id),
    
    -- Notification Details
    notification_type VARCHAR(50), -- 'order_confirmed', 'payment_received', 'shipped', 'delivered', 'cancelled'
    channel VARCHAR(20), -- 'email', 'sms', 'push', 'call', 'line'
    
    -- Content
    subject VARCHAR(200),
    message TEXT,
    
    -- Delivery Status
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'sent', 'delivered', 'failed', 'bounced'
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    opened_at TIMESTAMP,
    clicked_at TIMESTAMP,
    
    -- Error Handling
    failure_reason TEXT,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    
    -- Provider Details
    provider VARCHAR(50), -- Email/SMS service provider
    provider_message_id VARCHAR(100),
    provider_response JSONB,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- ================================================================
-- INDEXES FOR ENHANCED ORDERS MANAGEMENT
-- ================================================================

-- Enhanced Orders indexes
CREATE INDEX idx_orders_number ON orders(order_number);
CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_orders_address ON orders(customer_address_id);
CREATE INDEX idx_orders_status ON orders(status, delivery_status);
CREATE INDEX idx_orders_payment_status ON orders(payment_status);
CREATE INDEX idx_orders_delivery_date ON orders(assigned_delivery_date);
CREATE INDEX idx_orders_route ON orders(delivery_route_code);
CREATE INDEX idx_orders_sales_rep ON orders(sales_rep_id);
CREATE INDEX idx_orders_date ON orders(order_date);
CREATE INDEX idx_orders_total ON orders(total_amount);
CREATE INDEX idx_orders_created_by ON orders(created_by_user_id);
CREATE INDEX idx_orders_source ON orders(order_source);

-- Order Items indexes
CREATE INDEX idx_order_items_order ON order_items(order_id);
CREATE INDEX idx_order_items_product ON order_items(product_id);
CREATE INDEX idx_order_items_variant ON order_items(product_variant_id);
CREATE INDEX idx_order_items_sku ON order_items(sku);
CREATE INDEX idx_order_items_status ON order_items(item_status);

-- Order Status History indexes
CREATE INDEX idx_order_status_history_order ON order_status_history(order_id);
CREATE INDEX idx_order_status_history_status ON order_status_history(to_status);
CREATE INDEX idx_order_status_history_user ON order_status_history(changed_by_user_id);
CREATE INDEX idx_order_status_history_date ON order_status_history(changed_at);

-- Order Promotions indexes
CREATE INDEX idx_order_promotions_order ON order_promotions(order_id);
CREATE INDEX idx_order_promotions_promotion ON order_promotions(promotion_id);

-- Order Documents indexes
CREATE INDEX idx_order_documents_order ON order_documents(order_id);
CREATE INDEX idx_order_documents_type ON order_documents(document_type);
CREATE INDEX idx_order_documents_uploaded_by ON order_documents(uploaded_by_user_id);

-- Order Delivery Attempts indexes
CREATE INDEX idx_delivery_attempts_order ON order_delivery_attempts(order_id);
CREATE INDEX idx_delivery_attempts_round ON order_delivery_attempts(delivery_round_id);
CREATE INDEX idx_delivery_attempts_driver ON order_delivery_attempts(driver_id);
CREATE INDEX idx_delivery_attempts_date ON order_delivery_attempts(attempt_date);
CREATE INDEX idx_delivery_attempts_status ON order_delivery_attempts(attempt_status);

-- Order Returns indexes
CREATE INDEX idx_order_returns_number ON order_returns(return_number);
CREATE INDEX idx_order_returns_original_order ON order_returns(original_order_id);
CREATE INDEX idx_order_returns_customer ON order_returns(customer_id);
CREATE INDEX idx_order_returns_status ON order_returns(return_status);
CREATE INDEX idx_order_returns_date ON order_returns(requested_date);

-- Order Return Items indexes
CREATE INDEX idx_order_return_items_return ON order_return_items(order_return_id);
CREATE INDEX idx_order_return_items_order_item ON order_return_items(order_item_id);

-- Order Notifications indexes
CREATE INDEX idx_order_notifications_order ON order_notifications(order_id);
CREATE INDEX idx_order_notifications_customer ON order_notifications(customer_id);
CREATE INDEX idx_order_notifications_type ON order_notifications(notification_type);
CREATE INDEX idx_order_notifications_status ON order_notifications(status);
CREATE INDEX idx_order_notifications_channel ON order_notifications(channel);
