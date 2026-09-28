-- Durable, per-follower record of the "new product" push. Replaces the
-- per-shop-only shop_new_product_notifications guard (000044): that table
-- could only say "has ANY follower of this shop been notified today", so a
-- follower who followed AFTER an earlier attempt that day (even one that
-- delivered to zero devices) was silently skipped for the rest of the day
-- with no way to tell why. This table tracks the digest per (shop, day,
-- follower), so each follower who newly qualifies gets their own attempt,
-- and a failed send can be retried by the sweep ticker instead of being a
-- one-shot, unobservable goroutine call.
CREATE TABLE IF NOT EXISTS follower_notification_deliveries (
    id           VARCHAR(32)  PRIMARY KEY,
    shop_id      VARCHAR(32)  NOT NULL,
    notify_date  DATE         NOT NULL,
    follower_id  VARCHAR(32)  NOT NULL,
    title        TEXT         NOT NULL DEFAULT '',
    body         TEXT         NOT NULL DEFAULT '',
    image_url    TEXT         NOT NULL DEFAULT '',
    status       VARCHAR(20)  NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'sent', 'dead')),
    attempts     INT          NOT NULL DEFAULT 0,
    last_error   TEXT         NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (shop_id, notify_date, follower_id)
);

-- The retry sweep polls exactly this shape: oldest pending rows first.
CREATE INDEX IF NOT EXISTS idx_follower_notif_deliveries_pending
    ON follower_notification_deliveries (updated_at)
    WHERE status = 'pending';
