-- SaaN Analytics & Reporting Service - Complete Database Schema
-- Migration: 005_create_analytics_reporting_tables.sql
-- Date: 2025-07-02

-- ================================================================
-- ANALYTICS & REPORTING SERVICE TABLES
-- ================================================================

-- 1. Analytics Events (Event tracking for business intelligence)
CREATE TABLE analytics_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Event Details
    event_type VARCHAR(50) NOT NULL, -- 'order_created', 'customer_registered', 'delivery_completed'
    event_category VARCHAR(30), -- 'sales', 'logistics', 'customer', 'system'
    event_action VARCHAR(50), -- 'create', 'update', 'complete', 'cancel'
    
    -- Entity Information
    entity_type VARCHAR(30), -- 'order', 'customer', 'delivery', 'product'
    entity_id UUID,
    
    -- User Context
    user_id UUID,
    customer_id UUID,
    session_id VARCHAR(100),
    
    -- Event Properties
    properties JSONB, -- {"order_value": 1500, "payment_method": "cash", "delivery_zone": "A"}
    
    -- Location & Device
    ip_address VARCHAR(45),
    user_agent TEXT,
    device_info JSONB,
    
    -- Business Context
    revenue_impact DECIMAL(12,2), -- How much revenue this event generated/affected
    cost_impact DECIMAL(12,2), -- How much cost this event incurred
    
    -- Timestamp (partitioned by date for performance)
    event_timestamp TIMESTAMP DEFAULT NOW(),
    date_partition DATE GENERATED ALWAYS AS (event_timestamp::DATE) STORED,
    
    created_at TIMESTAMP DEFAULT NOW()
) PARTITION BY RANGE (date_partition);

-- 2. Daily Business Metrics
CREATE TABLE daily_business_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE UNIQUE NOT NULL,
    
    -- Sales Metrics
    total_orders INT DEFAULT 0,
    total_revenue DECIMAL(12,2) DEFAULT 0,
    total_items_sold INT DEFAULT 0,
    average_order_value DECIMAL(10,2) DEFAULT 0,
    new_customers INT DEFAULT 0,
    returning_customers INT DEFAULT 0,
    
    -- Delivery Metrics
    total_deliveries INT DEFAULT 0,
    successful_deliveries INT DEFAULT 0,
    failed_deliveries INT DEFAULT 0,
    delivery_success_rate DECIMAL(5,2) DEFAULT 0,
    average_delivery_time INT, -- minutes
    total_delivery_distance DECIMAL(10,2), -- kilometers
    
    -- Vehicle Metrics
    active_vehicles INT DEFAULT 0,
    total_fuel_cost DECIMAL(10,2) DEFAULT 0,
    total_maintenance_cost DECIMAL(10,2) DEFAULT 0,
    vehicle_utilization_rate DECIMAL(5,2) DEFAULT 0,
    
    -- Inventory Metrics
    total_products INT DEFAULT 0,
    low_stock_alerts INT DEFAULT 0,
    out_of_stock_items INT DEFAULT 0,
    inventory_turnover DECIMAL(5,2) DEFAULT 0,
    
    -- Financial Metrics
    total_payments DECIMAL(12,2) DEFAULT 0,
    cash_payments DECIMAL(12,2) DEFAULT 0,
    digital_payments DECIMAL(12,2) DEFAULT 0,
    cod_collected DECIMAL(12,2) DEFAULT 0,
    outstanding_invoices DECIMAL(12,2) DEFAULT 0,
    
    -- Customer Service Metrics
    customer_satisfaction DECIMAL(3,2) DEFAULT 0,
    total_complaints INT DEFAULT 0,
    resolved_complaints INT DEFAULT 0,
    response_time_hours DECIMAL(5,2) DEFAULT 0,
    
    -- System Metrics
    total_api_calls INT DEFAULT 0,
    system_uptime_percentage DECIMAL(5,2) DEFAULT 100,
    error_rate DECIMAL(5,2) DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 3. Customer Analytics
