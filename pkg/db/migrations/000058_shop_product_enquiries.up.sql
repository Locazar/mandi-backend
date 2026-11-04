CREATE TABLE IF NOT EXISTS shop_product_enquiries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id         VARCHAR(100) NOT NULL,
    shop_name       VARCHAR(255),
    enquiry_type    VARCHAR(50) NOT NULL,
    message         TEXT NOT NULL,
    customer_phone  VARCHAR(20),
    customer_city   VARCHAR(100),
    is_actioned     BOOLEAN NOT NULL DEFAULT false,
    admin_note      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_product_enquiries_shop_id ON shop_product_enquiries (shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_product_enquiries_created_at ON shop_product_enquiries (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shop_product_enquiries_is_actioned ON shop_product_enquiries (is_actioned);
