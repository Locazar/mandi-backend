-- Shop social table updates for follow/like/rating/review features
CREATE TABLE IF NOT EXISTS shop_socials (
    id BIGSERIAL PRIMARY KEY,
    shop_id BIGINT NOT NULL,
    admin_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    is_follower BOOLEAN NOT NULL DEFAULT FALSE,
    is_liked BOOLEAN NOT NULL DEFAULT FALSE,
    rating INTEGER NOT NULL DEFAULT 0,
    review TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE shop_socials
    ADD COLUMN IF NOT EXISTS is_liked BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_shop_socials_shop_id ON shop_socials(shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_socials_user_id ON shop_socials(user_id);
CREATE INDEX IF NOT EXISTS idx_shop_socials_admin_id ON shop_socials(admin_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_shop_socials_shop_user_unique ON shop_socials(shop_id, user_id);