CREATE TABLE customer_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL,
    
    -- Time Period
    period_type VARCHAR(20) NOT NULL, -- 'monthly', 'quarterly', 'yearly', 'lifetime'
    period_start DATE,
    period_end DATE,
    
    -- Purchase Behavior
    total_orders INT DEFAULT 0,
    total_spent DECIMAL(12,2) DEFAULT 0,
    average_order_value DECIMAL(10,2) DEFAULT 0,
    frequency_days DECIMAL(5,1), -- Average days between orders
    
    -- Product Preferences
    favorite_categories JSONB, -- [{"category": "Electronics", "count": 15}]
    favorite_products JSONB,
    
    -- Channel Preferences
    preferred_payment_method VARCHAR(20),
    preferred_delivery_time VARCHAR(20),
    
    -- Engagement
    last_order_date DATE,
    days_since_last_order INT,
    churn_risk_score DECIMAL(5,2), -- 0-100, higher = more likely to churn
    
    -- Demographics & Segmentation
    segment VARCHAR(20), -- 'vip', 'loyal', 'new', 'at_risk', 'churned'
    location_zone VARCHAR(10),
    
    -- Predictions
    predicted_ltv DECIMAL(12,2), -- Lifetime Value
    predicted_next_order_date DATE,
    predicted_churn_date DATE,
    
    calculated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(customer_id, period_type, period_start, period_end)
);

-- 4. Product Analytics
CREATE TABLE product_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    
    -- Time Period
    period_type VARCHAR(20) NOT NULL, -- 'daily', 'weekly', 'monthly', 'quarterly'
    period_start DATE,
    period_end DATE,
    
    -- Sales Performance
    units_sold INT DEFAULT 0,
    revenue_generated DECIMAL(12,2) DEFAULT 0,
    total_orders INT DEFAULT 0,
    
    -- Inventory Performance
    stock_turnover DECIMAL(5,2) DEFAULT 0,
    days_in_stock INT DEFAULT 0,
    stockout_days INT DEFAULT 0,
    
    -- Customer Behavior
    unique_customers INT DEFAULT 0,
    repeat_purchase_rate DECIMAL(5,2) DEFAULT 0,
    return_rate DECIMAL(5,2) DEFAULT 0,
    
    -- Profitability
    gross_profit DECIMAL(12,2) DEFAULT 0,
    profit_margin DECIMAL(5,2) DEFAULT 0,
    
    -- Market Position
    rank_by_revenue INT,
    rank_by_volume INT,
    market_share DECIMAL(5,2),
    
    calculated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(product_id, period_type, period_start, period_end)
);

-- 5. Route Performance Analytics
CREATE TABLE route_performance_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    route_code VARCHAR(10) NOT NULL,
    date DATE NOT NULL,
    
    -- Delivery Performance
    total_deliveries INT DEFAULT 0,
    successful_deliveries INT DEFAULT 0,
    failed_deliveries INT DEFAULT 0,
    success_rate DECIMAL(5,2) DEFAULT 0,
    
    -- Time & Efficiency
    average_delivery_time INT, -- minutes per delivery
    total_route_time INT, -- total minutes for route
    efficiency_score DECIMAL(5,2), -- deliveries per hour
    
    -- Distance & Cost
    total_distance DECIMAL(8,2), -- kilometers
    fuel_consumed DECIMAL(6,2), -- liters
    fuel_cost DECIMAL(8,2),
    cost_per_delivery DECIMAL(8,2),
    cost_per_km DECIMAL(6,3),
    
    -- Revenue & Profitability
    total_revenue DECIMAL(12,2),
    cod_collected DECIMAL(12,2),
    profit_margin DECIMAL(5,2),
    
    -- Customer Satisfaction
    customer_ratings DECIMAL(3,2),
    total_ratings INT,
    complaints INT,
    
    -- Vehicle & Driver
    vehicles_used INT,
    drivers_used INT,
    vehicle_utilization DECIMAL(5,2),
    
    calculated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(route_code, date)
);

-- 6. Driver Performance Analytics
CREATE TABLE driver_performance_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL,
    
    -- Time Period
    period_type VARCHAR(20) NOT NULL, -- 'daily', 'weekly', 'monthly'
    period_start DATE,
    period_end DATE,
    
    -- Delivery Performance
    total_deliveries INT DEFAULT 0,
    successful_deliveries INT DEFAULT 0,
    failed_deliveries INT DEFAULT 0,
    success_rate DECIMAL(5,2) DEFAULT 0,
    
    -- Time Management
    total_working_hours DECIMAL(5,2),
    total_driving_time INT, -- minutes
    total_delivery_time INT, -- minutes
    average_time_per_delivery INT, -- minutes
    
    -- Customer Service
    customer_ratings DECIMAL(3,2),
    total_ratings INT,
    complaints INT,
    compliments INT,
    
    -- Efficiency Metrics
    deliveries_per_hour DECIMAL(4,2),
    kilometers_driven DECIMAL(8,2),
    fuel_efficiency DECIMAL(5,2), -- km per liter
    
    -- Financial Performance
    cod_collected DECIMAL(12,2),
    revenue_generated DECIMAL(12,2),
    wage_earned DECIMAL(10,2),
    
    -- Safety & Compliance
    accidents INT DEFAULT 0,
    traffic_violations INT DEFAULT 0,
    safety_score DECIMAL(5,2) DEFAULT 100,
    
    calculated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(driver_id, period_type, period_start, period_end)
);

