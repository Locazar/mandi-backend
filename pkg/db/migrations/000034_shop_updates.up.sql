-- Shop Updates: a curated per-shop "What's New" advertisement. Each row
-- features a registered shop (shop_details) plus a few of that shop's catalog
-- products the seller wants customers to see. Purely additive.

CREATE TABLE IF NOT EXISTS shop_updates (
    id           VARCHAR(32) PRIMARY KEY,
    shop_id      VARCHAR(32) NOT NULL REFERENCES shop_details (id) ON DELETE CASCADE,
    -- Optional per-entry shop image override (bare CDN key). When empty the
    -- customer read falls back to the shop's own logo (shop_details.shop_image_url).
    image_url    TEXT,
    action_label VARCHAR(50) NOT NULL DEFAULT 'Visit',
    is_active    BOOLEAN     NOT NULL DEFAULT TRUE,
    sort_order   INT         NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_updates_shop_id ON shop_updates (shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_updates_active_order ON shop_updates (is_active, sort_order);

CREATE TABLE IF NOT EXISTS shop_update_products (
    id              VARCHAR(32) PRIMARY KEY,
    shop_update_id  VARCHAR(32) NOT NULL REFERENCES shop_updates (id) ON DELETE CASCADE,
    -- Links the real catalog product; nullable so a fully custom card is still
    -- possible. title/image_url are optional overrides — the read derives them
    -- from product_items when blank.
    product_item_id VARCHAR(32) REFERENCES product_items (id) ON DELETE CASCADE,
    title           VARCHAR(150),
    image_url       TEXT,
    attribute       VARCHAR(100),
    sort_order      INT         NOT NULL DEFAULT 0,
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_update_products_update_id ON shop_update_products (shop_update_id);
