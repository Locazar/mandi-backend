CREATE TABLE IF NOT EXISTS store_discounts (
    id          VARCHAR(32)    PRIMARY KEY,
    shop_id     VARCHAR(32)    NOT NULL UNIQUE REFERENCES shop_details(id) ON DELETE CASCADE,
    kind        VARCHAR(20)    NOT NULL CHECK (kind IN ('flat','upto','clearance')),
    value_unit  VARCHAR(20)    NOT NULL CHECK (value_unit IN ('percentage','inr')),
    value       DECIMAL(10,2)  NOT NULL DEFAULT 0 CHECK (value >= 0),
    is_active   BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_store_discounts_shop_id ON store_discounts(shop_id);