-- 7. Financial KPIs
CREATE TABLE financial_kpis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Time Period
    period_type VARCHAR(20) NOT NULL, -- 'daily', 'weekly', 'monthly', 'quarterly', 'yearly'
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    
    -- Revenue Metrics
    gross_revenue DECIMAL(12,2) DEFAULT 0,
    net_revenue DECIMAL(12,2) DEFAULT 0,
    recurring_revenue DECIMAL(12,2) DEFAULT 0,
    revenue_growth_rate DECIMAL(5,2) DEFAULT 0,
    
    -- Cost Metrics
    cost_of_goods_sold DECIMAL(12,2) DEFAULT 0,
    operating_expenses DECIMAL(12,2) DEFAULT 0,
    delivery_costs DECIMAL(12,2) DEFAULT 0,
    vehicle_costs DECIMAL(12,2) DEFAULT 0,
    labor_costs DECIMAL(12,2) DEFAULT 0,
    
    -- Profitability
    gross_profit DECIMAL(12,2) DEFAULT 0,
    gross_profit_margin DECIMAL(5,2) DEFAULT 0,
    net_profit DECIMAL(12,2) DEFAULT 0,
    net_profit_margin DECIMAL(5,2) DEFAULT 0,
    ebitda DECIMAL(12,2) DEFAULT 0,
    
    -- Cash Flow
    cash_inflow DECIMAL(12,2) DEFAULT 0,
    cash_outflow DECIMAL(12,2) DEFAULT 0,
    net_cash_flow DECIMAL(12,2) DEFAULT 0,
    
    -- Customer Metrics
    customer_acquisition_cost DECIMAL(10,2) DEFAULT 0,
    customer_lifetime_value DECIMAL(12,2) DEFAULT 0,
    avg_revenue_per_customer DECIMAL(10,2) DEFAULT 0,
    
    calculated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(period_type, period_start, period_end)
);

-- 8. System Performance Metrics
CREATE TABLE system_performance_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Timestamp (hourly data points)
    timestamp TIMESTAMP NOT NULL,
    date_partition DATE GENERATED ALWAYS AS (timestamp::DATE) STORED,
    
    -- API Performance
    total_api_requests INT DEFAULT 0,
    successful_requests INT DEFAULT 0,
    failed_requests INT DEFAULT 0,
    average_response_time DECIMAL(8,3), -- milliseconds
    max_response_time DECIMAL(8,3),
    error_rate DECIMAL(5,2),
    
    -- Database Performance
    db_connections_active INT,
    db_connections_max INT,
    db_query_time_avg DECIMAL(8,3), -- milliseconds
    db_slow_queries INT,
    
    -- Resource Usage
    cpu_usage_percent DECIMAL(5,2),
    memory_usage_percent DECIMAL(5,2),
    disk_usage_percent DECIMAL(5,2),
    network_in_mb DECIMAL(10,2),
    network_out_mb DECIMAL(10,2),
    
    -- Service Availability
    uptime_percentage DECIMAL(5,2),
    downtime_minutes INT DEFAULT 0,
    
    -- Business Metrics
    active_users INT,
    concurrent_sessions INT,
    orders_processed INT,
    deliveries_tracked INT,
    
    created_at TIMESTAMP DEFAULT NOW()
) PARTITION BY RANGE (date_partition);

-- 9. Report Definitions
CREATE TABLE report_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    report_type VARCHAR(30), -- 'sales', 'delivery', 'financial', 'operational'
    
    -- Report Configuration
    query_template TEXT NOT NULL, -- SQL template with placeholders
    parameters JSONB, -- {"date_range": "required", "route": "optional"}
    output_format VARCHAR(20) DEFAULT 'json', -- 'json', 'csv', 'pdf', 'excel'
    
    -- Scheduling
    is_scheduled BOOLEAN DEFAULT false,
    schedule_cron VARCHAR(50), -- "0 8 * * 1" for weekly Monday 8 AM
    timezone VARCHAR(50) DEFAULT 'Asia/Bangkok',
    
    -- Access Control
    visibility VARCHAR(20) DEFAULT 'private', -- 'public', 'private', 'role_based'
    allowed_roles JSONB, -- ["admin", "manager"]
    created_by_user_id UUID,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    last_run_at TIMESTAMP,
    next_run_at TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 10. Report Executions
