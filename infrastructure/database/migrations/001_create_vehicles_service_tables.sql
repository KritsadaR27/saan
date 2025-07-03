-- SaaN Vehicles Service - Complete Database Schema
-- Migration: 001_create_vehicles_service_tables.sql
-- Date: 2025-07-02

-- ================================================================
-- VEHICLES SERVICE TABLES
-- ================================================================

-- 1. Vehicles Table (Enhanced with Insurance/Tax/Maintenance)
CREATE TABLE vehicles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,                      -- ชื่อเรียกของรถ เช่น "Exceed"
    license_plate VARCHAR(20) UNIQUE NOT NULL, -- ทะเบียนรถ
    vehicle_number INT UNIQUE NOT NULL, -- 1-10 รหัสรถ
    cartrack_id VARCHAR(100), -- รหัสจาก Cartrack GPS
    
    -- Vehicle Details
    brand VARCHAR(50),
    model VARCHAR(50),
    year INT,
    vehicle_type VARCHAR(20), -- 'truck', 'van', 'motorcycle'
    fuel_type VARCHAR(20), -- 'gasoline', 'diesel', 'electric'
    
    -- ประกันภัย
    insurance_company VARCHAR(100),          -- บริษัทประกัน
    insurance_type VARCHAR(50),              -- ชั้นประกัน เช่น "ชั้น 1"
    insurance_policy_number VARCHAR(100),    -- เลขที่กรมธรรม์
    insurance_cost DECIMAL(10,2),            -- ค่าประกัน
    insurance_expiry_date DATE,              -- วันหมดอายุประกัน
    
    -- พรบ.
    compulsory_insurance_expiry_date DATE,   -- วันหมดอายุ พรบ.
    
    -- ภาษี
    tax_expiry_date DATE,                    -- วันหมดภาษีรถยนต์
    
    -- รายละเอียดทางเทคนิค
    chassis_number VARCHAR(100),             -- เลขตัวถัง
    renewal_premium DECIMAL(10,2),           -- เบี้ยต่ออายุประกัน
    insurance_sum DECIMAL(10,2),             -- ทุนประกัน
    base_premium DECIMAL(10,2),              -- เบี้ยประกันพื้นฐาน
    
    -- Capacity
    capacity_orders INT DEFAULT 20, -- จำนวนออเดอร์สูงสุด
    capacity_weight DECIMAL(8,2), -- กิโลกรัม
    capacity_volume DECIMAL(8,2), -- ลิตร
    
    -- Operational Info
    status VARCHAR(20) DEFAULT 'available', -- 'available', 'assigned', 'maintenance', 'retired'
    current_mileage INT DEFAULT 0,
    fuel_efficiency DECIMAL(5,2), -- กม./ลิตร
    last_service_date DATE,
    next_service_mileage INT,
    
    -- ซ่อมบำรุง
    maintenance_note TEXT,                   -- ข้อความหรือบันทึกการซ่อมล่าสุด
    
    -- Assignment
    preferred_routes JSONB, -- ["A", "B", "C"] routes this vehicle usually covers
    current_driver_id UUID, -- คนขับปัจจุบัน
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,          -- รถยังใช้งานอยู่หรือไม่
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 2. Drivers Table
CREATE TABLE drivers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_number INT UNIQUE NOT NULL, -- 1-10 หมายเลขพนักงาน
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    
    -- License Info
    license_number VARCHAR(50),
    license_expiry DATE,
    hire_date DATE,
    
    -- Skills & Certifications
    vehicle_types JSONB, -- ["truck", "van", "motorcycle"] ประเภทรถที่ขับได้
    preferred_routes JSONB, -- ["A", "B", "C"] เส้นทางที่ถนัด
    max_orders_per_round INT DEFAULT 20, -- จำนวนออเดอร์สูงสุดต่อรอบ
    
    -- Performance Tracking
    total_deliveries INT DEFAULT 0,
    success_rate DECIMAL(5,2) DEFAULT 100.00,
    average_delivery_time INT, -- นาที/ออเดอร์
    customer_rating DECIMAL(3,2) DEFAULT 5.00,
    
    -- Schedule & Availability
    work_schedule JSONB, -- {"monday": true, "tuesday": true, ...}
    shift_preference VARCHAR(20) DEFAULT 'morning', -- 'morning', 'afternoon', 'full_day'
    
    -- Status
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'on_leave', 'sick', 'terminated'
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 3. Driver Vehicle Assignments
CREATE TABLE driver_vehicle_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID REFERENCES drivers(id),
    vehicle_id UUID REFERENCES vehicles(id),
    assignment_date DATE NOT NULL,
    
    -- Assignment Details
    shift VARCHAR(20), -- 'morning', 'afternoon', 'evening'
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    
    -- Performance
    total_orders INT DEFAULT 0,
    completed_orders INT DEFAULT 0,
    total_distance DECIMAL(8,2),
    fuel_consumed DECIMAL(6,2),
    
    -- Status
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'completed', 'cancelled'
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 4. Delivery Rounds
CREATE TABLE delivery_rounds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    round_name VARCHAR(100), -- "รอบเช้า A-1", "รอบบ่าย B-2"
    delivery_date DATE NOT NULL,
    
    -- Vehicle & Driver Assignment
    vehicle_id UUID REFERENCES vehicles(id),
    driver_id UUID REFERENCES drivers(id),
    
    -- Route Information
    route_code VARCHAR(10), -- A, B, C, D, E, F, G, H, J
    delivery_days VARCHAR(50), -- "Tuesday,Friday"
    shift VARCHAR(20), -- 'morning', 'afternoon', 'evening'
    
    -- Planning Data
    planned_start_time TIME,
    planned_end_time TIME,
    estimated_orders INT,
    estimated_distance DECIMAL(8,2), -- กิโลเมตร
    estimated_fuel_cost DECIMAL(8,2),
    
    -- Execution Data
    actual_start_time TIMESTAMP,
    actual_end_time TIMESTAMP,
    actual_distance DECIMAL(8,2),
    actual_fuel_cost DECIMAL(8,2),
    total_orders INT DEFAULT 0,
    completed_orders INT DEFAULT 0,
    failed_orders INT DEFAULT 0,
    
    -- Status & Tracking
    status VARCHAR(20) DEFAULT 'planned', -- 'planned', 'driver_selected', 'active', 'completed', 'cancelled'
    driver_accepted_at TIMESTAMP, -- เมื่อไหร่ที่ driver เลือกรถ
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    
    -- GPS & Location
    start_location JSONB, -- {"lat": 13.7563, "lng": 100.5018, "name": "คลังสินค้าหลัก"}
    current_location JSONB, -- real-time location
    
    -- Performance Metrics
    efficiency_score DECIMAL(5,2), -- คะแนนประสิทธิภาพ
    customer_satisfaction DECIMAL(3,2), -- คะแนนความพอใจลูกค้า
    
    -- Notes
    planning_notes TEXT,
    driver_notes TEXT,
    admin_notes TEXT,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 5. Delivery Assignments
