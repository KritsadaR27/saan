-- SaaN Business Features & Configuration - Complete Database Schema
-- Migration: 006_create_business_features_tables.sql
-- Date: 2025-07-02

-- ================================================================
-- BUSINESS FEATURES & CONFIGURATION TABLES
-- ================================================================

-- 1. Business Settings & Configuration
CREATE TABLE business_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(100) UNIQUE NOT NULL, -- 'company_name', 'tax_rate', 'delivery_fee'
    value TEXT NOT NULL,
    value_type VARCHAR(20) DEFAULT 'string', -- 'string', 'number', 'boolean', 'json'
    
    -- Metadata
    category VARCHAR(50), -- 'company', 'tax', 'delivery', 'payment', 'notification'
    description TEXT,
    is_encrypted BOOLEAN DEFAULT false,
    is_public BOOLEAN DEFAULT false, -- Can be accessed without authentication
    
    -- Validation
    validation_rules JSONB, -- {"min": 0, "max": 100, "required": true}
    
    -- Audit
    updated_by_user_id UUID,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 2. Promotions & Discounts
CREATE TABLE promotions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    code VARCHAR(50) UNIQUE, -- Discount code for customers to use
    
    -- Promotion Type
    type VARCHAR(20) NOT NULL, -- 'percentage', 'fixed_amount', 'free_shipping', 'buy_x_get_y'
    
    -- Discount Rules
    discount_percentage DECIMAL(5,2), -- For percentage discounts
    discount_amount DECIMAL(10,2), -- For fixed amount discounts
    min_order_amount DECIMAL(10,2), -- Minimum order to qualify
    max_discount_amount DECIMAL(10,2), -- Maximum discount cap
    
    -- Buy X Get Y Rules
    buy_quantity INT, -- Buy X items
    get_quantity INT, -- Get Y items free
    buy_product_ids JSONB, -- Specific products for buy condition
    get_product_ids JSONB, -- Specific products for get condition
    
    -- Usage Limits
    max_uses_total INT, -- Total uses across all customers
    max_uses_per_customer INT, -- Per customer limit
    current_uses INT DEFAULT 0,
    
    -- Eligibility
    eligible_customer_tiers JSONB, -- ["gold", "vip"]
    eligible_routes JSONB, -- ["A", "B", "C"]
    eligible_product_categories JSONB,
    
    -- Validity Period
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    is_stackable BOOLEAN DEFAULT false, -- Can be combined with other promotions
    
    -- Audit
    created_by_user_id UUID,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 3. Promotion Usage Tracking
CREATE TABLE promotion_usages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    promotion_id UUID REFERENCES promotions(id),
    customer_id UUID,
    order_id UUID, -- Reference to orders table
    
    -- Usage Details
    discount_amount DECIMAL(10,2) NOT NULL,
    order_amount_before DECIMAL(12,2),
    order_amount_after DECIMAL(12,2),
    
    -- Metadata
    usage_date TIMESTAMP DEFAULT NOW(),
    user_agent TEXT,
    ip_address VARCHAR(45),
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 4. Customer Loyalty Program
CREATE TABLE loyalty_programs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    
    -- Program Configuration
    points_per_baht DECIMAL(5,2) DEFAULT 1, -- Points earned per THB spent
    points_to_baht_ratio DECIMAL(5,2) DEFAULT 1, -- Points to THB conversion
    
    -- Tier Thresholds
    bronze_threshold DECIMAL(12,2) DEFAULT 0,
    silver_threshold DECIMAL(12,2) DEFAULT 10000,
    gold_threshold DECIMAL(12,2) DEFAULT 50000,
    vip_threshold DECIMAL(12,2) DEFAULT 100000,
    
    -- Benefits
    tier_benefits JSONB, -- {"bronze": {"discount": 0}, "silver": {"discount": 5}, "gold": {"discount": 10}, "vip": {"discount": 15}}
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 5. Customer Loyalty Points
CREATE TABLE customer_loyalty_points (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL,
    
    -- Points Balance
    total_earned DECIMAL(12,2) DEFAULT 0,
    total_redeemed DECIMAL(12,2) DEFAULT 0,
    current_balance DECIMAL(12,2) GENERATED ALWAYS AS (total_earned - total_redeemed) STORED,
    
    -- Tier Information
    current_tier VARCHAR(20) DEFAULT 'bronze',
    tier_since DATE,
    points_to_next_tier DECIMAL(12,2),
    
    -- Metadata
    last_earning_date DATE,
    last_redemption_date DATE,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(customer_id)
);

