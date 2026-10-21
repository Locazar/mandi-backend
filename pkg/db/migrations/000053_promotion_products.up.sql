-- One promotion per product item (seller-applied).
-- Distinct from offer_products (admin legacy system); this table links a
-- seller's own promotion to one of their product items.
CREATE TABLE IF NOT EXISTS promotion_products (
    id              VARCHAR(32)  PRIMARY KEY,
    promotion_id    VARCHAR(32)  NOT NULL REFERENCES promotions(id) ON DELETE CASCADE,
    product_item_id VARCHAR(32)  NOT NULL REFERENCES product_items(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (product_item_id)  -- at most one active promotion per product
);

CREATE INDEX IF NOT EXISTS idx_promotion_products_promotion_id    ON promotion_products(promotion_id);
CREATE INDEX IF NOT EXISTS idx_promotion_products_product_item_id ON promotion_products(product_item_id);
