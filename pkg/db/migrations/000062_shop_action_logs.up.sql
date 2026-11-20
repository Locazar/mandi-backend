CREATE TABLE IF NOT EXISTS shop_action_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action          VARCHAR(50) NOT NULL,
    shop_id         VARCHAR(100) NOT NULL,
    shop_name       VARCHAR(255),
    seller_id       VARCHAR(100),
    customer_id     VARCHAR(100),
    customer_phone  VARCHAR(20),
    customer_lat    DECIMAL(10,7),
    customer_lon    DECIMAL(10,7),
    location_label  VARCHAR(255),
    session_id      VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_action_logs_shop_id     ON shop_action_logs (shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_action_logs_seller_id   ON shop_action_logs (seller_id);
CREATE INDEX IF NOT EXISTS idx_shop_action_logs_customer_id ON shop_action_logs (customer_id);
CREATE INDEX IF NOT EXISTS idx_shop_action_logs_action      ON shop_action_logs (action);
CREATE INDEX IF NOT EXISTS idx_shop_action_logs_created_at  ON shop_action_logs (created_at DESC);
