-- ============================================================================
-- Migration 008: Alert Flows, enriched AlertTemplate schema
-- ============================================================================

-- 1. Extend alert_templates with new columns
ALTER TABLE alert_templates
    ADD COLUMN IF NOT EXISTS template_type  VARCHAR(50)  DEFAULT 'webview',
    ADD COLUMN IF NOT EXISTS display_type   VARCHAR(50)  DEFAULT 'bottom_sheet',
    ADD COLUMN IF NOT EXISTS content_schema JSONB        DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS fcm_payload    JSONB        DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS flow_key       VARCHAR(100) DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_alert_templates_template_type ON alert_templates(template_type);

-- 2. alert_flows — multi-step guided seller journeys
CREATE TABLE IF NOT EXISTS alert_flows (
    id                SERIAL PRIMARY KEY,
    key               VARCHAR(100) NOT NULL UNIQUE,
    title             VARCHAR(255) NOT NULL,
    description       TEXT,
    trigger_condition JSONB        DEFAULT '{}',
    is_active         BOOLEAN      DEFAULT true,
    created_at        TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_flows_key       ON alert_flows(key);
CREATE INDEX IF NOT EXISTS idx_alert_flows_is_active ON alert_flows(is_active);

-- 3. alert_flow_steps — ordered steps within a flow
CREATE TABLE IF NOT EXISTS alert_flow_steps (
    id          SERIAL PRIMARY KEY,
    flow_id     INTEGER   NOT NULL REFERENCES alert_flows(id) ON DELETE CASCADE,
    step_number INTEGER   NOT NULL,
    title       VARCHAR(255) NOT NULL,
    content     JSONB     DEFAULT '{}',
    actions     JSONB     DEFAULT '[]',
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (flow_id, step_number)
);

CREATE INDEX IF NOT EXISTS idx_alert_flow_steps_flow_id ON alert_flow_steps(flow_id);

-- 4. seller_flow_progress — per-seller progress through a flow
CREATE TABLE IF NOT EXISTS seller_flow_progress (
    id           SERIAL PRIMARY KEY,
    seller_id    INTEGER   NOT NULL,
    flow_key     VARCHAR(100) NOT NULL,
    current_step INTEGER   DEFAULT 1,
    completed_at TIMESTAMP,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (seller_id, flow_key)
);

CREATE INDEX IF NOT EXISTS idx_seller_flow_progress_seller_id ON seller_flow_progress(seller_id);
CREATE INDEX IF NOT EXISTS idx_seller_flow_progress_flow_key  ON seller_flow_progress(flow_key);

-- ============================================================================
-- Seed: sample alert flows
-- ============================================================================

-- "Complete Your Shop" onboarding flow
INSERT INTO alert_flows (key, title, description, trigger_condition, is_active)
VALUES (
    'complete_shop_setup',
    'Complete Your Shop Setup',
    'A guided 3-step journey to set up your shop for success.',
    '{"conditions": [{"field": "is_verified", "operator": "=", "value": false}], "logic": "AND"}',
    true
) ON CONFLICT (key) DO NOTHING;

-- Steps for complete_shop_setup
WITH flow AS (SELECT id FROM alert_flows WHERE key = 'complete_shop_setup')
INSERT INTO alert_flow_steps (flow_id, step_number, title, content, actions)
SELECT
    flow.id,
    step_number,
    title,
    content::jsonb,
    actions::jsonb
FROM flow, (VALUES
    (1, 'Add a Shop Photo',
     '{"description": "Upload a clear photo of your shop. Sellers with photos get 3x more visibility.", "illustration_url": ""}',
     '[{"label": "Upload Photo", "action_type": "navigate", "link": "/seller/shop/edit"}, {"label": "Skip", "action_type": "dismiss"}]'),
    (2, 'Add Your First Product',
     '{"description": "List your first product to start receiving orders from customers.", "illustration_url": ""}',
     '[{"label": "Add Product", "action_type": "navigate", "link": "/seller/products/add"}, {"label": "Later", "action_type": "dismiss"}]'),
    (3, 'Complete Verification',
     '{"description": "Verify your shop to unlock promotions, higher order limits, and a trust badge.", "illustration_url": ""}',
     '[{"label": "Start Verification", "action_type": "navigate", "link": "/seller/verification"}, {"label": "Later", "action_type": "dismiss"}]')
) AS t(step_number, title, content, actions)
ON CONFLICT (flow_id, step_number) DO NOTHING;

-- "Boost Your Sales" engagement flow
INSERT INTO alert_flows (key, title, description, trigger_condition, is_active)
VALUES (
    'boost_sales',
    'Boost Your Sales',
    'Quick wins to increase your store''s visibility and revenue.',
    '{"conditions": [{"field": "product_count", "operator": ">", "value": 0}, {"field": "is_verified", "operator": "=", "value": true}], "logic": "AND"}',
    true
) ON CONFLICT (key) DO NOTHING;

WITH flow AS (SELECT id FROM alert_flows WHERE key = 'boost_sales')
INSERT INTO alert_flow_steps (flow_id, step_number, title, content, actions)
SELECT
    flow.id,
    step_number,
    title,
    content::jsonb,
    actions::jsonb
FROM flow, (VALUES
    (1, 'Apply a Promotion',
     '{"description": "Sellers running promotions see up to 40% more orders.", "illustration_url": ""}',
     '[{"label": "View Promotions", "action_type": "navigate", "link": "/seller/promotions"}, {"label": "Later", "action_type": "dismiss"}]'),
    (2, 'Update Product Images',
     '{"description": "High-quality images can increase conversions by 60%.", "illustration_url": ""}',
     '[{"label": "Manage Products", "action_type": "navigate", "link": "/seller/products"}, {"label": "Later", "action_type": "dismiss"}]')
) AS t(step_number, title, content, actions)
ON CONFLICT (flow_id, step_number) DO NOTHING;

-- ============================================================================
-- Seed: update existing templates with new columns (richer template types)
-- ============================================================================

-- Mark existing onboarding-style templates with appropriate template_type
UPDATE alert_templates SET template_type = 'warning_banner', display_type = 'bottom_sheet'
WHERE key IN ('missing_shop_photo', 'db_update_shop_photo', 'shop_not_verified', 'db_pending_verification');

UPDATE alert_templates SET template_type = 'onboarding_step', display_type = 'bottom_sheet'
WHERE key IN ('no_products_added', 'db_complete_profile');

UPDATE alert_templates SET template_type = 'promo_card', display_type = 'bottom_sheet'
WHERE key IN ('db_promotion_available', 'db_low_sales_velocity');

-- Remaining templates default to 'webview' (no-op, already the default)
