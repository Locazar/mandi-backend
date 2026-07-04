CREATE TABLE IF NOT EXISTS advertisement_requests (
    id            VARCHAR(32) PRIMARY KEY,
    admin_id      VARCHAR(32) NOT NULL REFERENCES admins (id) ON DELETE CASCADE,
    shop_id       VARCHAR(32) REFERENCES shop_details (id) ON DELETE SET NULL,
    title         VARCHAR(100),
    content       TEXT,
    start_date    TIMESTAMPTZ NOT NULL,
    end_date      TIMESTAMPTZ NOT NULL,
    plan_key      VARCHAR(20) NOT NULL,
    price_minor   BIGINT      NOT NULL DEFAULT 0,
    status        VARCHAR(20) NOT NULL DEFAULT 'pending',
    admin_comment TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_advertisement_requests_admin_id ON advertisement_requests (admin_id);
CREATE INDEX IF NOT EXISTS idx_advertisement_requests_status   ON advertisement_requests (status);
