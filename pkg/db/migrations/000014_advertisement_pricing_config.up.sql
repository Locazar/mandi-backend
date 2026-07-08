CREATE TABLE IF NOT EXISTS advertisement_price_plans (
    id                VARCHAR(32) PRIMARY KEY,
    plan_key          VARCHAR(20) NOT NULL UNIQUE,
    name              VARCHAR(100) NOT NULL,
    description       TEXT,
    rate_per_day_minor BIGINT     NOT NULL,
    sort_order        INT         NOT NULL DEFAULT 0,
    is_active         BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS advertisement_pricing_config (
    id                   VARCHAR(32) PRIMARY KEY,
    gst_rate_percent     NUMERIC(5,2) NOT NULL DEFAULT 18.00,
    platform_fee_percent NUMERIC(5,2) NOT NULL DEFAULT 2.00,
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Seed with the previously hardcoded values so behaviour is unchanged.
INSERT INTO advertisement_price_plans (id, plan_key, name, description, rate_per_day_minor, sort_order, is_active)
VALUES
    ('advp_seed_high',   'high',   'Premium',  'Top banner slot with highest visibility',   50000, 1, TRUE),
    ('advp_seed_medium', 'medium', 'Standard', 'Regular rotation in the banner carousel',   30000, 2, TRUE),
    ('advp_seed_low',    'low',    'Basic',    'Shown when premium slots are free',         15000, 3, TRUE)
ON CONFLICT (plan_key) DO NOTHING;

INSERT INTO advertisement_pricing_config (id, gst_rate_percent, platform_fee_percent)
VALUES ('advcfg_default', 18.00, 2.00)
ON CONFLICT (id) DO NOTHING;
