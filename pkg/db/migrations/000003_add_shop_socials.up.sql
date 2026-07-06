CREATE TABLE IF NOT EXISTS shop_socials (
    id          VARCHAR(32)  PRIMARY KEY,
    shop_id     VARCHAR(32)  NOT NULL,
    admin_id    VARCHAR(32)  NOT NULL DEFAULT '',
    user_id     VARCHAR(32)  NOT NULL DEFAULT '',
    is_follower BOOLEAN      NOT NULL DEFAULT FALSE,
    is_liked    BOOLEAN      NOT NULL DEFAULT FALSE,
    rating      INTEGER      NOT NULL DEFAULT 0,
    review      TEXT         NOT NULL DEFAULT '',
    comments    TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_socials_shop_id  ON shop_socials (shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_socials_admin_id ON shop_socials (admin_id);
CREATE INDEX IF NOT EXISTS idx_shop_socials_user_id  ON shop_socials (user_id);
