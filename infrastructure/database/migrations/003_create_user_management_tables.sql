-- SaaN User Management & Authentication Service - Complete Database Schema
-- Migration: 003_create_user_management_tables.sql
-- Date: 2025-07-02

-- ================================================================
-- USER MANAGEMENT & AUTHENTICATION TABLES
-- ================================================================

-- 1. Users (System Users - Admin, Staff, etc.)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    
    -- Authentication
    password_hash VARCHAR(255) NOT NULL,
    salt VARCHAR(100),
    
    -- Personal Information
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    display_name VARCHAR(200),
    avatar_url TEXT,
    
    -- Role & Permissions
    role VARCHAR(20) NOT NULL, -- 'super_admin', 'admin', 'manager', 'staff', 'driver', 'sales'
    department VARCHAR(50), -- 'logistics', 'sales', 'finance', 'operations', 'it'
    employee_id VARCHAR(20) UNIQUE,
    
    -- Account Status
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'inactive', 'suspended', 'terminated'
    is_verified BOOLEAN DEFAULT false,
    email_verified_at TIMESTAMP,
    
    -- Security
    last_login_at TIMESTAMP,
    last_login_ip VARCHAR(45),
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMP,
    
    -- Two-Factor Authentication
    two_factor_enabled BOOLEAN DEFAULT false,
    two_factor_secret VARCHAR(100),
    backup_codes JSONB,
    
    -- Preferences
    language VARCHAR(10) DEFAULT 'th',
    timezone VARCHAR(50) DEFAULT 'Asia/Bangkok',
    preferences JSONB, -- UI preferences, notifications, etc.
    
    -- Timestamps
    hire_date DATE,
    termination_date DATE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- 2. User Sessions
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    
    -- Session Details
    token_hash VARCHAR(255) NOT NULL,
    refresh_token_hash VARCHAR(255),
    
    -- Device & Browser Info
    device_info JSONB, -- {"browser": "Chrome", "os": "macOS", "device": "Desktop"}
    ip_address VARCHAR(45),
    user_agent TEXT,
    
    -- Session Status
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP NOT NULL,
    last_used_at TIMESTAMP DEFAULT NOW(),
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 3. User Permissions
CREATE TABLE user_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    
    -- Permission Details
    resource VARCHAR(50) NOT NULL, -- 'orders', 'customers', 'vehicles', 'inventory'
    action VARCHAR(20) NOT NULL, -- 'create', 'read', 'update', 'delete', 'approve'
    scope VARCHAR(20) DEFAULT 'all', -- 'all', 'own', 'department', 'limited'
    
    -- Conditions (JSON for complex permissions)
    conditions JSONB, -- {"department": "sales", "max_amount": 10000}
    
    -- Metadata
    granted_by_user_id UUID REFERENCES users(id),
    granted_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 4. Audit Logs (User Activity Tracking)
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    
    -- Action Details
    action VARCHAR(50) NOT NULL, -- 'login', 'logout', 'create_order', 'update_customer'
    resource_type VARCHAR(50), -- 'order', 'customer', 'vehicle', 'product'
    resource_id UUID,
    
    -- Request Details
    method VARCHAR(10), -- 'GET', 'POST', 'PUT', 'DELETE'
    endpoint VARCHAR(200),
    ip_address VARCHAR(45),
    user_agent TEXT,
    
    -- Changes (for update operations)
    old_values JSONB,
    new_values JSONB,
    
    -- Metadata
    description TEXT,
    severity VARCHAR(20) DEFAULT 'info', -- 'critical', 'warning', 'info', 'debug'
    
    -- Timestamp
    created_at TIMESTAMP DEFAULT NOW()
);

-- 5. Password Reset Tokens
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    
    -- Token Details
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    
    -- Status
    is_used BOOLEAN DEFAULT false,
    used_at TIMESTAMP,
    
    -- Request Info
    ip_address VARCHAR(45),
    user_agent TEXT,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 6. Email Verification Tokens
CREATE TABLE email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    
    -- Token Details
    token_hash VARCHAR(255) NOT NULL,
    email VARCHAR(100) NOT NULL, -- New email to verify
    expires_at TIMESTAMP NOT NULL,
    
    -- Status
    is_verified BOOLEAN DEFAULT false,
    verified_at TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 7. User Notifications
