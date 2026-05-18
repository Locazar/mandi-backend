-- ============================================================================
-- Seller Alerts System - Database Setup
-- ============================================================================
-- Run this script in your PostgreSQL database to create all alert tables
-- and populate with sample alert templates
-- ============================================================================

-- 1. Create alert_templates table for database-driven rules
CREATE TABLE IF NOT EXISTS alert_templates (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) DEFAULT 'info', -- info, warning, critical
    priority INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    enabled BOOLEAN DEFAULT true,
    conditions JSONB DEFAULT '{}', -- JSON condition structure for rule evaluation
    actions JSONB DEFAULT '{}', -- JSON actions (buttons, links, etc.)
    frequency_config JSONB DEFAULT '{"type": "daily", "limit": 1}', -- Frequency: once, daily, weekly, monthly
    validity JSONB DEFAULT '{}', -- Time window validity
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Create alerts table for alert instances
CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    seller_id INTEGER NOT NULL,
    key VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    description TEXT,
    type VARCHAR(50) DEFAULT 'info',
    priority INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    valid_from TIMESTAMP,
    valid_until TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    frequency VARCHAR(50),
    last_shown_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Create alert_actions table for actions associated with alerts
CREATE TABLE IF NOT EXISTS alert_actions (
    id SERIAL PRIMARY KEY,
    alert_id INTEGER NOT NULL,
    label VARCHAR(100) NOT NULL,
    action_url VARCHAR(500),
    action_type VARCHAR(50) NOT NULL, -- navigate, api_call, dismiss
    payload JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_alert_actions_alert_id FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE
);

-- 4. Create seller_alert_logs table for audit trail
CREATE TABLE IF NOT EXISTS seller_alert_logs (
    id SERIAL PRIMARY KEY,
    seller_id INTEGER NOT NULL,
    alert_key VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL, -- shown, dismissed, clicked
    action_url VARCHAR(500),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- Create Indexes for Performance
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_alert_templates_key ON alert_templates(key);
CREATE INDEX IF NOT EXISTS idx_alert_templates_enabled ON alert_templates(enabled);
CREATE INDEX IF NOT EXISTS idx_alert_templates_is_active ON alert_templates(is_active);

CREATE INDEX IF NOT EXISTS idx_alerts_seller_id ON alerts(seller_id);
CREATE INDEX IF NOT EXISTS idx_alerts_key ON alerts(key);
CREATE INDEX IF NOT EXISTS idx_alerts_valid_from ON alerts(valid_from);
CREATE INDEX IF NOT EXISTS idx_alerts_valid_until ON alerts(valid_until);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at);

CREATE INDEX IF NOT EXISTS idx_seller_alert_logs_seller_id ON seller_alert_logs(seller_id);
CREATE INDEX IF NOT EXISTS idx_seller_alert_logs_alert_key ON seller_alert_logs(alert_key);
CREATE INDEX IF NOT EXISTS idx_seller_alert_logs_created_at ON seller_alert_logs(created_at);

-- ============================================================================
-- Sample Alert Templates - Database-Driven Rules
-- ============================================================================
-- These templates allow admins to create alerts without modifying code
-- Conditions are evaluated dynamically based on seller's shop data

DELETE FROM alert_templates WHERE key LIKE 'db_%';

INSERT INTO alert_templates (key, title, description, type, priority, is_active, enabled, conditions, actions, frequency_config, validity)
VALUES
    (
        'db_high_product_price',
        'Price Alert',
        'You have products with unusually high prices compared to market average',
        'warning',
        1,
        true,
        true,
        '{"conditions": [{"field": "product_count", "operator": ">", "value": 5}], "logic": "AND"}',
        '[{"label": "Review Pricing", "action_type": "navigate", "action_url": "/seller/products"}]',
        '{"type": "weekly", "limit": 1}',
        '{}'
    ),
    (
        'db_low_sales_velocity',
        'Sales Performance Alert',
        'Your sales velocity is lower than sellers in your category',
        'info',
        1,
        true,
        true,
        '{"conditions": [{"field": "product_count", "operator": ">", "value": 0}], "logic": "AND"}',
        '[{"label": "View Analytics", "action_type": "navigate", "action_url": "/seller/analytics"}]',
        '{"type": "daily", "limit": 1}',
        '{}'
    ),
    (
        'db_pending_verification',
        'Shop Verification Pending',
        'Your shop verification is still pending administrative review',
        'warning',
        2,
        true,
        true,
        '{"conditions": [{"field": "is_verified", "operator": "=", "value": false}], "logic": "AND"}',
        '[{"label": "Check Status", "action_type": "navigate", "action_url": "/seller/verification/status"}, {"label": "Resubmit", "action_type": "navigate", "action_url": "/seller/verification"}]',
        '{"type": "weekly", "limit": 1}',
        '{}'
    ),
    (
        'db_complete_profile',
        'Complete Your Profile',
        'Completing your seller profile increases your visibility and customer trust',
        'info',
        0,
        true,
        true,
        '{"conditions": [{"field": "shop_name", "operator": "=", "value": ""}], "logic": "OR"}',
        '[{"label": "Edit Profile", "action_type": "navigate", "action_url": "/seller/profile"}]',
        '{"type": "weekly", "limit": 1}',
        '{}'
    ),
    (
        'db_update_shop_photo',
        'Update Shop Photo',
        'Your shop photo is outdated or missing. A professional photo increases customer trust',
        'warning',
        1,
        true,
        true,
        '{"conditions": [{"field": "has_shop_photo", "operator": "=", "value": false}], "logic": "AND"}',
        '[{"label": "Upload Photo", "action_type": "navigate", "action_url": "/seller/shop/edit"}]',
        '{"type": "daily", "limit": 1}',
        '{}'
    ),
    (
        'db_inventory_status',
        'Check Inventory',
        'Review your inventory levels to avoid stockouts',
        'info',
        0,
        true,
        true,
        '{"conditions": [{"field": "product_count", "operator": ">", "value": 0}], "logic": "AND"}',
        '[{"label": "Manage Inventory", "action_type": "navigate", "action_url": "/seller/inventory"}]',
        '{"type": "weekly", "limit": 1}',
        '{}'
    ),
    (
        'db_negative_feedback',
        'Customer Feedback Alert',
        'You have received negative feedback. Please review and respond promptly',
        'warning',
        2,
        true,
        true,
        '{"conditions": [{"field": "product_count", "operator": ">", "value": 0}], "logic": "AND"}',
        '[{"label": "View Feedback", "action_type": "navigate", "action_url": "/seller/reviews"}]',
        '{"type": "daily", "limit": 1}',
        '{}'
    ),
    (
        'db_promotion_available',
        'Promotion Opportunity',
        'Limited-time promotions are available to boost your sales',
        'info',
        0,
        true,
        true,
        '{"conditions": [{"field": "is_verified", "operator": "=", "value": true}], "logic": "AND"}',
        '[{"label": "View Promotions", "action_type": "navigate", "action_url": "/seller/promotions"}]',
        '{"type": "weekly", "limit": 1}',
        '{}'
    );

-- ============================================================================
-- Verification: Check if tables were created successfully
-- ============================================================================
-- Uncomment to verify:
-- SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE 'alert%';
-- SELECT * FROM alert_templates;