CREATE TABLE delivery_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delivery_round_id UUID REFERENCES delivery_rounds(id),
    order_id UUID, -- Will reference orders table when created
    
    -- Delivery Sequence (ลำดับการส่ง 1-20)
    delivery_sequence INT NOT NULL,
    estimated_arrival_time TIMESTAMP,
    estimated_service_duration INT DEFAULT 5, -- นาที
    
    -- Customer Info (denormalized for performance)
    customer_name VARCHAR(200),
    customer_phone VARCHAR(20),
    delivery_address TEXT,
    delivery_notes TEXT,
    special_instructions TEXT,
    
    -- GPS Location
    customer_lat DECIMAL(10,8),
    customer_lng DECIMAL(11,8),
    
    -- Execution Tracking
    actual_arrival_time TIMESTAMP,
    actual_departure_time TIMESTAMP,
    actual_completion_time TIMESTAMP,
    delivery_attempts INT DEFAULT 0,
    delivery_status VARCHAR(20) DEFAULT 'pending', 
    -- 'pending', 'en_route', 'arrived', 'delivered', 'failed', 'rescheduled'
    
    -- Customer Interaction
    customer_signature_url TEXT,
    delivery_photo_url TEXT,
    customer_rating INT, -- 1-5 ดาว
    customer_feedback TEXT,
    delivery_notes_actual TEXT,
    
    -- Failure Handling
    failure_reason TEXT,
    failure_category VARCHAR(50), -- 'customer_not_home', 'wrong_address', 'refused_delivery'
    reschedule_requested BOOLEAN DEFAULT false,
    reschedule_date DATE,
    
    -- COD Tracking
    cod_amount DECIMAL(10,2),
    cod_collected BOOLEAN DEFAULT false,
    cod_collection_time TIMESTAMP,
    cod_method VARCHAR(20), -- 'cash', 'qr_code', 'transfer'
    
    -- Performance Metrics
    service_time_minutes INT, -- เวลาที่ใช้ในการส่งจริง
    waiting_time_minutes INT, -- เวลาที่รอลูกค้า
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 6. Vehicle Maintenance
CREATE TABLE vehicle_maintenance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID REFERENCES vehicles(id),
    maintenance_type VARCHAR(50), -- 'inspection', 'repair', 'cleaning', 'tire_change'
    title VARCHAR(200),
    description TEXT,
    
    -- Scheduling
    scheduled_date DATE,
    actual_date DATE,
    estimated_cost DECIMAL(10,2),
    actual_cost DECIMAL(10,2),
    
    -- Vendor Info
    vendor_name VARCHAR(100),
    vendor_contact VARCHAR(100),
    receipt_url TEXT,
    
    -- Maintenance Details
    mileage_at_service INT,
    parts_replaced JSONB, -- ["brake_pads", "oil_filter"]
    next_service_mileage INT,
    next_service_date DATE,
    
    -- Status
    status VARCHAR(20) DEFAULT 'scheduled', -- 'scheduled', 'in_progress', 'completed', 'cancelled'
    created_at TIMESTAMP DEFAULT NOW()
);

