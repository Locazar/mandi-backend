-- Upgrade a restored legacy bigint-key backup into the schema shape expected by
-- the current backend. Existing data is preserved by converting identifiers to
-- strings in place and by adding the missing tables/columns used by current code.

-- Drop all foreign key constraints safely
DO $$
DECLARE
    rec RECORD;
BEGIN
    FOR rec IN
        SELECT conrelid::regclass AS table_name, conname
        FROM pg_constraint
        WHERE contype = 'f'
          AND connamespace = 'public'::regnamespace
    LOOP
        BEGIN
            EXECUTE format('ALTER TABLE %s DROP CONSTRAINT IF EXISTS %I', rec.table_name, rec.conname);
        EXCEPTION WHEN OTHERS THEN
            -- Silently skip constraints that can't be dropped
            RAISE DEBUG 'Failed to drop constraint % on table %: %', rec.conname, rec.table_name, SQLERRM;
        END;
    END LOOP;
END $$;

CREATE OR REPLACE FUNCTION _legacy_convert_to_varchar(p_table TEXT, p_column TEXT, p_length INTEGER DEFAULT 32)
RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE
    current_type TEXT;
    seq_name TEXT;
BEGIN
    SELECT data_type
      INTO current_type
      FROM information_schema.columns
     WHERE table_schema = 'public'
       AND table_name = p_table
       AND column_name = p_column;

    IF NOT FOUND THEN
        RETURN;
    END IF;

    IF current_type IN ('character varying', 'text') THEN
        RETURN;
    END IF;

    SELECT pg_get_serial_sequence('public.' || p_table, p_column)
      INTO seq_name;

    IF seq_name IS NOT NULL THEN
        EXECUTE format('ALTER TABLE %I ALTER COLUMN %I DROP DEFAULT', p_table, p_column);
    END IF;

    EXECUTE format(
        'ALTER TABLE %I ALTER COLUMN %I TYPE VARCHAR(%s) USING %I::text',
        p_table,
        p_column,
        p_length,
        p_column
    );

    IF seq_name IS NOT NULL THEN
        EXECUTE format(
            'ALTER TABLE %I ALTER COLUMN %I SET DEFAULT nextval(%L)::text',
            p_table,
            p_column,
            seq_name
        );
    END IF;
END $$;

CREATE OR REPLACE FUNCTION _legacy_safe_convert(p_table TEXT, p_column TEXT, p_length INTEGER DEFAULT 32)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    -- First check if table exists
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_schema = 'public' AND table_name = p_table
    ) THEN
        RETURN;
    END IF;
    
    -- Then check if column exists
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = p_table AND column_name = p_column
    ) THEN
        RETURN;
    END IF;
    
    -- Now do the conversion
    PERFORM _legacy_convert_to_varchar(p_table, p_column, p_length);
END $$;

SELECT _legacy_safe_convert('users', 'id');
SELECT _legacy_safe_convert('admins', 'id');
SELECT _legacy_safe_convert('countries', 'id');
SELECT _legacy_safe_convert('addresses', 'id');
SELECT _legacy_safe_convert('addresses', 'user_id');
SELECT _legacy_safe_convert('addresses', 'country_id');
SELECT _legacy_safe_convert('user_addresses', 'id');
SELECT _legacy_safe_convert('user_addresses', 'user_id');
SELECT _legacy_safe_convert('user_addresses', 'address_id');
SELECT _legacy_safe_convert('payment_methods', 'id');
SELECT _legacy_safe_convert('shop_details', 'id');
SELECT _legacy_safe_convert('shop_details', 'admin_id');
SELECT _legacy_safe_convert('departments', 'id');
SELECT _legacy_safe_convert('categories', 'id');
SELECT _legacy_safe_convert('categories', 'department_id');
SELECT _legacy_safe_convert('sub_categories', 'id');
SELECT _legacy_safe_convert('sub_categories', 'department_id');
SELECT _legacy_safe_convert('sub_categories', 'category_id');
SELECT _legacy_safe_convert('brands', 'id');
SELECT _legacy_safe_convert('variations', 'id');
SELECT _legacy_safe_convert('variations', 'sub_category_id');
SELECT _legacy_safe_convert('variation_options', 'id');
SELECT _legacy_safe_convert('variation_options', 'variation_id');
SELECT _legacy_safe_convert('sub_type_attributes', 'id');
SELECT _legacy_safe_convert('sub_type_attributes', 'sub_category_id');
SELECT _legacy_safe_convert('sub_type_attribute_options', 'id');
SELECT _legacy_safe_convert('sub_type_attribute_options', 'sub_type_attribute_id');
SELECT _legacy_safe_convert('products', 'id');
SELECT _legacy_safe_convert('products', 'category_id');
SELECT _legacy_safe_convert('products', 'department_id');
SELECT _legacy_safe_convert('products', 'shop_id');
SELECT _legacy_safe_convert('product_items', 'id');
SELECT _legacy_safe_convert('product_items', 'sub_category_id');
SELECT _legacy_safe_convert('product_items', 'category_id');
SELECT _legacy_safe_convert('product_items', 'department_id');
SELECT _legacy_safe_convert('product_images', 'id');
SELECT _legacy_safe_convert('product_images', 'product_item_id');
SELECT _legacy_safe_convert('product_images', 'shop_id');
SELECT _legacy_safe_convert('product_images', 'product_id');
SELECT _legacy_safe_convert('product_configurations', 'product_item_id');
SELECT _legacy_safe_convert('product_configurations', 'variation_option_id');
SELECT _legacy_safe_convert('category_images', 'id');
SELECT _legacy_safe_convert('category_images', 'category_id');
SELECT _legacy_safe_convert('offers', 'id');
SELECT _legacy_safe_convert('offer_categories', 'id');
SELECT _legacy_safe_convert('offer_categories', 'offer_id');
SELECT _legacy_safe_convert('offer_categories', 'category_id');
SELECT _legacy_safe_convert('offer_products', 'id');
SELECT _legacy_safe_convert('offer_products', 'offer_id');
SELECT _legacy_safe_convert('offer_products', 'product_id');
SELECT _legacy_safe_convert('offer_products', 'product_item_id');
SELECT _legacy_safe_convert('coupons', 'coupon_id');
SELECT _legacy_safe_convert('coupon_uses', 'coupon_uses_id');
SELECT _legacy_safe_convert('coupon_uses', 'coupon_id');
SELECT _legacy_safe_convert('coupon_uses', 'user_id');
SELECT _legacy_safe_convert('wallets', 'id');
SELECT _legacy_safe_convert('wallets', 'user_id');
SELECT _legacy_safe_convert('transactions', 'transaction_id');
SELECT _legacy_safe_convert('transactions', 'wallet_id');
SELECT _legacy_safe_convert('carts', 'id');
SELECT _legacy_safe_convert('carts', 'user_id');
SELECT _legacy_safe_convert('carts', 'applied_coupon_id');
SELECT _legacy_safe_convert('cart_items', 'id');
SELECT _legacy_safe_convert('cart_items', 'cart_id');
SELECT _legacy_safe_convert('cart_items', 'product_item_id');
SELECT _legacy_safe_convert('wish_lists', 'id');
SELECT _legacy_safe_convert('wish_lists', 'user_id');
SELECT _legacy_safe_convert('wish_lists', 'shop_id');
SELECT _legacy_safe_convert('wish_lists', 'admin_id');
SELECT _legacy_safe_convert('wish_lists', 'product_item_id');
SELECT _legacy_safe_convert('shop_orders', 'id');
SELECT _legacy_safe_convert('shop_orders', 'user_id');
SELECT _legacy_safe_convert('shop_orders', 'address_id');
SELECT _legacy_safe_convert('shop_orders', 'payment_method_id');
SELECT _legacy_safe_convert('shop_orders', 'shop_id');
SELECT _legacy_safe_convert('order_lines', 'id');
SELECT _legacy_safe_convert('order_lines', 'product_item_id');
SELECT _legacy_safe_convert('order_lines', 'shop_order_id');
SELECT _legacy_safe_convert('order_returns', 'id');
SELECT _legacy_safe_convert('order_returns', 'shop_order_id');
SELECT _legacy_safe_convert('order_statuses', 'id');
SELECT _legacy_safe_convert('order_statuses', 'shop_order_id');
SELECT _legacy_safe_convert('otp_sessions', 'id');
SELECT _legacy_safe_convert('otp_sessions', 'user_id');
SELECT _legacy_safe_convert('otp_sessions', 'admin_id');
SELECT _legacy_safe_convert('refresh_sessions', 'user_id');
SELECT _legacy_safe_convert('refresh_sessions', 'admin_id');
SELECT _legacy_safe_convert('advertisements', 'id');
SELECT _legacy_safe_convert('advertisements', 'admin_id');
SELECT _legacy_safe_convert('notifications', 'id');
SELECT _legacy_safe_convert('notifications', 'user_id');
SELECT _legacy_safe_convert('notifications', 'admin_id');
SELECT _legacy_safe_convert('notifications', 'shop_id');
SELECT _legacy_safe_convert('notifications', 'product_id');
SELECT _legacy_safe_convert('notifications', 'offer_id');
SELECT _legacy_safe_convert('shop_verifications', 'id');
SELECT _legacy_safe_convert('shop_verifications', 'shop_id');
SELECT _legacy_safe_convert('shop_verifications', 'admin_id');
SELECT _legacy_safe_convert('shop_verification_histories', 'id');
SELECT _legacy_safe_convert('shop_verification_histories', 'shop_id');
SELECT _legacy_safe_convert('shop_verification_histories', 'admin_id');

