-- Tracks the one-time "new shop nearby" broadcast per shop. When a shop first
-- goes live (shop_status -> active), customers within a short radius are told
-- once and only once: the shop_id primary key lets the app atomically claim
-- the slot with INSERT ... ON CONFLICT DO NOTHING, so neither a seller
-- re-saving onboarding details (CreateShop upserts on admin_id) nor an
-- approve -> suspend -> approve cycle can re-broadcast to the same customers.
CREATE TABLE IF NOT EXISTS shop_launch_notifications (
    shop_id    VARCHAR(32) NOT NULL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- The nearby-customer lookup filters addresses on a non-null coordinate pair
-- before doing haversine math; this keeps that pre-filter cheap.
CREATE INDEX IF NOT EXISTS idx_addresses_lat_lng
    ON addresses (latitude, longitude)
    WHERE latitude IS NOT NULL AND longitude IS NOT NULL;