-- 7. Vehicle Documents
CREATE TABLE vehicle_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID REFERENCES vehicles(id),
    document_type VARCHAR(50), -- 'insurance', 'registration', 'inspection'
    
    -- Document Details
    provider VARCHAR(100), -- "กรุงเทพประกันภัย", "กรมการขนส่ง"
    policy_number VARCHAR(100),
    start_date DATE,
    end_date DATE,
    cost DECIMAL(10,2),
    
    -- File Storage
    document_file_url TEXT,
    renewal_reminder_days INT DEFAULT 30, -- แจ้งเตือนก่อนหมดอายุ 30 วัน
    
    -- Status
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'expired', 'cancelled'
    auto_renew BOOLEAN DEFAULT false,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 8. Vehicle GPS Logs
CREATE TABLE vehicle_gps_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID REFERENCES vehicles(id),
    delivery_round_id UUID REFERENCES delivery_rounds(id),
    cartrack_id VARCHAR(100),
    
    -- Location Data
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    speed DECIMAL(5,2), -- km/h
    heading INT, -- 0-360 degrees
    altitude DECIMAL(8,2),
    
    -- Vehicle Status
    engine_status VARCHAR(20), -- 'on', 'off', 'idle'
    fuel_level DECIMAL(5,2), -- เปอร์เซ็นต์
    mileage DECIMAL(10,2),
    
    -- Context
    event_type VARCHAR(50), -- 'delivery_start', 'en_route', 'arrived', 'delivered', 'return_to_base'
    assignment_id UUID, -- ถ้าเกี่ยวข้องกับ delivery assignment
    
    -- Timestamp
    gps_timestamp TIMESTAMP,
    received_at TIMESTAMP DEFAULT NOW()
);