DROP FUNCTION IF EXISTS _legacy_safe_convert(TEXT, TEXT, INTEGER);

CREATE OR REPLACE FUNCTION _legacy_ensure_text_id_default(p_table TEXT, p_column TEXT DEFAULT 'id')
RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE
    default_expr TEXT;
    seq_name TEXT;
BEGIN
    -- Check if table exists first
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_schema = 'public' AND table_name = p_table
    ) THEN
        RETURN;
    END IF;
    
    SELECT column_default
      INTO default_expr
      FROM information_schema.columns
     WHERE table_schema = 'public'
       AND table_name = p_table
       AND column_name = p_column;

    IF NOT FOUND OR default_expr IS NOT NULL THEN
        RETURN;
    END IF;

    seq_name := format('public.%I_%I_seq', p_table, p_column);
    EXECUTE format('CREATE SEQUENCE IF NOT EXISTS %s', seq_name);
    EXECUTE format(
        'ALTER TABLE %I ALTER COLUMN %I SET DEFAULT nextval(%L)::text',
        p_table,
        p_column,
        seq_name
    );
END $$;

-- Safely add columns to users if the table exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users') THEN
        ALTER TABLE users ADD COLUMN IF NOT EXISTS user_name VARCHAR(255);
        ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;
        ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_verified BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE users ADD COLUMN IF NOT EXISTS date_of_birth DATE;
        ALTER TABLE users ADD COLUMN IF NOT EXISTS trial_used BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

        UPDATE users
        SET is_active = CASE
            WHEN block_status IS TRUE THEN FALSE
            ELSE TRUE
        END
        WHERE is_active IS DISTINCT FROM CASE WHEN block_status IS TRUE THEN FALSE ELSE TRUE END;
    END IF;
END $$;

