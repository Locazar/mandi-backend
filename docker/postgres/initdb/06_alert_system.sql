-- Alert System Tables
-- This file is auto-executed when the Postgres container starts

CREATE TABLE IF NOT EXISTS alert_templates (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) DEFAULT 'info',
    priority INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    enabled BOOLEAN DEFAULT true,
    conditions JSONB DEFAULT '{}',
    actions JSONB DEFAULT '{}',
    frequency_config JSONB DEFAULT '{"type": "daily", "limit": 1}',
    validity JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

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

CREATE TABLE IF NOT EXISTS alert_actions (
    id SERIAL PRIMARY KEY,
    alert_id INTEGER NOT NULL,
    label VARCHAR(100) NOT NULL,
    action_url VARCHAR(500),
    action_type VARCHAR(50) NOT NULL,
    payload JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_alert_actions_alert_id FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS seller_alert_logs (
    id SERIAL PRIMARY KEY,
    seller_id INTEGER NOT NULL,
    alert_key VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    action_url VARCHAR(500),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
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

-- Insert sample alert templates
INSERT INTO alert_templates (key, title, description, type, priority, is_active, enabled, conditions, actions, frequency_config)
VALUES
    ('db_high_product_price', 'Price Alert', 'You have products with unusually high prices', 'warning', 1, true, true, 
     '{"conditions": [{"field": "product_count", "operator": ">", "value": 5}], "logic": "AND"}',
     '[{"label": "Review Pricing", "action_type": "navigate", "action_url": "/seller/products"}]',
     '{"type": "weekly", "limit": 1}'),
    
    ('db_low_sales_velocity', 'Sales Performance', 'Your sales velocity is lower than average', 'info', 1, true, true,
     '{"conditions": [{"field": "product_count", "operator": ">", "value": 0}], "logic": "AND"}',
     '[{"label": "View Analytics", "action_type": "navigate", "action_url": "/seller/analytics"}]',
     '{"type": "daily", "limit": 1}'),
    
    ('db_pending_verification', 'Verification Pending', 'Your shop verification is pending review', 'warning', 2, true, true,
     '{"conditions": [{"field": "is_verified", "operator": "=", "value": false}], "logic": "AND"}',
     '[{"label": "Check Status", "action_type": "navigate", "action_url": "/seller/verification/status"}]',
     '{"type": "weekly", "limit": 1}'),
    
    ('db_complete_profile', 'Complete Profile', 'Complete your seller profile for better visibility', 'info', 0, true, true,
     '{"conditions": [{"field": "shop_name", "operator": "=", "value": ""}], "logic": "OR"}',
     '[{"label": "Edit Profile", "action_type": "navigate", "action_url": "/seller/profile"}]',
     '{"type": "weekly", "limit": 1}'),
    
    ('db_update_shop_photo', 'Update Shop Photo', 'Your shop photo is outdated or missing', 'warning', 1, true, true,
     '{"conditions": [{"field": "has_shop_photo", "operator": "=", "value": false}], "logic": "AND"}',
     '[{"label": "Upload Photo", "action_type": "navigate", "action_url": "/seller/shop/edit"}]',
     '{"type": "daily", "limit": 1}')
ON CONFLICT (key) DO NOTHING;
