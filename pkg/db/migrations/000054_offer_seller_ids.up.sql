-- Add seller_id to shop_offers so we know which seller applied the offer.
ALTER TABLE shop_offers ADD COLUMN IF NOT EXISTS seller_id VARCHAR(32);

-- Add shop_id and seller_id to promotion_products for reporting / filtering.
ALTER TABLE promotion_products ADD COLUMN IF NOT EXISTS shop_id   VARCHAR(32);
ALTER TABLE promotion_products ADD COLUMN IF NOT EXISTS seller_id VARCHAR(32);

CREATE INDEX IF NOT EXISTS idx_shop_offers_seller_id          ON shop_offers(seller_id);
CREATE INDEX IF NOT EXISTS idx_promotion_products_shop_id     ON promotion_products(shop_id);
CREATE INDEX IF NOT EXISTS idx_promotion_products_seller_id   ON promotion_products(seller_id);