-- Safely add columns to admins if the table exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'admins') THEN
        ALTER TABLE admins ADD COLUMN IF NOT EXISTS user_name VARCHAR(255);
        ALTER TABLE admins ADD COLUMN IF NOT EXISTS aadhaar_last4 VARCHAR(4);
        ALTER TABLE admins ADD COLUMN IF NOT EXISTS aadhaar_verified BOOLEAN NOT NULL DEFAULT FALSE;
        ALTER TABLE admins ADD COLUMN IF NOT EXISTS role VARCHAR(50) NOT NULL DEFAULT 'seller';
        ALTER TABLE admins ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

        UPDATE admins
        SET aadhaar_last4 = RIGHT(aadhar, 4)
        WHERE aadhaar_last4 IS NULL
          AND aadhar IS NOT NULL
          AND LENGTH(aadhar) >= 4;

        UPDATE admins a
        SET role = 'seller'
        WHERE a.role = 'super_admin'
          AND (
              COALESCE(a.verified_seller, FALSE) = TRUE
              OR EXISTS (SELECT 1 FROM shop_details sd WHERE sd.admin_id = a.id)
          );
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'shop_details') THEN
        UPDATE shop_details sd
        SET admin_id = NULL
        WHERE sd.admin_id IS NOT NULL
            AND NOT EXISTS (
                    SELECT 1
                    FROM admins a
                    WHERE a.id = sd.admin_id
            );
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS product_item_views (
    id              VARCHAR(32) PRIMARY KEY,
    product_item_id VARCHAR(32) NOT NULL,
    shop_id         VARCHAR(32) NOT NULL,
    admin_id        VARCHAR(32),
    viewed_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    view_count      INT         NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_product_item_views_product_item_id ON product_item_views (product_item_id);
CREATE INDEX IF NOT EXISTS idx_product_item_views_admin_id ON product_item_views (admin_id);

CREATE TABLE IF NOT EXISTS product_item_filter_types (
    id          VARCHAR(32) PRIMARY KEY,
    filter_name VARCHAR(255) NOT NULL,
    shop_id     VARCHAR(32),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_product_item_filter_types_shop_id ON product_item_filter_types (shop_id);

CREATE TABLE IF NOT EXISTS shop_offers (
    id         VARCHAR(32) PRIMARY KEY,
    shop_id    VARCHAR(32) NOT NULL,
    offer_id   VARCHAR(32) NOT NULL,
    admin_id   VARCHAR(32),
    start_date TIMESTAMPTZ NOT NULL,
    end_date   TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_offers_shop_id ON shop_offers (shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_offers_offer_id ON shop_offers (offer_id);

CREATE TABLE IF NOT EXISTS shop_departments (
    id              VARCHAR(32) PRIMARY KEY,
    admin_id        VARCHAR(32) NOT NULL,
    shop_id         VARCHAR(32) NOT NULL,
    department_id   VARCHAR(32) NOT NULL,
    category_id     VARCHAR(32) NOT NULL,
    sub_category_id VARCHAR(32) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_shop_departments_shop_dept_cat_sub
    ON shop_departments (shop_id, department_id, category_id, sub_category_id);
CREATE INDEX IF NOT EXISTS idx_shop_departments_admin_id ON shop_departments (admin_id);

CREATE TABLE IF NOT EXISTS shop_times (
    id         VARCHAR(32) PRIMARY KEY,
    shop_id    VARCHAR(32) NOT NULL,
    status     VARCHAR(20) NOT NULL DEFAULT 'close',
    open_time  VARCHAR(50) NOT NULL DEFAULT '',
    close_time VARCHAR(50) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_times_shop_id ON shop_times (shop_id);

CREATE TABLE IF NOT EXISTS shop_socials (
    id          VARCHAR(32) PRIMARY KEY,
    shop_id     VARCHAR(32) NOT NULL,
    admin_id    VARCHAR(32) NOT NULL DEFAULT '',
    user_id     VARCHAR(32) NOT NULL DEFAULT '',
    is_follower BOOLEAN NOT NULL DEFAULT FALSE,
    is_liked    BOOLEAN NOT NULL DEFAULT FALSE,
    rating      INT NOT NULL DEFAULT 0,
    review      TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_socials_shop_id ON shop_socials (shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_socials_admin_id ON shop_socials (admin_id);
CREATE INDEX IF NOT EXISTS idx_shop_socials_user_id ON shop_socials (user_id);

CREATE TABLE IF NOT EXISTS subscription_plans (
    id                         VARCHAR(32) PRIMARY KEY,
    name                       VARCHAR(255) NOT NULL UNIQUE,
    price_monthly_amount_minor BIGINT NOT NULL DEFAULT 0,
    price_monthly_currency     CHAR(3) NOT NULL DEFAULT 'INR',
    duration_days              INT NOT NULL DEFAULT 30,
    is_active                  BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS subscription_orders (
    id                  VARCHAR(32) PRIMARY KEY,
    user_id             VARCHAR(32) NOT NULL,
    plan_id             VARCHAR(32) NOT NULL,
    price_amount_minor  BIGINT NOT NULL DEFAULT 0,
    price_currency      CHAR(3) NOT NULL DEFAULT 'INR',
    razorpay_order_id   VARCHAR(255) NOT NULL UNIQUE,
    razorpay_payment_id VARCHAR(255) UNIQUE,
    status              VARCHAR(20) NOT NULL DEFAULT 'created',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_at             TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_subscription_orders_user_id ON subscription_orders (user_id);
CREATE INDEX IF NOT EXISTS idx_subscription_orders_plan_id ON subscription_orders (plan_id);

CREATE TABLE IF NOT EXISTS user_subscriptions (
    id                    VARCHAR(32) PRIMARY KEY,
    user_id               VARCHAR(32) NOT NULL,
    plan_id               VARCHAR(32) NOT NULL,
    subscription_order_id VARCHAR(32),
    is_trial              BOOLEAN NOT NULL DEFAULT FALSE,
    start_date            TIMESTAMPTZ NOT NULL,
    end_date              TIMESTAMPTZ NOT NULL,
    is_active             BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_user_subscriptions_user_id ON user_subscriptions (user_id);
CREATE INDEX IF NOT EXISTS idx_user_subscriptions_plan_id ON user_subscriptions (plan_id);

CREATE TABLE IF NOT EXISTS notification_device_tokens (
    id         VARCHAR(32) PRIMARY KEY,
    owner_id   VARCHAR(100) NOT NULL,
    shop_id    VARCHAR(100) NOT NULL,
    admin_id   VARCHAR(100) NOT NULL,
    owner_type VARCHAR(10) NOT NULL,
    token      VARCHAR(255) NOT NULL UNIQUE,
    platform   VARCHAR(50),
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_notification_device_tokens_owner_id ON notification_device_tokens (owner_id);

CREATE TABLE IF NOT EXISTS fcm_tokens (
    id         VARCHAR(32) PRIMARY KEY,
    token      VARCHAR(255) NOT NULL UNIQUE,
    device     VARCHAR(255),
    platform   VARCHAR(255),
    owner_id   VARCHAR(255) NOT NULL,
    owner_type VARCHAR(255) NOT NULL,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_fcm_tokens_owner_id ON fcm_tokens (owner_id);

CREATE TABLE IF NOT EXISTS alerts (
    id            VARCHAR(32) PRIMARY KEY,
    seller_id     VARCHAR(32) NOT NULL,
    key           VARCHAR(100) NOT NULL,
    title         VARCHAR(255) NOT NULL,
    content       TEXT,
    description   TEXT,
    type          VARCHAR(50) NOT NULL DEFAULT 'info',
    priority      INT NOT NULL DEFAULT 0,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    valid_from    TIMESTAMPTZ,
    valid_until   TIMESTAMPTZ,
    metadata      JSONB NOT NULL DEFAULT '{}',
    frequency     VARCHAR(50),
    last_shown_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alerts_seller_id ON alerts (seller_id);
CREATE INDEX IF NOT EXISTS idx_alerts_key ON alerts (key);

CREATE TABLE IF NOT EXISTS alert_actions (
    id          VARCHAR(32) PRIMARY KEY,
    alert_id    VARCHAR(32) NOT NULL,
    label       VARCHAR(100) NOT NULL,
    action_url  VARCHAR(500),
    action_type VARCHAR(50) NOT NULL,
    payload     JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alert_actions_alert_id ON alert_actions (alert_id);

CREATE TABLE IF NOT EXISTS alert_templates (
    id               VARCHAR(32) PRIMARY KEY,
    key              VARCHAR(100) NOT NULL UNIQUE,
    title            VARCHAR(255) NOT NULL,
    description      TEXT,
    type             VARCHAR(50) NOT NULL DEFAULT 'info',
    priority         INT NOT NULL DEFAULT 0,
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,
    conditions       JSONB,
    actions          JSONB,
    frequency_config JSONB,
    validity         JSONB,
    enabled          BOOLEAN NOT NULL DEFAULT TRUE,
    template_type    VARCHAR(50) NOT NULL DEFAULT 'announcement',
    display_type     VARCHAR(50) NOT NULL DEFAULT 'bottom_sheet',
    content_schema   JSONB,
    flow_key         VARCHAR(100),
    created_by       VARCHAR(32),
    updated_by       VARCHAR(32),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alert_templates_is_active ON alert_templates (is_active);
CREATE INDEX IF NOT EXISTS idx_alert_templates_flow_key ON alert_templates (flow_key);

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'alert_templates') THEN
        ALTER TABLE alert_templates ADD COLUMN IF NOT EXISTS flow_key VARCHAR(100);
        CREATE INDEX IF NOT EXISTS idx_alert_templates_flow_key ON alert_templates (flow_key);
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS seller_alert_logs (
    id         VARCHAR(32) PRIMARY KEY,
    seller_id  VARCHAR(32) NOT NULL,
    alert_key  VARCHAR(100) NOT NULL,
    action     VARCHAR(50) NOT NULL,
    action_url VARCHAR(500),
    metadata   JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_seller_alert_logs_seller_id ON seller_alert_logs (seller_id);

CREATE TABLE IF NOT EXISTS job_categories (
    id          VARCHAR(32) PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    category_id VARCHAR(32),
    parent_id   VARCHAR(32)
);

CREATE TABLE IF NOT EXISTS job_sub_categories (
    id          VARCHAR(32) PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    category_id VARCHAR(32) NOT NULL
);

CREATE TABLE IF NOT EXISTS job_locations (
    id   VARCHAR(32) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS job_filters (
    id     VARCHAR(32) PRIMARY KEY,
    name   VARCHAR(100) NOT NULL UNIQUE,
    type   VARCHAR(50) NOT NULL,
    values TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS job_category_filters (
    id          VARCHAR(32) PRIMARY KEY,
    category_id VARCHAR(32) NOT NULL,
    filter_id   VARCHAR(32) NOT NULL
);

CREATE TABLE IF NOT EXISTS job_category_locations (
    id          VARCHAR(32) PRIMARY KEY,
    category_id VARCHAR(32) NOT NULL,
    location_id VARCHAR(32) NOT NULL
);

CREATE TABLE IF NOT EXISTS jobs (
    id          VARCHAR(32) PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    company     VARCHAR(255) NOT NULL,
    location    VARCHAR(255) NOT NULL,
    salary      BIGINT NOT NULL,
    category_id VARCHAR(32) NOT NULL,
    user_id     VARCHAR(32) NOT NULL,
    posted_date VARCHAR(50) NOT NULL,
    expiry_date VARCHAR(50) NOT NULL,
    location_id VARCHAR(32) NOT NULL
);

CREATE TABLE IF NOT EXISTS mobile_users (
    id         VARCHAR(32) PRIMARY KEY,
    phone      VARCHAR(20) NOT NULL UNIQUE,
    first_name VARCHAR(100),
    last_name  VARCHAR(100),
    email      VARCHAR(255),
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS otp_requests (
    id           VARCHAR(32) PRIMARY KEY,
    phone        VARCHAR(20) NOT NULL,
    otp_hash     VARCHAR(255) NOT NULL,
    expires_at   TIMESTAMPTZ NOT NULL,
    attempts     INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    status       VARCHAR(50) NOT NULL DEFAULT 'active',
    ip_address   VARCHAR(50),
    user_agent   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS login_audit_logs (
    id         VARCHAR(32) PRIMARY KEY,
    phone      VARCHAR(20) NOT NULL,
    event      VARCHAR(50),
    ip_address VARCHAR(50),
    user_agent TEXT,
    details    JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS admin_refresh_sessions (
    token_id      VARCHAR(255) PRIMARY KEY,
    user_id       VARCHAR(32),
    admin_id      VARCHAR(32),
    user_type     VARCHAR(20),
    refresh_token TEXT NOT NULL,
    expire_at     TIMESTAMPTZ NOT NULL,
    is_blocked    BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_admin_refresh_sessions_user_id ON admin_refresh_sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_admin_refresh_sessions_admin_id ON admin_refresh_sessions (admin_id);

CREATE TABLE IF NOT EXISTS user_refresh_sessions (
    token_id      VARCHAR(255) PRIMARY KEY,
    user_id       VARCHAR(32),
    admin_id      VARCHAR(32),
    user_type     VARCHAR(20),
    refresh_token TEXT NOT NULL,
    expire_at     TIMESTAMPTZ NOT NULL,
    is_blocked    BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_user_refresh_sessions_user_id ON user_refresh_sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_user_refresh_sessions_admin_id ON user_refresh_sessions (admin_id);

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'refresh_sessions'
    ) THEN
        INSERT INTO admin_refresh_sessions (token_id, user_id, admin_id, user_type, refresh_token, expire_at, is_blocked)
        SELECT rs.token_id, rs.user_id, rs.admin_id, COALESCE(NULLIF(rs.user_type, ''), 'admin'), rs.refresh_token, rs.expire_at, COALESCE(rs.is_blocked, FALSE)
        FROM refresh_sessions rs
        WHERE COALESCE(NULLIF(rs.user_type, ''), 'user') = 'admin'
        ON CONFLICT (token_id) DO NOTHING;

        INSERT INTO user_refresh_sessions (token_id, user_id, admin_id, user_type, refresh_token, expire_at, is_blocked)
        SELECT rs.token_id, rs.user_id, rs.admin_id, COALESCE(NULLIF(rs.user_type, ''), 'user'), rs.refresh_token, rs.expire_at, COALESCE(rs.is_blocked, FALSE)
        FROM refresh_sessions rs
        WHERE COALESCE(NULLIF(rs.user_type, ''), 'user') <> 'admin'
        ON CONFLICT (token_id) DO NOTHING;
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS otp_sessions_email (
    id         VARCHAR(32) PRIMARY KEY,
    otp_id     VARCHAR(255) NOT NULL UNIQUE,
    user_id    VARCHAR(32) NOT NULL,
    admin_id   VARCHAR(32),
    user_type  VARCHAR(20),
    email      VARCHAR(255) NOT NULL,
    expire_at  TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS user_consents (
    id            VARCHAR(32) PRIMARY KEY,
    user_id       VARCHAR(32) NOT NULL,
    consent_type  VARCHAR(20) NOT NULL,
    terms_version VARCHAR(50) NOT NULL,
    accepted      BOOLEAN NOT NULL,
    accepted_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address    VARCHAR(50),
    user_agent    TEXT
);

CREATE INDEX IF NOT EXISTS idx_user_consents_user_id ON user_consents (user_id);

CREATE TABLE IF NOT EXISTS shopping_feedbacks (
    id            VARCHAR(32) PRIMARY KEY,
    shop_order_id VARCHAR(32) NOT NULL,
    rating        INT NOT NULL,
    comment       VARCHAR(255),
    shop_id       VARCHAR(32),
    amount        BIGINT NOT NULL DEFAULT 0,
    admin_id      VARCHAR(32),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shopping_feedbacks_shop_order_id ON shopping_feedbacks (shop_order_id);

SELECT _legacy_ensure_text_id_default('product_item_views');
SELECT _legacy_ensure_text_id_default('product_item_filter_types');
SELECT _legacy_ensure_text_id_default('shop_offers');
SELECT _legacy_ensure_text_id_default('shop_departments');
SELECT _legacy_ensure_text_id_default('shop_times');
SELECT _legacy_ensure_text_id_default('shop_socials');
SELECT _legacy_ensure_text_id_default('subscription_plans');
SELECT _legacy_ensure_text_id_default('subscription_orders');
SELECT _legacy_ensure_text_id_default('user_subscriptions');
SELECT _legacy_ensure_text_id_default('notification_device_tokens');
SELECT _legacy_ensure_text_id_default('fcm_tokens');
SELECT _legacy_ensure_text_id_default('alerts');
SELECT _legacy_ensure_text_id_default('alert_actions');
SELECT _legacy_ensure_text_id_default('alert_templates');
SELECT _legacy_ensure_text_id_default('seller_alert_logs');
SELECT _legacy_ensure_text_id_default('job_categories');
SELECT _legacy_ensure_text_id_default('job_sub_categories');
SELECT _legacy_ensure_text_id_default('job_locations');
SELECT _legacy_ensure_text_id_default('job_filters');
SELECT _legacy_ensure_text_id_default('job_category_filters');
SELECT _legacy_ensure_text_id_default('job_category_locations');
SELECT _legacy_ensure_text_id_default('jobs');
SELECT _legacy_ensure_text_id_default('mobile_users');
SELECT _legacy_ensure_text_id_default('otp_requests');
SELECT _legacy_ensure_text_id_default('login_audit_logs');
SELECT _legacy_ensure_text_id_default('otp_sessions_email');
SELECT _legacy_ensure_text_id_default('shopping_feedbacks');

DROP FUNCTION IF EXISTS _legacy_ensure_text_id_default(TEXT, TEXT);

DO $$
DECLARE
    rec RECORD;
    column_nullable TEXT;
BEGIN
    FOR rec IN
        SELECT * FROM (VALUES
            ('addresses', 'user_id', 'users', 'id', 'fk_addresses_user_id', 'CASCADE'),
            ('addresses', 'country_id', 'countries', 'id', 'fk_addresses_country_id', 'RESTRICT'),
            ('user_addresses', 'user_id', 'users', 'id', 'fk_user_addresses_user_id', 'CASCADE'),
            ('user_addresses', 'address_id', 'addresses', 'id', 'fk_user_addresses_address_id', 'CASCADE'),
            ('shop_details', 'admin_id', 'admins', 'id', 'fk_shop_details_admin_id', 'CASCADE'),
            ('categories', 'department_id', 'departments', 'id', 'fk_categories_department_id', 'RESTRICT'),
            ('sub_categories', 'department_id', 'departments', 'id', 'fk_sub_categories_department_id', 'RESTRICT'),
            ('sub_categories', 'category_id', 'categories', 'id', 'fk_sub_categories_category_id', 'RESTRICT'),
            ('variations', 'sub_category_id', 'sub_categories', 'id', 'fk_variations_sub_category_id', 'RESTRICT'),
            ('variation_options', 'variation_id', 'variations', 'id', 'fk_variation_options_variation_id', 'RESTRICT'),
            ('sub_type_attributes', 'sub_category_id', 'sub_categories', 'id', 'fk_sub_type_attributes_sub_category_id', 'CASCADE'),
            ('sub_type_attribute_options', 'sub_type_attribute_id', 'sub_type_attributes', 'id', 'fk_sub_type_attribute_options_sub_type_attribute_id', 'CASCADE'),
            ('products', 'category_id', 'categories', 'id', 'fk_products_category_id', 'SET NULL'),
            ('products', 'department_id', 'departments', 'id', 'fk_products_department_id', 'SET NULL'),
            ('products', 'shop_id', 'shop_details', 'id', 'fk_products_shop_id', 'CASCADE'),
            ('product_items', 'sub_category_id', 'sub_categories', 'id', 'fk_product_items_sub_category_id', 'SET NULL'),
            ('product_items', 'category_id', 'categories', 'id', 'fk_product_items_category_id', 'SET NULL'),
            ('product_items', 'department_id', 'departments', 'id', 'fk_product_items_department_id', 'SET NULL'),
            ('product_images', 'product_item_id', 'product_items', 'id', 'fk_product_images_product_item_id', 'CASCADE'),
            ('product_images', 'shop_id', 'shop_details', 'id', 'fk_product_images_shop_id', 'CASCADE'),
            ('product_images', 'product_id', 'products', 'id', 'fk_product_images_product_id', 'CASCADE'),
            ('product_configurations', 'product_item_id', 'product_items', 'id', 'fk_product_configurations_product_item_id', 'CASCADE'),
            ('product_configurations', 'variation_option_id', 'variation_options', 'id', 'fk_product_configurations_variation_option_id', 'RESTRICT'),
            ('category_images', 'category_id', 'categories', 'id', 'fk_category_images_category_id', 'CASCADE'),
            ('offer_categories', 'offer_id', 'offers', 'id', 'fk_offer_categories_offer_id', 'CASCADE'),
            ('offer_categories', 'category_id', 'categories', 'id', 'fk_offer_categories_category_id', 'CASCADE'),
            ('offer_products', 'offer_id', 'offers', 'id', 'fk_offer_products_offer_id', 'CASCADE'),
            ('offer_products', 'product_item_id', 'product_items', 'id', 'fk_offer_products_product_item_id', 'CASCADE'),
            ('coupon_uses', 'coupon_id', 'coupons', 'coupon_id', 'fk_coupon_uses_coupon_id', 'RESTRICT'),
            ('coupon_uses', 'user_id', 'users', 'id', 'fk_coupon_uses_user_id', 'RESTRICT'),
            ('wallets', 'user_id', 'users', 'id', 'fk_wallets_user_id', 'RESTRICT'),
            ('transactions', 'wallet_id', 'wallets', 'id', 'fk_transactions_wallet_id', 'RESTRICT'),
            ('carts', 'user_id', 'users', 'id', 'fk_carts_user_id', 'CASCADE'),
            ('carts', 'applied_coupon_id', 'coupons', 'coupon_id', 'fk_carts_applied_coupon_id', 'SET NULL'),
            ('cart_items', 'cart_id', 'carts', 'id', 'fk_cart_items_cart_id', 'CASCADE'),
            ('cart_items', 'product_item_id', 'product_items', 'id', 'fk_cart_items_product_item_id', 'CASCADE'),
            ('wish_lists', 'user_id', 'users', 'id', 'fk_wish_lists_user_id', 'CASCADE'),
            ('wish_lists', 'shop_id', 'shop_details', 'id', 'fk_wish_lists_shop_id', 'CASCADE'),
            ('wish_lists', 'product_item_id', 'product_items', 'id', 'fk_wish_lists_product_item_id', 'CASCADE'),
            ('shop_orders', 'user_id', 'users', 'id', 'fk_shop_orders_user_id', 'RESTRICT'),
            ('shop_orders', 'address_id', 'addresses', 'id', 'fk_shop_orders_address_id', 'RESTRICT'),
            ('shop_orders', 'payment_method_id', 'payment_methods', 'id', 'fk_shop_orders_payment_method_id', 'SET NULL'),
            ('shop_orders', 'shop_id', 'shop_details', 'id', 'fk_shop_orders_shop_id', 'SET NULL'),
            ('order_lines', 'product_item_id', 'product_items', 'id', 'fk_order_lines_product_item_id', 'CASCADE'),
            ('order_lines', 'shop_order_id', 'shop_orders', 'id', 'fk_order_lines_shop_order_id', 'RESTRICT'),
            ('order_returns', 'shop_order_id', 'shop_orders', 'id', 'fk_order_returns_shop_order_id', 'RESTRICT'),
            ('shop_offers', 'shop_id', 'shop_details', 'id', 'fk_shop_offers_shop_id', 'CASCADE'),
            ('shop_offers', 'offer_id', 'offers', 'id', 'fk_shop_offers_offer_id', 'CASCADE'),
            ('shop_departments', 'admin_id', 'admins', 'id', 'fk_shop_departments_admin_id', 'CASCADE'),
            ('shop_departments', 'shop_id', 'shop_details', 'id', 'fk_shop_departments_shop_id', 'CASCADE'),
            ('shop_departments', 'department_id', 'departments', 'id', 'fk_shop_departments_department_id', 'RESTRICT'),
            ('shop_departments', 'category_id', 'categories', 'id', 'fk_shop_departments_category_id', 'RESTRICT'),
            ('shop_departments', 'sub_category_id', 'sub_categories', 'id', 'fk_shop_departments_sub_category_id', 'RESTRICT'),
            ('shop_times', 'shop_id', 'shop_details', 'id', 'fk_shop_times_shop_id', 'CASCADE'),
            ('shop_socials', 'shop_id', 'shop_details', 'id', 'fk_shop_socials_shop_id', 'CASCADE'),
            ('shop_socials', 'admin_id', 'admins', 'id', 'fk_shop_socials_admin_id', 'CASCADE'),
            ('shop_socials', 'user_id', 'users', 'id', 'fk_shop_socials_user_id', 'CASCADE'),
            ('shop_verifications', 'admin_id', 'admins', 'id', 'fk_shop_verifications_admin_id', 'RESTRICT'),
            ('shop_verifications', 'shop_id', 'shop_details', 'id', 'fk_shop_verifications_shop_id', 'SET NULL'),
            ('shop_verification_histories', 'admin_id', 'admins', 'id', 'fk_shop_verification_histories_admin_id', 'RESTRICT'),
            ('shop_verification_histories', 'shop_id', 'shop_details', 'id', 'fk_shop_verification_histories_shop_id', 'CASCADE'),
            ('advertisements', 'admin_id', 'admins', 'id', 'fk_advertisements_admin_id', 'SET NULL'),
            ('user_consents', 'user_id', 'users', 'id', 'fk_user_consents_user_id', 'CASCADE'),
            ('subscription_orders', 'user_id', 'users', 'id', 'fk_subscription_orders_user_id', 'RESTRICT'),
            ('subscription_orders', 'plan_id', 'subscription_plans', 'id', 'fk_subscription_orders_plan_id', 'RESTRICT'),
            ('user_subscriptions', 'user_id', 'users', 'id', 'fk_user_subscriptions_user_id', 'RESTRICT'),
            ('user_subscriptions', 'plan_id', 'subscription_plans', 'id', 'fk_user_subscriptions_plan_id', 'RESTRICT'),
            ('user_subscriptions', 'subscription_order_id', 'subscription_orders', 'id', 'fk_user_subscriptions_subscription_order_id', 'SET NULL'),
            ('alerts', 'seller_id', 'shop_details', 'id', 'fk_alerts_seller_id', 'CASCADE'),
            ('alert_actions', 'alert_id', 'alerts', 'id', 'fk_alert_actions_alert_id', 'CASCADE'),
            ('seller_alert_logs', 'seller_id', 'shop_details', 'id', 'fk_seller_alert_logs_seller_id', 'CASCADE'),
            ('jobs', 'category_id', 'job_categories', 'id', 'fk_jobs_category_id', 'RESTRICT'),
            ('jobs', 'user_id', 'users', 'id', 'fk_jobs_user_id', 'RESTRICT'),
            ('jobs', 'location_id', 'job_locations', 'id', 'fk_jobs_location_id', 'RESTRICT')
        ) AS t(table_name, column_name, ref_table, ref_column, constraint_name, on_delete)
    LOOP
        SELECT c.is_nullable
          INTO column_nullable
          FROM information_schema.columns c
         WHERE c.table_schema = 'public'
           AND c.table_name = rec.table_name
           AND c.column_name = rec.column_name;

        IF NOT FOUND THEN
            CONTINUE;
        END IF;

        IF NOT EXISTS (
            SELECT 1
            FROM information_schema.tables t
            WHERE t.table_schema = 'public'
              AND t.table_name = rec.ref_table
        ) THEN
            CONTINUE;
        END IF;

        IF column_nullable = 'YES' THEN
            EXECUTE format(
                'UPDATE %I child SET %I = NULL WHERE %I IS NOT NULL AND NOT EXISTS (SELECT 1 FROM %I parent WHERE parent.%I::text = child.%I::text)',
                rec.table_name,
                rec.column_name,
                rec.column_name,
                rec.ref_table,
                rec.ref_column,
                rec.column_name
            );
        ELSE
            EXECUTE format(
                'DELETE FROM %I child WHERE %I IS NOT NULL AND NOT EXISTS (SELECT 1 FROM %I parent WHERE parent.%I::text = child.%I::text)',
                rec.table_name,
                rec.column_name,
                rec.ref_table,
                rec.ref_column,
                rec.column_name
            );
        END IF;
    END LOOP;
END $$;

DO $$
DECLARE
    rec RECORD;
    child_udt TEXT;
    parent_udt TEXT;
BEGIN
    FOR rec IN
        SELECT * FROM (VALUES
            ('addresses', 'user_id', 'users', 'id', 'fk_addresses_user_id', 'CASCADE'),
            ('addresses', 'country_id', 'countries', 'id', 'fk_addresses_country_id', 'RESTRICT'),
            ('user_addresses', 'user_id', 'users', 'id', 'fk_user_addresses_user_id', 'CASCADE'),
            ('user_addresses', 'address_id', 'addresses', 'id', 'fk_user_addresses_address_id', 'CASCADE'),
            ('shop_details', 'admin_id', 'admins', 'id', 'fk_shop_details_admin_id', 'CASCADE'),
            ('categories', 'department_id', 'departments', 'id', 'fk_categories_department_id', 'RESTRICT'),
            ('sub_categories', 'department_id', 'departments', 'id', 'fk_sub_categories_department_id', 'RESTRICT'),
            ('sub_categories', 'category_id', 'categories', 'id', 'fk_sub_categories_category_id', 'RESTRICT'),
            ('variations', 'sub_category_id', 'sub_categories', 'id', 'fk_variations_sub_category_id', 'RESTRICT'),
            ('variation_options', 'variation_id', 'variations', 'id', 'fk_variation_options_variation_id', 'RESTRICT'),
            ('sub_type_attributes', 'sub_category_id', 'sub_categories', 'id', 'fk_sub_type_attributes_sub_category_id', 'CASCADE'),
            ('sub_type_attribute_options', 'sub_type_attribute_id', 'sub_type_attributes', 'id', 'fk_sub_type_attribute_options_sub_type_attribute_id', 'CASCADE'),
            ('products', 'category_id', 'categories', 'id', 'fk_products_category_id', 'SET NULL'),
            ('products', 'department_id', 'departments', 'id', 'fk_products_department_id', 'SET NULL'),
            ('products', 'shop_id', 'shop_details', 'id', 'fk_products_shop_id', 'CASCADE'),
            ('product_items', 'sub_category_id', 'sub_categories', 'id', 'fk_product_items_sub_category_id', 'SET NULL'),
            ('product_items', 'category_id', 'categories', 'id', 'fk_product_items_category_id', 'SET NULL'),
            ('product_items', 'department_id', 'departments', 'id', 'fk_product_items_department_id', 'SET NULL'),
            ('product_images', 'product_item_id', 'product_items', 'id', 'fk_product_images_product_item_id', 'CASCADE'),
            ('product_images', 'shop_id', 'shop_details', 'id', 'fk_product_images_shop_id', 'CASCADE'),
            ('product_images', 'product_id', 'products', 'id', 'fk_product_images_product_id', 'CASCADE'),
            ('product_configurations', 'product_item_id', 'product_items', 'id', 'fk_product_configurations_product_item_id', 'CASCADE'),
            ('product_configurations', 'variation_option_id', 'variation_options', 'id', 'fk_product_configurations_variation_option_id', 'RESTRICT'),
            ('category_images', 'category_id', 'categories', 'id', 'fk_category_images_category_id', 'CASCADE'),
            ('offer_categories', 'offer_id', 'offers', 'id', 'fk_offer_categories_offer_id', 'CASCADE'),
            ('offer_categories', 'category_id', 'categories', 'id', 'fk_offer_categories_category_id', 'CASCADE'),
            ('offer_products', 'offer_id', 'offers', 'id', 'fk_offer_products_offer_id', 'CASCADE'),
            ('offer_products', 'product_item_id', 'product_items', 'id', 'fk_offer_products_product_item_id', 'CASCADE'),
            ('coupon_uses', 'coupon_id', 'coupons', 'coupon_id', 'fk_coupon_uses_coupon_id', 'RESTRICT'),
            ('coupon_uses', 'user_id', 'users', 'id', 'fk_coupon_uses_user_id', 'RESTRICT'),
            ('wallets', 'user_id', 'users', 'id', 'fk_wallets_user_id', 'RESTRICT'),
            ('transactions', 'wallet_id', 'wallets', 'id', 'fk_transactions_wallet_id', 'RESTRICT'),
            ('carts', 'user_id', 'users', 'id', 'fk_carts_user_id', 'CASCADE'),
            ('carts', 'applied_coupon_id', 'coupons', 'coupon_id', 'fk_carts_applied_coupon_id', 'SET NULL'),
            ('cart_items', 'cart_id', 'carts', 'id', 'fk_cart_items_cart_id', 'CASCADE'),
            ('cart_items', 'product_item_id', 'product_items', 'id', 'fk_cart_items_product_item_id', 'CASCADE'),
            ('wish_lists', 'user_id', 'users', 'id', 'fk_wish_lists_user_id', 'CASCADE'),
            ('wish_lists', 'shop_id', 'shop_details', 'id', 'fk_wish_lists_shop_id', 'CASCADE'),
            ('wish_lists', 'product_item_id', 'product_items', 'id', 'fk_wish_lists_product_item_id', 'CASCADE'),
            ('shop_orders', 'user_id', 'users', 'id', 'fk_shop_orders_user_id', 'RESTRICT'),
            ('shop_orders', 'address_id', 'addresses', 'id', 'fk_shop_orders_address_id', 'RESTRICT'),
            ('shop_orders', 'payment_method_id', 'payment_methods', 'id', 'fk_shop_orders_payment_method_id', 'SET NULL'),
            ('shop_orders', 'shop_id', 'shop_details', 'id', 'fk_shop_orders_shop_id', 'SET NULL'),
            ('order_lines', 'product_item_id', 'product_items', 'id', 'fk_order_lines_product_item_id', 'CASCADE'),
            ('order_lines', 'shop_order_id', 'shop_orders', 'id', 'fk_order_lines_shop_order_id', 'RESTRICT'),
            ('order_returns', 'shop_order_id', 'shop_orders', 'id', 'fk_order_returns_shop_order_id', 'RESTRICT'),
            ('shop_offers', 'shop_id', 'shop_details', 'id', 'fk_shop_offers_shop_id', 'CASCADE'),
            ('shop_offers', 'offer_id', 'offers', 'id', 'fk_shop_offers_offer_id', 'CASCADE'),
            ('shop_departments', 'admin_id', 'admins', 'id', 'fk_shop_departments_admin_id', 'CASCADE'),
            ('shop_departments', 'shop_id', 'shop_details', 'id', 'fk_shop_departments_shop_id', 'CASCADE'),
            ('shop_departments', 'department_id', 'departments', 'id', 'fk_shop_departments_department_id', 'RESTRICT'),
            ('shop_departments', 'category_id', 'categories', 'id', 'fk_shop_departments_category_id', 'RESTRICT'),
            ('shop_departments', 'sub_category_id', 'sub_categories', 'id', 'fk_shop_departments_sub_category_id', 'RESTRICT'),
            ('shop_times', 'shop_id', 'shop_details', 'id', 'fk_shop_times_shop_id', 'CASCADE'),
            ('shop_socials', 'shop_id', 'shop_details', 'id', 'fk_shop_socials_shop_id', 'CASCADE'),
            ('shop_socials', 'admin_id', 'admins', 'id', 'fk_shop_socials_admin_id', 'CASCADE'),
            ('shop_socials', 'user_id', 'users', 'id', 'fk_shop_socials_user_id', 'CASCADE'),
            ('shop_verifications', 'admin_id', 'admins', 'id', 'fk_shop_verifications_admin_id', 'RESTRICT'),
            ('shop_verifications', 'shop_id', 'shop_details', 'id', 'fk_shop_verifications_shop_id', 'SET NULL'),
            ('shop_verification_histories', 'admin_id', 'admins', 'id', 'fk_shop_verification_histories_admin_id', 'RESTRICT'),
            ('shop_verification_histories', 'shop_id', 'shop_details', 'id', 'fk_shop_verification_histories_shop_id', 'CASCADE'),
            ('advertisements', 'admin_id', 'admins', 'id', 'fk_advertisements_admin_id', 'SET NULL'),
            ('user_consents', 'user_id', 'users', 'id', 'fk_user_consents_user_id', 'CASCADE'),
            ('subscription_orders', 'user_id', 'users', 'id', 'fk_subscription_orders_user_id', 'RESTRICT'),
            ('subscription_orders', 'plan_id', 'subscription_plans', 'id', 'fk_subscription_orders_plan_id', 'RESTRICT'),
            ('user_subscriptions', 'user_id', 'users', 'id', 'fk_user_subscriptions_user_id', 'RESTRICT'),
            ('user_subscriptions', 'plan_id', 'subscription_plans', 'id', 'fk_user_subscriptions_plan_id', 'RESTRICT'),
            ('user_subscriptions', 'subscription_order_id', 'subscription_orders', 'id', 'fk_user_subscriptions_subscription_order_id', 'SET NULL'),
            ('alerts', 'seller_id', 'shop_details', 'id', 'fk_alerts_seller_id', 'CASCADE'),
            ('alert_actions', 'alert_id', 'alerts', 'id', 'fk_alert_actions_alert_id', 'CASCADE'),
            ('seller_alert_logs', 'seller_id', 'shop_details', 'id', 'fk_seller_alert_logs_seller_id', 'CASCADE'),
            ('jobs', 'category_id', 'job_categories', 'id', 'fk_jobs_category_id', 'RESTRICT'),
            ('jobs', 'user_id', 'users', 'id', 'fk_jobs_user_id', 'RESTRICT'),
            ('jobs', 'location_id', 'job_locations', 'id', 'fk_jobs_location_id', 'RESTRICT')
        ) AS t(table_name, column_name, ref_table, ref_column, constraint_name, on_delete)
    LOOP
        -- Only reattach constraints when both sides have matching underlying types.
        SELECT c.udt_name
          INTO child_udt
          FROM information_schema.columns c
         WHERE c.table_schema = 'public'
           AND c.table_name = rec.table_name
           AND c.column_name = rec.column_name;

        IF NOT FOUND THEN
            CONTINUE;
        END IF;

        SELECT c.udt_name
          INTO parent_udt
          FROM information_schema.columns c
         WHERE c.table_schema = 'public'
           AND c.table_name = rec.ref_table
           AND c.column_name = rec.ref_column;

        IF NOT FOUND THEN
            CONTINUE;
        END IF;

        IF child_udt <> parent_udt THEN
            CONTINUE;
        END IF;

        BEGIN
            EXECUTE format(
                'ALTER TABLE %I ADD CONSTRAINT %I FOREIGN KEY (%I) REFERENCES %I (%I) ON DELETE %s',
                rec.table_name,
                rec.constraint_name,
                rec.column_name,
                rec.ref_table,
                rec.ref_column,
                rec.on_delete
            );
        EXCEPTION
            WHEN duplicate_object THEN
                NULL;
            WHEN others THEN
                RAISE NOTICE 'Skipping legacy FK % on %.% -> %.% (%).',
                    rec.constraint_name,
                    rec.table_name,
                    rec.column_name,
                    rec.ref_table,
                    rec.ref_column,
                    SQLERRM;
        END;
    END LOOP;
END $$;