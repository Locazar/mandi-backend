-- Tracks the once-per-day "new product" follower notification per shop. When a
-- seller adds a product, followers are notified at most once per calendar day —
-- the (shop_id, notify_date) primary key lets the app atomically claim a day's
-- slot with INSERT ... ON CONFLICT DO NOTHING, so rapid repeat adds send one push.
CREATE TABLE IF NOT EXISTS shop_new_product_notifications (
    shop_id     VARCHAR(32) NOT NULL,
    notify_date DATE        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (shop_id, notify_date)
);