CREATE TABLE report_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_definition_id UUID REFERENCES report_definitions(id),
    
    -- Execution Details
    execution_type VARCHAR(20), -- 'manual', 'scheduled', 'api'
    parameters JSONB, -- Actual parameters used
    
    -- Results
    status VARCHAR(20) DEFAULT 'running', -- 'running', 'completed', 'failed', 'cancelled'
    result_data JSONB,
    file_url TEXT, -- If exported to file
    row_count INT,
    
    -- Performance
    execution_time_ms INT,
    memory_used_mb DECIMAL(8,2),
    
    -- User Context
    executed_by_user_id UUID,
    ip_address VARCHAR(45),
    
    -- Timestamps
    started_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    
    -- Error Handling
    error_message TEXT,
    error_details JSONB,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- ================================================================
-- PARTITIONED TABLES (for time-series data)
-- ================================================================

-- Create monthly partitions for analytics_events
CREATE TABLE analytics_events_202507 PARTITION OF analytics_events
    FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');

CREATE TABLE analytics_events_202508 PARTITION OF analytics_events
    FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');

-- Create monthly partitions for system_performance_metrics
CREATE TABLE system_performance_metrics_202507 PARTITION OF system_performance_metrics
    FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');

CREATE TABLE system_performance_metrics_202508 PARTITION OF system_performance_metrics
    FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');

-- ================================================================
-- INDEXES FOR ANALYTICS & REPORTING
-- ================================================================

-- Analytics Events indexes
CREATE INDEX idx_analytics_events_type ON analytics_events(event_type, event_timestamp);
CREATE INDEX idx_analytics_events_entity ON analytics_events(entity_type, entity_id);
CREATE INDEX idx_analytics_events_user ON analytics_events(user_id, event_timestamp);
CREATE INDEX idx_analytics_events_customer ON analytics_events(customer_id, event_timestamp);
CREATE INDEX idx_analytics_events_category ON analytics_events(event_category, event_action);

-- Daily Business Metrics indexes
CREATE INDEX idx_daily_metrics_date ON daily_business_metrics(date);

-- Customer Analytics indexes
CREATE INDEX idx_customer_analytics_customer ON customer_analytics(customer_id, period_type);
CREATE INDEX idx_customer_analytics_segment ON customer_analytics(segment);
CREATE INDEX idx_customer_analytics_churn ON customer_analytics(churn_risk_score);

-- Product Analytics indexes
CREATE INDEX idx_product_analytics_product ON product_analytics(product_id, period_type);
CREATE INDEX idx_product_analytics_revenue ON product_analytics(revenue_generated);
CREATE INDEX idx_product_analytics_period ON product_analytics(period_start, period_end);

-- Route Performance indexes
CREATE INDEX idx_route_performance_route_date ON route_performance_analytics(route_code, date);
CREATE INDEX idx_route_performance_success ON route_performance_analytics(success_rate);

-- Driver Performance indexes
CREATE INDEX idx_driver_performance_driver ON driver_performance_analytics(driver_id, period_type);
CREATE INDEX idx_driver_performance_rating ON driver_performance_analytics(customer_ratings);

-- Financial KPIs indexes
CREATE INDEX idx_financial_kpis_period ON financial_kpis(period_type, period_start, period_end);
CREATE INDEX idx_financial_kpis_revenue ON financial_kpis(net_revenue);

-- System Performance indexes
CREATE INDEX idx_system_performance_timestamp ON system_performance_metrics(timestamp);
CREATE INDEX idx_system_performance_error_rate ON system_performance_metrics(error_rate);

-- Report Definitions indexes
CREATE INDEX idx_report_definitions_type ON report_definitions(report_type);
CREATE INDEX idx_report_definitions_scheduled ON report_definitions(is_scheduled, next_run_at);

-- Report Executions indexes
CREATE INDEX idx_report_executions_definition ON report_executions(report_definition_id);
CREATE INDEX idx_report_executions_status ON report_executions(status);
CREATE INDEX idx_report_executions_user ON report_executions(executed_by_user_id, started_at);
