-- Create alert_templates table if it doesn't exist
CREATE TABLE IF NOT EXISTS alert_templates (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) DEFAULT 'info', -- info, warning, critical
    priority INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    enabled BOOLEAN DEFAULT true,
    conditions JSONB DEFAULT '{}',
    actions JSONB DEFAULT '{}',
    frequency_config JSONB DEFAULT '{"type": "daily", "limit": 1}',
    validity JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_alert_templates_key (key),
    INDEX idx_alert_templates_is_active (is_active)
);

-- Create alerts table if it doesn't exist
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
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_alerts_seller_id (seller_id),
    INDEX idx_alerts_key (key),
    INDEX idx_alerts_valid_from (valid_from),
    INDEX idx_alerts_valid_until (valid_until)
);

-- Create alert_actions table if it doesn't exist
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

-- Create seller_alert_logs table if it doesn't exist
CREATE TABLE IF NOT EXISTS seller_alert_logs (
    id SERIAL PRIMARY KEY,
    seller_id INTEGER NOT NULL,
    alert_key VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL, -- shown, dismissed, clicked
    action_url VARCHAR(500),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_seller_alert_logs_seller_id (seller_id),
    INDEX idx_seller_alert_logs_created_at (created_at)
);

-- Insert sample alert templates for database-driven rules
INSERT INTO alert_templates (key, title, description, type, priority, is_active, enabled, conditions, actions, frequency_config, validity)
VALUES
    ('db_high_product_price', 'Price Alert', 'You have products with unusually high prices', 'warning', 1, true, true, 
     '{"conditions": [{"field": "product_count", "operator": ">", "value": 5}], "logic": "AND"}',
     '[{"label": "Review Pricing", "action_type": "navigate", "action_url": "/seller/products"}]',
     '{"type": "weekly", "limit": 1}',
     '{}'),
    
    ('db_low_sales_velocity', 'Sales Performance', 'Your sales velocity is lower than average', 'info', 1, true, true,
     '{"conditions": [{"field": "product_count", "operator": ">", "value": 0}], "logic": "AND"}',
     '[{"label": "View Analytics", "action_type": "navigate", "action_url": "/seller/analytics"}]',
     '{"type": "daily", "limit": 1}',
     '{}'),
    
    ('db_pending_verification', 'Verification Pending', 'Your shop verification is pending review', 'warning', 2, true, true,
     '{"conditions": [{"field": "is_verified", "operator": "=", "value": false}], "logic": "AND"}',
     '[{"label": "Check Status", "action_type": "navigate", "action_url": "/seller/verification/status"}]',
     '{"type": "weekly", "limit": 1}',
     '{}'),
    
    ('db_complete_profile', 'Complete Profile', 'Complete your seller profile for better visibility', 'info', 0, true, true,
     '{"conditions": [{"field": "shop_name", "operator": "=", "value": ""}], "logic": "OR"}',
     '[{"label": "Edit Profile", "action_type": "navigate", "action_url": "/seller/profile"}]',
     '{"type": "weekly", "limit": 1}',
     '{}'),
    
    ('db_update_shop_photo', 'Update Shop Photo', 'Your shop photo is outdated or missing', 'warning', 1, true, true,
     '{"conditions": [{"field": "has_shop_photo", "operator": "=", "value": false}], "logic": "AND"}',
     '[{"label": "Upload Photo", "action_type": "navigate", "action_url": "/seller/shop/edit"}]',
     '{"type": "daily", "limit": 1}',
     '{}');

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_alert_templates_enabled ON alert_templates(enabled);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at);
CREATE INDEX IF NOT EXISTS idx_seller_alert_logs_alert_key ON seller_alert_logs(alert_key);