-- 6. Loyalty Point Transactions
CREATE TABLE loyalty_point_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL,
    
    -- Transaction Details
    transaction_type VARCHAR(20) NOT NULL, -- 'earned', 'redeemed', 'expired', 'adjusted'
    points DECIMAL(12,2) NOT NULL, -- Positive for earned, negative for redeemed
    
    -- Reference
    reference_type VARCHAR(20), -- 'order', 'promotion', 'manual_adjustment'
    reference_id UUID,
    reference_number VARCHAR(50),
    
    -- Balance Tracking
    balance_before DECIMAL(12,2),
    balance_after DECIMAL(12,2),
    
    -- Expiry (for earned points)
    expires_at DATE,
    
    -- Description
    description TEXT,
    notes TEXT,
    
    -- Audit
    created_by_user_id UUID,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 7. Delivery Time Slots
CREATE TABLE delivery_time_slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Slot Definition
    name VARCHAR(100) NOT NULL, -- "Morning Slot", "Afternoon Slot"
    code VARCHAR(20) UNIQUE NOT NULL, -- "MORNING", "AFTERNOON", "EVENING"
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    
    -- Availability
    available_days JSONB, -- ["monday", "tuesday", "wednesday", "thursday", "friday"]
    available_routes JSONB, -- ["A", "B", "C", "D"] - which routes support this slot
    
    -- Capacity Management
    max_orders_per_slot INT DEFAULT 50,
    
    -- Pricing
    additional_fee DECIMAL(8,2) DEFAULT 0, -- Extra fee for this time slot
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 8. Holiday & Special Day Calendar
CREATE TABLE holidays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    date DATE NOT NULL,
    
    -- Holiday Type
    type VARCHAR(20), -- 'national', 'religious', 'company', 'seasonal'
    
    -- Business Impact
    is_working_day BOOLEAN DEFAULT false, -- Whether business operates
    delivery_available BOOLEAN DEFAULT false, -- Whether deliveries are made
    
    -- Special Rules
    special_delivery_fee DECIMAL(8,2), -- Extra fee for holiday delivery
    limited_routes JSONB, -- Only certain routes available
    modified_hours JSONB, -- {"start": "10:00", "end": "16:00"}
    
    -- Notes
    description TEXT,
    customer_message TEXT, -- Message to display to customers
    
    -- Recurring
    is_recurring BOOLEAN DEFAULT false,
    recurrence_rule VARCHAR(100), -- "annually", "first_monday_of_may"
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(date)
);

-- 9. Dynamic Pricing Rules
CREATE TABLE pricing_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    
    -- Rule Type
    rule_type VARCHAR(20) NOT NULL, -- 'delivery_fee', 'product_price', 'time_based', 'demand_based'
    
    -- Conditions
    conditions JSONB NOT NULL, -- {"day_of_week": ["saturday", "sunday"], "time_range": {"start": "18:00", "end": "22:00"}}
    
    -- Pricing Adjustments
    adjustment_type VARCHAR(20), -- 'percentage', 'fixed_amount', 'replacement'
    adjustment_value DECIMAL(10,2),
    
    -- Scope
    applicable_routes JSONB, -- ["A", "B", "C"]
    applicable_products JSONB, -- Product IDs or categories
    applicable_customer_tiers JSONB, -- ["gold", "vip"]
    
    -- Priority & Stacking
    priority INT DEFAULT 0, -- Higher numbers = higher priority
    can_stack BOOLEAN DEFAULT false,
    
    -- Validity
    start_date DATE,
    end_date DATE,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 10. Customer Feedback & Reviews