-- 9. Daily Vehicle Costs
CREATE TABLE daily_vehicle_costs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID REFERENCES vehicles(id),
    delivery_round_id UUID REFERENCES delivery_rounds(id),
    date DATE,
    
    -- Cost Breakdown
    fuel_cost DECIMAL(8,2) DEFAULT 0,
    toll_cost DECIMAL(8,2) DEFAULT 0,
    parking_cost DECIMAL(8,2) DEFAULT 0,
    maintenance_cost DECIMAL(8,2) DEFAULT 0,
    driver_wage DECIMAL(8,2) DEFAULT 0,
    insurance_cost DECIMAL(8,2) DEFAULT 0,
    other_costs DECIMAL(8,2) DEFAULT 0,
    
    -- Performance Metrics
    total_distance DECIMAL(8,2),
    total_delivery_time INT, -- นาที
    successful_deliveries INT,
    failed_deliveries INT,
    rescheduled_deliveries INT,
    
    -- Customer Satisfaction
    total_ratings INT,
    average_rating DECIMAL(3,2),
    complaints INT,
    
    -- Calculated Metrics
    cost_per_km DECIMAL(6,3),
    cost_per_delivery DECIMAL(8,2),
    revenue_generated DECIMAL(10,2),
    profit_margin DECIMAL(5,2),
    efficiency_score DECIMAL(5,2),
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- ================================================================
-- INDEXES FOR VEHICLES SERVICE
-- ================================================================

-- Vehicles table indexes
CREATE INDEX idx_vehicles_plate ON vehicles(license_plate);
CREATE INDEX idx_vehicles_number ON vehicles(vehicle_number);
CREATE INDEX idx_vehicles_status ON vehicles(status);
CREATE INDEX idx_vehicles_insurance_expiry ON vehicles(insurance_expiry_date);
CREATE INDEX idx_vehicles_tax_expiry ON vehicles(tax_expiry_date);
CREATE INDEX idx_vehicles_compulsory_expiry ON vehicles(compulsory_insurance_expiry_date);

-- Drivers table indexes
CREATE INDEX idx_drivers_number ON drivers(driver_number);
CREATE INDEX idx_drivers_phone ON drivers(phone);
CREATE INDEX idx_drivers_status ON drivers(status);

-- Driver Vehicle Assignments indexes
CREATE INDEX idx_driver_vehicle_date ON driver_vehicle_assignments(driver_id, vehicle_id, assignment_date);
CREATE INDEX idx_assignment_date ON driver_vehicle_assignments(assignment_date);

-- Delivery Rounds indexes
CREATE INDEX idx_delivery_rounds_date ON delivery_rounds(delivery_date);
CREATE INDEX idx_delivery_rounds_route ON delivery_rounds(route_code, delivery_date);
CREATE INDEX idx_delivery_rounds_driver ON delivery_rounds(driver_id, delivery_date);
CREATE INDEX idx_delivery_rounds_status ON delivery_rounds(status);

-- Delivery Assignments indexes
CREATE INDEX idx_assignments_round_sequence ON delivery_assignments(delivery_round_id, delivery_sequence);
CREATE INDEX idx_assignments_order ON delivery_assignments(order_id);
CREATE INDEX idx_assignments_status ON delivery_assignments(delivery_status);
CREATE INDEX idx_assignments_date ON delivery_assignments(created_at);

-- Vehicle Maintenance indexes
CREATE INDEX idx_vehicle_maintenance ON vehicle_maintenance(vehicle_id, scheduled_date);
CREATE INDEX idx_maintenance_type ON vehicle_maintenance(maintenance_type);
CREATE INDEX idx_maintenance_status ON vehicle_maintenance(status);

-- Vehicle Documents indexes
CREATE INDEX idx_vehicle_documents ON vehicle_documents(vehicle_id, document_type);
CREATE INDEX idx_documents_expiry_reminder ON vehicle_documents(end_date, status);

-- Vehicle GPS Logs indexes
CREATE INDEX idx_vehicle_gps_time ON vehicle_gps_logs(vehicle_id, gps_timestamp);
CREATE INDEX idx_cartrack_gps_time ON vehicle_gps_logs(cartrack_id, gps_timestamp);
CREATE INDEX idx_round_gps_tracking ON vehicle_gps_logs(delivery_round_id, gps_timestamp);

-- Daily Vehicle Costs indexes
CREATE INDEX idx_vehicle_costs_date ON daily_vehicle_costs(vehicle_id, date);
CREATE INDEX idx_round_costs ON daily_vehicle_costs(delivery_round_id);
