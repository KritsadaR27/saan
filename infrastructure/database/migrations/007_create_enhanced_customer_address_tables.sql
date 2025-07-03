-- SaaN Enhanced Customer & Address Management
-- Migration: 007_create_enhanced_customer_address_tables.sql
-- Date: 2025-07-02

-- ================================================================
-- ENHANCED CUSTOMER & ADDRESS MANAGEMENT TABLES
-- ================================================================

-- 1. Enhanced Customers Table (Updated)
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Basic Information
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    display_name VARCHAR(200), -- Preferred name for display
    email VARCHAR(100) UNIQUE,
    phone VARCHAR(20) UNIQUE NOT NULL,
    
    -- Alternative Contact
    secondary_phone VARCHAR(20),
    line_id VARCHAR(100),
    facebook_id VARCHAR(100),
    
    -- Demographics
    date_of_birth DATE,
    gender VARCHAR(10), -- 'male', 'female', 'other', 'prefer_not_to_say'
    occupation VARCHAR(100),
    
    -- Business Information
    company_name VARCHAR(200),
    tax_id VARCHAR(50),
    business_type VARCHAR(50), -- 'individual', 'company', 'partnership'
    
    -- Customer Segmentation
    tier VARCHAR(20) DEFAULT 'bronze', -- 'bronze', 'silver', 'gold', 'vip'
    segment VARCHAR(20), -- 'new', 'regular', 'vip', 'at_risk', 'churned'
    source VARCHAR(50), -- 'website', 'referral', 'social_media', 'walk_in'
    referred_by_customer_id UUID REFERENCES customers(id),
    
    -- Financial Summary
    total_spent DECIMAL(12,2) DEFAULT 0,
    total_orders INT DEFAULT 0,
    average_order_value DECIMAL(10,2) DEFAULT 0,
    credit_limit DECIMAL(12,2) DEFAULT 0,
    outstanding_balance DECIMAL(12,2) DEFAULT 0,
    
    -- Behavioral Data
    last_order_date DATE,
    last_login_date DATE,
    preferred_payment_method VARCHAR(20),
    preferred_contact_method VARCHAR(20), -- 'phone', 'email', 'sms', 'line'
    
    -- Preferences
    communication_preferences JSONB, -- {"email": true, "sms": false, "promotions": true}
    delivery_preferences JSONB, -- {"time_slot": "morning", "special_instructions": "..."}
    language_preference VARCHAR(10) DEFAULT 'th',
    
    -- Account Status
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'inactive', 'suspended', 'blacklisted'
    is_verified BOOLEAN DEFAULT false,
    verification_method VARCHAR(20), -- 'phone', 'email', 'id_card'
    verification_date DATE,
    
    -- Risk Management
    risk_level VARCHAR(20) DEFAULT 'low', -- 'low', 'medium', 'high'
    blacklist_reason TEXT,
    blacklisted_by_user_id UUID,
    blacklisted_at TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_activity_at TIMESTAMP DEFAULT NOW()
);

-- 2. Enhanced Customer Addresses Table (Updated)
CREATE TABLE customer_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID REFERENCES customers(id),
    
    -- Address Details
    address_line1 TEXT NOT NULL,
    address_line2 TEXT,
    subdistrict VARCHAR(100) NOT NULL,
    district VARCHAR(100) NOT NULL,
    province VARCHAR(100) NOT NULL,
    postal_code VARCHAR(10) NOT NULL,
    
    -- Address Metadata
    address_type VARCHAR(20) DEFAULT 'home', -- 'home', 'office', 'warehouse', 'other'
    label VARCHAR(100), -- "บ้าน", "ออฟฟิศ", "คลัง", "บ้านแม่"
    is_default BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    
    -- GPS Coordinates
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    location_accuracy DECIMAL(5,2), -- GPS accuracy in meters
    
    -- Delivery Route Mapping
    delivery_route_id UUID, -- Will reference delivery_routes table
    
    -- Access Information
    building_name VARCHAR(200),
    floor_number VARCHAR(10),
    room_number VARCHAR(20),
    landmark TEXT, -- "ใกล้ 7-Eleven", "ตรงข้าม Big C"
    
    -- Delivery Instructions
    delivery_instructions TEXT,
    access_notes TEXT, -- "ผ่านประตูหลัง", "เรียกก่อนส่ง"
    special_requirements TEXT, -- "ไม่มีลิฟต์", "ห้ามส่งหลัง 20:00"
    
    -- Contact at Address
    contact_person VARCHAR(100), -- If different from customer
    contact_phone VARCHAR(20),
    contact_relationship VARCHAR(50), -- 'self', 'spouse', 'parent', 'sibling', 'colleague'
    
    -- Validation & Quality
    is_verified BOOLEAN DEFAULT false,
    verification_method VARCHAR(20), -- 'delivery', 'phone', 'manual'
    verification_date DATE,
    delivery_success_rate DECIMAL(5,2) DEFAULT 100, -- % of successful deliveries
    last_successful_delivery DATE,
    
    -- Statistics
    total_deliveries INT DEFAULT 0,
    failed_deliveries INT DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 3. Delivery Routes Table (Enhanced)
