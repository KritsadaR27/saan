-- SaaN Master Migration Script
-- Migration: 000_run_all_migrations.sql
-- Date: 2025-07-02
-- Description: Master script to run all SaaN database migrations in correct order

-- This script will create all tables for the complete SaaN system
-- Run this script on a clean PostgreSQL database

BEGIN;

-- ================================================================
-- MIGRATION ORDER AND DEPENDENCIES
-- ================================================================

-- 1. First, run the vehicles service tables (independent)
\i 001_create_vehicles_service_tables.sql

-- 2. Run inventory service tables (independent)
\i 002_create_inventory_service_tables.sql

-- 3. Run user management tables (independent)
\i 003_create_user_management_tables.sql

-- 4. Run payment & finance tables (references users)
\i 004_create_payment_finance_tables.sql

-- 5. Run analytics & reporting tables (references multiple other tables)
\i 005_create_analytics_reporting_tables.sql

-- 6. Run business features tables (references users, promotions, etc.)
\i 006_create_business_features_tables.sql

-- 7. Run enhanced customer & address tables (references delivery routes)
\i 007_create_enhanced_customer_address_tables.sql

-- 8. Run enhanced orders tables (references customers, products, users)
\i 008_create_enhanced_orders_tables.sql

-- ================================================================
-- ADD FOREIGN KEY CONSTRAINTS BETWEEN SERVICES
-- ================================================================