CREATE TABLE user_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    
    -- Notification Details
    type VARCHAR(50) NOT NULL, -- 'order_update', 'low_stock', 'system_alert', 'document_expiry'
    title VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    
    -- Metadata
    data JSONB, -- Additional data for the notification
    priority VARCHAR(20) DEFAULT 'normal', -- 'critical', 'high', 'normal', 'low'
    
    -- Delivery Channels
    channels JSONB, -- ["email", "sms", "push", "in_app"]
    
    -- Status
    is_read BOOLEAN DEFAULT false,
    read_at TIMESTAMP,
    is_sent BOOLEAN DEFAULT false,
    sent_at TIMESTAMP,
    
    -- Scheduling
    scheduled_for TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 8. User Notification Preferences
CREATE TABLE user_notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    
    -- Notification Type
    notification_type VARCHAR(50) NOT NULL,
    
    -- Channel Preferences
    email_enabled BOOLEAN DEFAULT true,
    sms_enabled BOOLEAN DEFAULT false,
    push_enabled BOOLEAN DEFAULT true,
    in_app_enabled BOOLEAN DEFAULT true,
    
    -- Timing Preferences
    quiet_hours_start TIME,
    quiet_hours_end TIME,
    weekend_enabled BOOLEAN DEFAULT false,
    
    -- Frequency
    frequency VARCHAR(20) DEFAULT 'immediate', -- 'immediate', 'hourly', 'daily', 'weekly'
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 9. API Keys (for external integrations)
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    
    -- Key Details
    name VARCHAR(100) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(20) NOT NULL, -- First few chars for identification
    
    -- Permissions
    scopes JSONB, -- ["orders:read", "customers:write", "inventory:read"]
    rate_limit_per_minute INT DEFAULT 60,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP,
    
    -- Usage Stats
    total_requests INT DEFAULT 0,
    last_used_at TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 10. User Activity Summary (for reporting)
CREATE TABLE user_activity_summary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    date DATE NOT NULL,
    
    -- Activity Counts
    login_count INT DEFAULT 0,
    orders_created INT DEFAULT 0,
    orders_updated INT DEFAULT 0,
    customers_created INT DEFAULT 0,
    deliveries_completed INT DEFAULT 0,
    
    -- Time Tracking
    total_session_time INT DEFAULT 0, -- minutes
    first_login_at TIMESTAMP,
    last_logout_at TIMESTAMP,
    
    -- Performance Metrics (for drivers/sales)
    revenue_generated DECIMAL(12,2) DEFAULT 0,
    tasks_completed INT DEFAULT 0,
    customer_satisfaction DECIMAL(3,2),
    
    created_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(user_id, date)
);

-- ================================================================
-- INDEXES FOR USER MANAGEMENT
-- ================================================================

-- Users table indexes
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_department ON users(department);
CREATE INDEX idx_users_employee_id ON users(employee_id);

-- User Sessions indexes
CREATE INDEX idx_user_sessions_user ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token ON user_sessions(token_hash);
CREATE INDEX idx_user_sessions_active ON user_sessions(is_active, expires_at);

-- User Permissions indexes
CREATE INDEX idx_user_permissions_user ON user_permissions(user_id);
CREATE INDEX idx_user_permissions_resource ON user_permissions(resource, action);
CREATE INDEX idx_user_permissions_active ON user_permissions(is_active);

-- Audit Logs indexes
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_date ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_severity ON audit_logs(severity);

-- Password Reset Tokens indexes
CREATE INDEX idx_password_reset_user ON password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_token ON password_reset_tokens(token_hash);
CREATE INDEX idx_password_reset_expiry ON password_reset_tokens(expires_at);

-- Email Verification Tokens indexes
CREATE INDEX idx_email_verification_user ON email_verification_tokens(user_id);
CREATE INDEX idx_email_verification_token ON email_verification_tokens(token_hash);

-- User Notifications indexes
CREATE INDEX idx_user_notifications_user ON user_notifications(user_id);
CREATE INDEX idx_user_notifications_type ON user_notifications(type);
CREATE INDEX idx_user_notifications_read ON user_notifications(is_read);
CREATE INDEX idx_user_notifications_priority ON user_notifications(priority);
CREATE INDEX idx_user_notifications_scheduled ON user_notifications(scheduled_for);

-- User Notification Preferences indexes
CREATE INDEX idx_notification_preferences_user ON user_notification_preferences(user_id);
CREATE INDEX idx_notification_preferences_type ON user_notification_preferences(notification_type);

-- API Keys indexes
CREATE INDEX idx_api_keys_user ON api_keys(user_id);
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_prefix ON api_keys(key_prefix);
CREATE INDEX idx_api_keys_active ON api_keys(is_active);

-- User Activity Summary indexes
CREATE INDEX idx_user_activity_user_date ON user_activity_summary(user_id, date);
CREATE INDEX idx_user_activity_date ON user_activity_summary(date);