CREATE TABLE delivery_routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Route Identification
    route_code VARCHAR(10) UNIQUE NOT NULL, -- A, B, C, D, E, F, G, H, J
    route_name VARCHAR(100), -- "เส้นทาง A - กรุงเทพเหนือ"
    
    -- Geographic Coverage
    provinces JSONB, -- ["กรุงเทพมหานคร", "นนทบุรี"]
    districts JSONB, -- ["บางกะปิ", "ห้วยขวาง"]
    subdistricts JSONB, -- ["คลองจั่น", "รามอินทรา"]
    postal_codes JSONB, -- ["10240", "10310"]
    
    -- Delivery Schedule
    delivery_days VARCHAR(100), -- "Tuesday,Friday" or "Daily" or "Weekdays"
    delivery_time_slots JSONB, -- ["morning", "afternoon"]
    
    -- Route Characteristics
    route_type VARCHAR(20), -- 'urban', 'suburban', 'rural', 'mixed'
    difficulty_level VARCHAR(20) DEFAULT 'normal', -- 'easy', 'normal', 'difficult'
    traffic_level VARCHAR(20) DEFAULT 'normal', -- 'light', 'normal', 'heavy'
    
    -- Capacity & Limits
    max_orders_per_day INT DEFAULT 100,
    max_vehicles_per_day INT DEFAULT 5,
    estimated_travel_time INT, -- minutes per delivery
    
    -- Geographic Data
    center_latitude DECIMAL(10,8),
    center_longitude DECIMAL(11,8),
    coverage_radius DECIMAL(8,2), -- kilometers
    boundary_polygon JSONB, -- GeoJSON polygon
    
    -- Costs & Pricing
    base_delivery_fee DECIMAL(8,2),
    cost_per_km DECIMAL(6,3),
    fuel_cost_factor DECIMAL(5,2) DEFAULT 1.0,
    
    -- Performance Metrics
    average_delivery_time INT, -- minutes
    success_rate DECIMAL(5,2) DEFAULT 100,
    customer_satisfaction DECIMAL(3,2) DEFAULT 5.0,
    
    -- Status & Management
    is_active BOOLEAN DEFAULT true,
    priority INT DEFAULT 0, -- Higher priority routes get preference
    
    -- Notes
    description TEXT,
    special_requirements TEXT,
    manager_notes TEXT,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 4. Thai Address Reference Tables (Enhanced)

