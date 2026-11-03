CREATE TABLE IF NOT EXISTS notify_customer_requests (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enquiry_id   VARCHAR(100) NOT NULL,
    product_id   VARCHAR(100) NOT NULL,
    product_name VARCHAR(255),
    image_url    TEXT,
    deal_amount  DECIMAL(12,2) NOT NULL DEFAULT 0,
    customer_id  VARCHAR(100) NOT NULL,
    customer_name VARCHAR(255),
    seller_id    VARCHAR(100) NOT NULL,
    shop_id      VARCHAR(100),
    status       VARCHAR(50) NOT NULL,
    admin_note   TEXT,
    is_actioned  BOOLEAN NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT notify_customer_requests_enquiry_id_key UNIQUE (enquiry_id)
);
