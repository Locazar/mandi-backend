DROP INDEX IF EXISTS idx_promotion_products_seller_id;
DROP INDEX IF EXISTS idx_promotion_products_shop_id;
DROP INDEX IF EXISTS idx_shop_offers_seller_id;

ALTER TABLE promotion_products DROP COLUMN IF EXISTS seller_id;
ALTER TABLE promotion_products DROP COLUMN IF EXISTS shop_id;
ALTER TABLE shop_offers DROP COLUMN IF EXISTS seller_id;