-- Provinces table
CREATE TABLE thai_provinces (
    id INT PRIMARY KEY,
    name_th VARCHAR(100) NOT NULL,
    name_en VARCHAR(100),
    code VARCHAR(10),
    region VARCHAR(20), -- 'north', 'northeast', 'central', 'south'
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- Districts table
CREATE TABLE thai_districts (
    id INT PRIMARY KEY,
    province_id INT REFERENCES thai_provinces(id),
    name_th VARCHAR(100) NOT NULL,
    name_en VARCHAR(100),
    code VARCHAR(10),
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- Subdistricts table
CREATE TABLE thai_subdistricts (
    id INT PRIMARY KEY,
    district_id INT REFERENCES thai_districts(id),
    province_id INT REFERENCES thai_provinces(id),
    name_th VARCHAR(100) NOT NULL,
    name_en VARCHAR(100),
    postal_code VARCHAR(10),
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 5. Customer Groups & Segments
CREATE TABLE customer_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    
    -- Group Type
    group_type VARCHAR(20), -- 'tier', 'segment', 'custom', 'promotional'
    
    -- Criteria (for automatic assignment)
    criteria JSONB, -- {"min_spent": 10000, "min_orders": 5, "last_order_days": 30}
    
    -- Benefits
    benefits JSONB, -- {"discount_percentage": 10, "free_delivery": true, "priority_support": true}
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    is_automatic BOOLEAN DEFAULT false, -- Auto-assign customers based on criteria
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 6. Customer Group Memberships
CREATE TABLE customer_group_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID REFERENCES customers(id),
    customer_group_id UUID REFERENCES customer_groups(id),
    
    -- Membership Details
    joined_date DATE DEFAULT CURRENT_DATE,
    expires_date DATE,
    
    -- Assignment
    assignment_type VARCHAR(20) DEFAULT 'manual', -- 'manual', 'automatic', 'promotional'
    assigned_by_user_id UUID,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(customer_id, customer_group_id)
);

-- 7. Customer Communication Log
CREATE TABLE customer_communications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID REFERENCES customers(id),
    
    -- Communication Details
    communication_type VARCHAR(20) NOT NULL, -- 'call', 'email', 'sms', 'meeting', 'chat'
    direction VARCHAR(10) NOT NULL, -- 'inbound', 'outbound'
    subject VARCHAR(200),
    content TEXT,
    
    -- Metadata
    channel VARCHAR(20), -- 'phone', 'email', 'line', 'facebook', 'website_chat'
    status VARCHAR(20) DEFAULT 'completed', -- 'pending', 'completed', 'failed', 'scheduled'
    
    -- Staff & Response
    handled_by_user_id UUID,
    response_time_minutes INT,
    satisfaction_rating INT, -- 1-5 stars from customer
    
    -- Follow-up
    requires_followup BOOLEAN DEFAULT false,
    followup_date DATE,
    followup_notes TEXT,
    
    -- Attachments
    attachments JSONB, -- File URLs or references
    
    -- Scheduling (for future communications)
    scheduled_for TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 8. Customer Visit History (for walk-in customers)
CREATE TABLE customer_visits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID REFERENCES customers(id),
    
    -- Visit Details
    visit_date DATE NOT NULL,
    visit_time TIME,
    visit_duration INT, -- minutes
    
    -- Location
    location VARCHAR(100), -- 'main_office', 'warehouse_1', 'showroom'
    
    -- Purpose
    visit_purpose VARCHAR(50), -- 'purchase', 'inquiry', 'complaint', 'meeting', 'pickup'
    visit_result VARCHAR(50), -- 'sale_made', 'information_provided', 'issue_resolved'
    
    -- Staff
    served_by_user_id UUID,
    
    -- Notes
    notes TEXT,
    customer_feedback TEXT,
    
    -- Follow-up
    requires_followup BOOLEAN DEFAULT false,
    followup_assigned_to UUID,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- ================================================================
-- INDEXES FOR ENHANCED CUSTOMER & ADDRESS MANAGEMENT
-- ================================================================

-- Enhanced Customers indexes
CREATE INDEX idx_customers_phone ON customers(phone);
CREATE INDEX idx_customers_email ON customers(email);
CREATE INDEX idx_customers_tier ON customers(tier);
CREATE INDEX idx_customers_segment ON customers(segment);
CREATE INDEX idx_customers_status ON customers(status);
CREATE INDEX idx_customers_last_order ON customers(last_order_date);
CREATE INDEX idx_customers_total_spent ON customers(total_spent);
CREATE INDEX idx_customers_risk_level ON customers(risk_level);
CREATE INDEX idx_customers_referred_by ON customers(referred_by_customer_id);

-- Enhanced Customer Addresses indexes
CREATE INDEX idx_customer_addresses_customer ON customer_addresses(customer_id);
CREATE INDEX idx_customer_addresses_delivery_route ON customer_addresses(delivery_route_id);
CREATE INDEX idx_customer_addresses_location ON customer_addresses(latitude, longitude);
CREATE INDEX idx_customer_addresses_postal ON customer_addresses(postal_code);
CREATE INDEX idx_customer_addresses_default ON customer_addresses(customer_id, is_default);
CREATE INDEX idx_customer_addresses_active ON customer_addresses(is_active);
CREATE INDEX idx_customer_addresses_type ON customer_addresses(address_type);

-- Delivery Routes indexes
CREATE INDEX idx_delivery_routes_code ON delivery_routes(route_code);
CREATE INDEX idx_delivery_routes_active ON delivery_routes(is_active);
CREATE INDEX idx_delivery_routes_priority ON delivery_routes(priority);
CREATE INDEX idx_delivery_routes_difficulty ON delivery_routes(difficulty_level);

-- Thai Address Reference indexes
CREATE INDEX idx_thai_provinces_name ON thai_provinces(name_th);
CREATE INDEX idx_thai_districts_province ON thai_districts(province_id);
CREATE INDEX idx_thai_districts_name ON thai_districts(name_th);
CREATE INDEX idx_thai_subdistricts_district ON thai_subdistricts(district_id);
CREATE INDEX idx_thai_subdistricts_province ON thai_subdistricts(province_id);
CREATE INDEX idx_thai_subdistricts_postal ON thai_subdistricts(postal_code);

-- Customer Groups indexes
CREATE INDEX idx_customer_groups_type ON customer_groups(group_type);
CREATE INDEX idx_customer_groups_active ON customer_groups(is_active);
CREATE INDEX idx_customer_groups_automatic ON customer_groups(is_automatic);

-- Customer Group Memberships indexes
CREATE INDEX idx_customer_memberships_customer ON customer_group_memberships(customer_id);
CREATE INDEX idx_customer_memberships_group ON customer_group_memberships(customer_group_id);
CREATE INDEX idx_customer_memberships_active ON customer_group_memberships(is_active);

-- Customer Communications indexes
CREATE INDEX idx_customer_communications_customer ON customer_communications(customer_id);
CREATE INDEX idx_customer_communications_type ON customer_communications(communication_type);
CREATE INDEX idx_customer_communications_handled_by ON customer_communications(handled_by_user_id);
CREATE INDEX idx_customer_communications_date ON customer_communications(created_at);
CREATE INDEX idx_customer_communications_followup ON customer_communications(requires_followup, followup_date);

-- Customer Visits indexes
CREATE INDEX idx_customer_visits_customer ON customer_visits(customer_id);
CREATE INDEX idx_customer_visits_date ON customer_visits(visit_date);
CREATE INDEX idx_customer_visits_served_by ON customer_visits(served_by_user_id);
CREATE INDEX idx_customer_visits_purpose ON customer_visits(visit_purpose);