CREATE TABLE customer_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID,
    
    -- Feedback Type
    feedback_type VARCHAR(20) NOT NULL, -- 'delivery', 'product', 'service', 'general'
    
    -- Reference
    reference_type VARCHAR(20), -- 'order', 'delivery', 'product', 'driver'
    reference_id UUID,
    reference_number VARCHAR(50),
    
    -- Rating & Review
    rating INT CHECK (rating >= 1 AND rating <= 5), -- 1-5 stars
    title VARCHAR(200),
    comment TEXT,
    
    -- Specific Ratings
    delivery_rating INT,
    product_quality_rating INT,
    service_rating INT,
    value_rating INT,
    
    -- Customer Info (for anonymous feedback)
    customer_name VARCHAR(100),
    customer_email VARCHAR(100),
    customer_phone VARCHAR(20),
    
    -- Response & Resolution
    response TEXT,
    responded_by_user_id UUID,
    responded_at TIMESTAMP,
    
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'responded', 'resolved', 'escalated'
    resolution_notes TEXT,
    
    -- Visibility
    is_public BOOLEAN DEFAULT false, -- Show on website/app
    is_featured BOOLEAN DEFAULT false, -- Featured review
    
    -- Metadata
    source VARCHAR(20), -- 'website', 'app', 'sms', 'call', 'email'
    ip_address VARCHAR(45),
    user_agent TEXT,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 11. Automated Notifications Templates
CREATE TABLE notification_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL, -- 'order_confirmed', 'delivery_scheduled'
    
    -- Template Content
    subject VARCHAR(200),
    body_text TEXT, -- Plain text version
    body_html TEXT, -- HTML version
    sms_text TEXT, -- SMS version (max 160 chars)
    
    -- Templating
    variables JSONB, -- ["customer_name", "order_number", "delivery_date"]
    
    -- Channels
    email_enabled BOOLEAN DEFAULT true,
    sms_enabled BOOLEAN DEFAULT false,
    push_enabled BOOLEAN DEFAULT false,
    in_app_enabled BOOLEAN DEFAULT true,
    
    -- Trigger Conditions
    trigger_event VARCHAR(50), -- 'order_created', 'delivery_assigned', 'payment_received'
    trigger_conditions JSONB, -- Additional conditions
    
    -- Timing
    send_immediately BOOLEAN DEFAULT true,
    delay_minutes INT DEFAULT 0,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    -- Versioning
    version INT DEFAULT 1,
    is_current BOOLEAN DEFAULT true,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 12. API Rate Limiting
CREATE TABLE api_rate_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Client Identification
    client_type VARCHAR(20), -- 'user', 'api_key', 'ip_address'
    client_id VARCHAR(100), -- User ID, API key, or IP address
    
    -- Rate Limit Configuration
    endpoint_pattern VARCHAR(200), -- '/api/orders/*', '/api/customers'
    method VARCHAR(10), -- 'GET', 'POST', 'PUT', 'DELETE', '*'
    
    -- Limits
    requests_per_minute INT DEFAULT 60,
    requests_per_hour INT DEFAULT 1000,
    requests_per_day INT DEFAULT 10000,
    
    -- Current Usage (reset periodically)
    current_minute_count INT DEFAULT 0,
    current_hour_count INT DEFAULT 0,
    current_day_count INT DEFAULT 0,
    
    -- Reset Timestamps
    minute_reset_at TIMESTAMP DEFAULT NOW(),
    hour_reset_at TIMESTAMP DEFAULT NOW(),
    day_reset_at TIMESTAMP DEFAULT NOW(),
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    is_blocked BOOLEAN DEFAULT false,
    blocked_until TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(client_type, client_id, endpoint_pattern, method)
);