-- Add foreign key constraints that reference across services
-- (These couldn't be added in individual migration files due to dependencies)

-- Orders → Delivery Rounds (from vehicles service)
ALTER TABLE delivery_assignments 
ADD CONSTRAINT fk_delivery_assignments_order_id 
FOREIGN KEY (order_id) REFERENCES orders(id);

-- Orders → Products (from inventory service)
ALTER TABLE order_items 
ADD CONSTRAINT fk_order_items_product_id 
FOREIGN KEY (product_id) REFERENCES products(id);

ALTER TABLE order_items 
ADD CONSTRAINT fk_order_items_product_variant_id 
FOREIGN KEY (product_variant_id) REFERENCES product_variants(id);

-- Customer Addresses → Delivery Routes
ALTER TABLE customer_addresses 
ADD CONSTRAINT fk_customer_addresses_delivery_route_id 
FOREIGN KEY (delivery_route_id) REFERENCES delivery_routes(id);

-- Payment Transactions → Customers
ALTER TABLE payment_transactions 
ADD CONSTRAINT fk_payment_transactions_customer_id 
FOREIGN KEY (customer_id) REFERENCES customers(id);

-- Invoices → Customers
ALTER TABLE invoices 
ADD CONSTRAINT fk_invoices_customer_id 
FOREIGN KEY (customer_id) REFERENCES customers(id);

-- Invoice Items → Products
ALTER TABLE invoice_items 
ADD CONSTRAINT fk_invoice_items_product_id 
FOREIGN KEY (product_id) REFERENCES products(id);

-- Analytics Events → Users and Customers
ALTER TABLE analytics_events 
ADD CONSTRAINT fk_analytics_events_user_id 
FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE analytics_events 
ADD CONSTRAINT fk_analytics_events_customer_id 
FOREIGN KEY (customer_id) REFERENCES customers(id);

-- Customer Analytics → Customers
ALTER TABLE customer_analytics 
ADD CONSTRAINT fk_customer_analytics_customer_id 
FOREIGN KEY (customer_id) REFERENCES customers(id);

-- Product Analytics → Products
ALTER TABLE product_analytics 
ADD CONSTRAINT fk_product_analytics_product_id 
FOREIGN KEY (product_id) REFERENCES products(id);

-- Driver Performance Analytics → Drivers
ALTER TABLE driver_performance_analytics 
ADD CONSTRAINT fk_driver_performance_analytics_driver_id 
FOREIGN KEY (driver_id) REFERENCES drivers(id);

-- Promotion Usages → Customers and Orders
ALTER TABLE promotion_usages 
ADD CONSTRAINT fk_promotion_usages_customer_id 
FOREIGN KEY (customer_id) REFERENCES customers(id);

ALTER TABLE promotion_usages 
ADD CONSTRAINT fk_promotion_usages_order_id 
FOREIGN KEY (order_id) REFERENCES orders(id);

-- Customer Feedback → Customers
ALTER TABLE customer_feedback 
ADD CONSTRAINT fk_customer_feedback_customer_id 
FOREIGN KEY (customer_id) REFERENCES customers(id);

-- Loyalty Points → Customers
ALTER TABLE customer_loyalty_points 
ADD CONSTRAINT fk_customer_loyalty_points_customer_id 
FOREIGN KEY (customer_id) REFERENCES customers(id);

ALTER TABLE loyalty_point_transactions 
ADD CONSTRAINT fk_loyalty_point_transactions_customer_id 
FOREIGN KEY (customer_id) REFERENCES customers(id);

-- ================================================================
-- CREATE VIEWS FOR COMMON QUERIES
-- ================================================================

-- View: Complete Order Information
CREATE VIEW v_orders_complete AS
SELECT 
    o.*,
    c.first_name,
    c.last_name,
    c.phone as customer_phone,
    c.email as customer_email,
    c.tier as customer_tier,
    ca.address_line1,
    ca.subdistrict,
    ca.district,
    ca.province,
    ca.postal_code,
    dr.route_name,
    dr.delivery_days,
    u_created.first_name as created_by_name,
    u_sales.first_name as sales_rep_name
FROM orders o
LEFT JOIN customers c ON o.customer_id = c.id
LEFT JOIN customer_addresses ca ON o.customer_address_id = ca.id
LEFT JOIN delivery_routes dr ON ca.delivery_route_id = dr.id
LEFT JOIN users u_created ON o.created_by_user_id = u_created.id
LEFT JOIN users u_sales ON o.sales_rep_id = u_sales.id;

-- View: Customer Summary
CREATE VIEW v_customers_summary AS
SELECT 
    c.*,
    clp.current_balance as loyalty_points,
    clp.current_tier as loyalty_tier,
    COUNT(DISTINCT ca.id) as total_addresses,
    COUNT(DISTINCT o.id) as total_orders,
    SUM(o.total_amount) as lifetime_value,
    MAX(o.order_date) as last_order_date,
    AVG(o.total_amount) as average_order_value
FROM customers c
LEFT JOIN customer_loyalty_points clp ON c.id = clp.customer_id
LEFT JOIN customer_addresses ca ON c.id = ca.customer_id
LEFT JOIN orders o ON c.id = o.customer_id AND o.status != 'cancelled'
GROUP BY c.id, clp.current_balance, clp.current_tier;

-- View: Product Performance
CREATE VIEW v_products_performance AS
SELECT 
    p.*,
    pc.name as category_name,
    s.name as supplier_name,
    COALESCE(SUM(oi.quantity_ordered), 0) as total_sold,
    COALESCE(SUM(oi.line_total), 0) as total_revenue,
    COUNT(DISTINCT oi.order_id) as order_count,
    COUNT(DISTINCT o.customer_id) as unique_customers,
    AVG(oi.unit_price) as average_selling_price
FROM products p
LEFT JOIN product_categories pc ON p.category_id = pc.id
LEFT JOIN suppliers s ON p.supplier_id = s.id
LEFT JOIN order_items oi ON p.id = oi.product_id
LEFT JOIN orders o ON oi.order_id = o.id AND o.status != 'cancelled'
GROUP BY p.id, pc.name, s.name;

-- View: Daily Delivery Performance
CREATE VIEW v_daily_delivery_performance AS
SELECT 
    dr.delivery_date,
    dr.route_code,
    COUNT(DISTINCT dr.id) as total_rounds,
    COUNT(DISTINCT da.id) as total_assignments,
    COUNT(CASE WHEN da.delivery_status = 'delivered' THEN 1 END) as successful_deliveries,
    COUNT(CASE WHEN da.delivery_status = 'failed' THEN 1 END) as failed_deliveries,
    ROUND(
        COUNT(CASE WHEN da.delivery_status = 'delivered' THEN 1 END) * 100.0 / 
        NULLIF(COUNT(da.id), 0), 
        2
    ) as success_rate,
    SUM(dvc.fuel_cost) as total_fuel_cost,
    SUM(dvc.total_distance) as total_distance,
    AVG(dvc.cost_per_delivery) as avg_cost_per_delivery
FROM delivery_rounds dr
LEFT JOIN delivery_assignments da ON dr.id = da.delivery_round_id
LEFT JOIN daily_vehicle_costs dvc ON dr.id = dvc.delivery_round_id
GROUP BY dr.delivery_date, dr.route_code;

-- ================================================================
-- CREATE FUNCTIONS FOR COMMON BUSINESS LOGIC
-- ================================================================

-- Function: Calculate Customer Tier based on spending
CREATE OR REPLACE FUNCTION calculate_customer_tier(customer_total_spent DECIMAL)
RETURNS VARCHAR(20) AS $$
BEGIN
    IF customer_total_spent >= 100000 THEN
        RETURN 'vip';
    ELSIF customer_total_spent >= 50000 THEN
        RETURN 'gold';
    ELSIF customer_total_spent >= 10000 THEN
        RETURN 'silver';
    ELSE
        RETURN 'bronze';
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Function: Update customer statistics
CREATE OR REPLACE FUNCTION update_customer_statistics(customer_uuid UUID)
RETURNS VOID AS $$
BEGIN
    UPDATE customers 
    SET 
        total_orders = (
            SELECT COUNT(*) 
            FROM orders 
            WHERE customer_id = customer_uuid 
            AND status NOT IN ('cancelled', 'draft')
        ),
        total_spent = (
            SELECT COALESCE(SUM(total_amount), 0) 
            FROM orders 
            WHERE customer_id = customer_uuid 
            AND status NOT IN ('cancelled', 'draft')
        ),
        last_order_date = (
            SELECT MAX(order_date) 
            FROM orders 
            WHERE customer_id = customer_uuid 
            AND status NOT IN ('cancelled', 'draft')
        ),
        updated_at = NOW()
    WHERE id = customer_uuid;
    
    -- Update customer tier based on total spent
    UPDATE customers 
    SET tier = calculate_customer_tier(total_spent)
    WHERE id = customer_uuid;
END;
$$ LANGUAGE plpgsql;

-- Function: Generate next order number
CREATE OR REPLACE FUNCTION generate_order_number()
RETURNS VARCHAR(20) AS $$
DECLARE
    current_date_str VARCHAR(8);
    sequence_num INT;
    order_number VARCHAR(20);
BEGIN
    current_date_str := TO_CHAR(CURRENT_DATE, 'YYYYMMDD');
    
    -- Get next sequence number for today
    SELECT COALESCE(MAX(
        CAST(
            SUBSTRING(order_number FROM '([0-9]+)$') AS INT
        )
    ), 0) + 1
    INTO sequence_num
    FROM orders 
    WHERE order_number LIKE 'ORD-' || current_date_str || '-%';
    
    order_number := 'ORD-' || current_date_str || '-' || LPAD(sequence_num::TEXT, 3, '0');
    
    RETURN order_number;
END;
$$ LANGUAGE plpgsql;

-- ================================================================
-- CREATE TRIGGERS FOR AUTOMATED BUSINESS LOGIC
-- ================================================================

-- Trigger: Update customer statistics when order is modified
CREATE OR REPLACE FUNCTION trigger_update_customer_stats()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        PERFORM update_customer_statistics(NEW.customer_id);
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        PERFORM update_customer_statistics(OLD.customer_id);
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_orders_update_customer_stats
    AFTER INSERT OR UPDATE OR DELETE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION trigger_update_customer_stats();

-- Trigger: Auto-generate order number
CREATE OR REPLACE FUNCTION trigger_generate_order_number()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.order_number IS NULL OR NEW.order_number = '' THEN
        NEW.order_number := generate_order_number();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_orders_generate_number
    BEFORE INSERT ON orders
    FOR EACH ROW
    EXECUTE FUNCTION trigger_generate_order_number();

-- Trigger: Update timestamps
CREATE OR REPLACE FUNCTION trigger_update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply update timestamp trigger to key tables
CREATE TRIGGER trigger_customers_updated_at
    BEFORE UPDATE ON customers
    FOR EACH ROW
    EXECUTE FUNCTION trigger_update_timestamp();

CREATE TRIGGER trigger_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION trigger_update_timestamp();

CREATE TRIGGER trigger_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION trigger_update_timestamp();

-- ================================================================
-- INSERT INITIAL DATA
-- ================================================================

-- Insert default payment methods
INSERT INTO payment_methods (name, code, type, is_active) VALUES
('Cash', 'CASH', 'cash', true),
('Bank Transfer', 'TRANSFER', 'bank_transfer', true),
('QR Code Payment', 'QR', 'digital_wallet', true),
('Cash on Delivery', 'COD', 'cash', true);

-- Insert default delivery time slots
INSERT INTO delivery_time_slots (name, code, start_time, end_time, available_days) VALUES
('Morning Delivery', 'MORNING', '08:00', '12:00', '["monday", "tuesday", "wednesday", "thursday", "friday", "saturday"]'),
('Afternoon Delivery', 'AFTERNOON', '13:00', '17:00', '["monday", "tuesday", "wednesday", "thursday", "friday", "saturday"]'),
('Evening Delivery', 'EVENING', '18:00', '20:00', '["monday", "tuesday", "wednesday", "thursday", "friday"]');

-- Insert default business settings
INSERT INTO business_settings (key, value, value_type, category, description) VALUES
('company_name', 'SaaN Trading Co., Ltd.', 'string', 'company', 'Company name'),
('company_tax_id', '0123456789012', 'string', 'company', 'Company tax ID'),
('default_tax_rate', '7.00', 'number', 'tax', 'Default VAT rate (%)'),
('base_delivery_fee', '50.00', 'number', 'delivery', 'Base delivery fee (THB)'),
('free_delivery_threshold', '1000.00', 'number', 'delivery', 'Free delivery minimum order amount'),
('loyalty_points_per_baht', '1.00', 'number', 'loyalty', 'Loyalty points earned per THB spent'),
('business_hours_start', '08:00', 'string', 'operations', 'Business hours start time'),
('business_hours_end', '18:00', 'string', 'operations', 'Business hours end time');

-- Insert default notification templates
INSERT INTO notification_templates (name, code, subject, body_text, sms_text, trigger_event) VALUES
('Order Confirmation', 'order_confirmed', 'Order Confirmation - {{order_number}}', 
 'Dear {{customer_name}}, your order {{order_number}} has been confirmed. Total amount: {{total_amount}} THB. Delivery date: {{delivery_date}}.', 
 'Order {{order_number}} confirmed. Amount: {{total_amount}} THB. Delivery: {{delivery_date}}', 
 'order_created'),
('Delivery Scheduled', 'delivery_scheduled', 'Delivery Scheduled - {{order_number}}', 
 'Your order {{order_number}} has been scheduled for delivery on {{delivery_date}} during {{time_slot}}.', 
 'Delivery scheduled: {{delivery_date}} {{time_slot}} for order {{order_number}}', 
 'delivery_assigned'),
('Order Delivered', 'order_delivered', 'Order Delivered - {{order_number}}', 
 'Your order {{order_number}} has been successfully delivered. Thank you for choosing SaaN!', 
 'Order {{order_number}} delivered successfully. Thank you!', 
 'delivery_completed');

COMMIT;

-- ================================================================
-- MIGRATION COMPLETED SUCCESSFULLY
-- ================================================================

-- Display completion message
DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'SaaN Database Migration Completed Successfully!';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Total Tables Created: ~80 tables';
    RAISE NOTICE 'Services Included:';
    RAISE NOTICE '  ✓ Vehicles Service (9 tables)';
    RAISE NOTICE '  ✓ Inventory Service (10 tables)';  
    RAISE NOTICE '  ✓ User Management (10 tables)';
    RAISE NOTICE '  ✓ Payment & Finance (10 tables)';
    RAISE NOTICE '  ✓ Analytics & Reporting (10 tables)';
    RAISE NOTICE '  ✓ Business Features (12 tables)';
    RAISE NOTICE '  ✓ Customer & Address Management (8 tables)';
    RAISE NOTICE '  ✓ Orders & Order Management (9 tables)';
    RAISE NOTICE '  ✓ Thai Address Reference (3 tables)';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Features Available:';
    RAISE NOTICE '  ✓ Complete order management';
    RAISE NOTICE '  ✓ Customer loyalty program';
    RAISE NOTICE '  ✓ Inventory tracking';
    RAISE NOTICE '  ✓ Vehicle & delivery management';
    RAISE NOTICE '  ✓ Payment processing';
    RAISE NOTICE '  ✓ Analytics & reporting';
    RAISE NOTICE '  ✓ User management & permissions';
    RAISE NOTICE '  ✓ Promotions & discounts';
    RAISE NOTICE '  ✓ Multi-address customer support';
    RAISE NOTICE '  ✓ Real-time delivery tracking';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Next Steps:';
    RAISE NOTICE '  1. Import Thai address data';
    RAISE NOTICE '  2. Configure delivery routes';
    RAISE NOTICE '  3. Set up initial users';
    RAISE NOTICE '  4. Import product catalog';
    RAISE NOTICE '  5. Configure payment gateways';
    RAISE NOTICE '========================================';
END $$;