-- ================================================================
-- INDEXES FOR BUSINESS FEATURES
-- ================================================================

-- Business Settings indexes
CREATE INDEX idx_business_settings_key ON business_settings(key);
CREATE INDEX idx_business_settings_category ON business_settings(category);

-- Promotions indexes
CREATE INDEX idx_promotions_code ON promotions(code);
CREATE INDEX idx_promotions_active_dates ON promotions(is_active, start_date, end_date);
CREATE INDEX idx_promotions_type ON promotions(type);

-- Promotion Usages indexes
CREATE INDEX idx_promotion_usages_promotion ON promotion_usages(promotion_id);
CREATE INDEX idx_promotion_usages_customer ON promotion_usages(customer_id);
CREATE INDEX idx_promotion_usages_order ON promotion_usages(order_id);

-- Loyalty Programs indexes
CREATE INDEX idx_loyalty_programs_active ON loyalty_programs(is_active);

-- Customer Loyalty Points indexes
CREATE INDEX idx_customer_loyalty_points_customer ON customer_loyalty_points(customer_id);
CREATE INDEX idx_customer_loyalty_points_tier ON customer_loyalty_points(current_tier);

-- Loyalty Point Transactions indexes
CREATE INDEX idx_loyalty_transactions_customer ON loyalty_point_transactions(customer_id);
CREATE INDEX idx_loyalty_transactions_type ON loyalty_point_transactions(transaction_type);
CREATE INDEX idx_loyalty_transactions_reference ON loyalty_point_transactions(reference_type, reference_id);
CREATE INDEX idx_loyalty_transactions_date ON loyalty_point_transactions(created_at);

-- Delivery Time Slots indexes
CREATE INDEX idx_delivery_time_slots_active ON delivery_time_slots(is_active);
CREATE INDEX idx_delivery_time_slots_time ON delivery_time_slots(start_time, end_time);

-- Holidays indexes
CREATE INDEX idx_holidays_date ON holidays(date);
CREATE INDEX idx_holidays_type ON holidays(type);
CREATE INDEX idx_holidays_working ON holidays(is_working_day, delivery_available);

-- Pricing Rules indexes
CREATE INDEX idx_pricing_rules_type ON pricing_rules(rule_type);
CREATE INDEX idx_pricing_rules_active_dates ON pricing_rules(is_active, start_date, end_date);
CREATE INDEX idx_pricing_rules_priority ON pricing_rules(priority);

-- Customer Feedback indexes
CREATE INDEX idx_customer_feedback_customer ON customer_feedback(customer_id);
CREATE INDEX idx_customer_feedback_type ON customer_feedback(feedback_type);
CREATE INDEX idx_customer_feedback_reference ON customer_feedback(reference_type, reference_id);
CREATE INDEX idx_customer_feedback_rating ON customer_feedback(rating);
CREATE INDEX idx_customer_feedback_status ON customer_feedback(status);
CREATE INDEX idx_customer_feedback_public ON customer_feedback(is_public, is_featured);

-- Notification Templates indexes
CREATE INDEX idx_notification_templates_code ON notification_templates(code);
CREATE INDEX idx_notification_templates_event ON notification_templates(trigger_event);
CREATE INDEX idx_notification_templates_active ON notification_templates(is_active);
CREATE INDEX idx_notification_templates_current ON notification_templates(is_current);

-- API Rate Limits indexes
CREATE INDEX idx_api_rate_limits_client ON api_rate_limits(client_type, client_id);
CREATE INDEX idx_api_rate_limits_endpoint ON api_rate_limits(endpoint_pattern, method);
CREATE INDEX idx_api_rate_limits_active ON api_rate_limits(is_active, is_blocked);
